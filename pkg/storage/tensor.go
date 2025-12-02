package storage

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/google/uuid"
)

// tensorImpl implements the Tensor interface
type tensorImpl struct {
	name   string
	schema TensorSchema
	engine Engine
	data   []float32
}

// Name returns the tensor name
func (t *tensorImpl) Name() string {
	return t.name
}

// Schema returns the tensor schema
func (t *tensorImpl) Schema() TensorSchema {
	return t.schema
}

// Shape returns the tensor shape
func (t *tensorImpl) Shape() []int {
	return t.schema.Shape
}

// DType returns the tensor data type
func (t *tensorImpl) DType() string {
	return t.schema.DType
}

// StoreChunk stores a chunk of data at the specified indices
func (t *tensorImpl) StoreChunk(ctx context.Context, indices []int, data []byte) error {
	// Validate indices
	if len(indices) != len(t.schema.Shape) {
		return fmt.Errorf("indices length %d doesn't match tensor dimensions %d", len(indices), len(t.schema.Shape))
	}

	for i, idx := range indices {
		if idx < 0 {
			return fmt.Errorf("negative index %d at dimension %d", idx, i)
		}
		if i < len(t.schema.ChunkSize) && idx >= (t.schema.Shape[i]+t.schema.ChunkSize[i]-1)/t.schema.ChunkSize[i] {
			return fmt.Errorf("chunk index %d exceeds dimension %d size", idx, i)
		}
	}

	// Validate and convert data
	if len(data) == 0 {
		return fmt.Errorf("empty data provided")
	}

	floatData := bytesToFloat32Slice(data)
	if floatData == nil {
		return fmt.Errorf("invalid data format: byte length must be multiple of 4")
	}

	// Calculate chunk size from schema
	chunkSize := t.calculateChunkSize()
	if len(floatData) != chunkSize {
		return fmt.Errorf("data size %d doesn't match expected chunk size %d", len(floatData), chunkSize)
	}

	// Calculate starting flat index for the chunk
	startFlatIndex := t.calculateChunkStartIndex(indices)

	// Validate bounds
	if startFlatIndex < 0 || startFlatIndex+chunkSize > len(t.data) {
		return fmt.Errorf("chunk indices out of bounds: start=%d, size=%d, tensor_size=%d",
			startFlatIndex, chunkSize, len(t.data))
	}

	// Store chunk data
	for i, value := range floatData {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return fmt.Errorf("invalid value at position %d: NaN or Inf", i)
		}
		t.data[startFlatIndex+i] = value
	}

	// Save to disk
	if err := t.save(); err != nil {
		return fmt.Errorf("failed to save tensor: %w", err)
	}

	return nil
}

// GetChunk retrieves a chunk of data at the specified indices
func (t *tensorImpl) GetChunk(ctx context.Context, indices []int) ([]byte, error) {
	// Validate indices
	if len(indices) != len(t.schema.Shape) {
		return nil, fmt.Errorf("indices length %d doesn't match tensor dimensions %d", len(indices), len(t.schema.Shape))
	}

	for i, idx := range indices {
		if idx < 0 {
			return nil, fmt.Errorf("negative index %d at dimension %d", idx, i)
		}
		if i < len(t.schema.ChunkSize) && idx >= (t.schema.Shape[i]+t.schema.ChunkSize[i]-1)/t.schema.ChunkSize[i] {
			return nil, fmt.Errorf("chunk index %d exceeds dimension %d size", idx, i)
		}
	}

	// Calculate chunk size
	chunkSize := t.calculateChunkSize()
	if chunkSize <= 0 {
		return nil, fmt.Errorf("invalid chunk size: %d", chunkSize)
	}

	// Calculate starting flat index for the chunk
	startFlatIndex := t.calculateChunkStartIndex(indices)

	// Check bounds
	if startFlatIndex < 0 || startFlatIndex+chunkSize > len(t.data) {
		return nil, fmt.Errorf("chunk indices out of bounds: start=%d, size=%d, tensor_size=%d",
			startFlatIndex, chunkSize, len(t.data))
	}

	// Extract chunk data
	chunk := t.data[startFlatIndex : startFlatIndex+chunkSize]
	return float32SliceToBytes(chunk), nil
}

// Slice returns a slice of the tensor
func (t *tensorImpl) Slice(ctx context.Context, ranges []Range) (Tensor, error) {
	// Validate ranges length
	if len(ranges) != len(t.schema.Shape) {
		return nil, fmt.Errorf("ranges length %d doesn't match tensor dimensions %d", len(ranges), len(t.schema.Shape))
	}

	// Validate each range
	for i, r := range ranges {
		if r.Start < 0 || r.End > t.schema.Shape[i] || r.Start >= r.End {
			return nil, fmt.Errorf("invalid range for dimension %d: [%d, %d), tensor size=%d",
				i, r.Start, r.End, t.schema.Shape[i])
		}
		if r.End-r.Start <= 0 {
			return nil, fmt.Errorf("empty slice for dimension %d: size=%d", i, r.End-r.Start)
		}
	}

	// Calculate new shape
	newShape := make([]int, len(t.schema.Shape))
	totalSize := 1
	for i, r := range ranges {
		newShape[i] = r.End - r.Start
		totalSize *= newShape[i]
	}

	// Check for reasonable size limits
	if totalSize > 1000000 { // 1M elements limit
		return nil, fmt.Errorf("slice too large: %d elements exceeds limit", totalSize)
	}

	// Create new tensor
	newSchema := TensorSchema{
		Shape:       newShape,
		DType:       t.schema.DType,
		ChunkSize:   t.schema.ChunkSize,
		Compression: t.schema.Compression,
		Metadata:    t.schema.Metadata,
	}

	newTensor := &tensorImpl{
		name:   fmt.Sprintf("%s_slice_%s", t.name, uuid.New().String()[:8]),
		schema: newSchema,
		engine: t.engine,
		data:   make([]float32, totalSize),
	}

	// Copy slice data using proper multi-dimensional indexing
	for destIdx := range newTensor.data {
		// Convert flat destination index to multi-dimensional indices in new tensor
		destIndices := t.flatToMultiDimIndex(destIdx, newShape)

		// Convert to source indices by adding range offsets
		srcIndices := make([]int, len(destIndices))
		for i := range destIndices {
			srcIndices[i] = destIndices[i] + ranges[i].Start
		}

		// Convert source indices to flat index in original tensor
		srcFlatIdx := t.calculateFlatIndex(srcIndices)

		// Validate source index
		if srcFlatIdx < 0 || srcFlatIdx >= len(t.data) {
			return nil, fmt.Errorf("source index out of bounds: %d", srcFlatIdx)
		}

		// Copy data
		newTensor.data[destIdx] = t.data[srcFlatIdx]
	}

	return newTensor, nil
}

// Reshape changes the tensor shape
func (t *tensorImpl) Reshape(ctx context.Context, newShape []int) error {
	// Calculate total size
	oldSize := 1
	for _, dim := range t.schema.Shape {
		oldSize *= dim
	}

	newSize := 1
	for _, dim := range newShape {
		newSize *= dim
	}

	if oldSize != newSize {
		return fmt.Errorf("cannot reshape: size mismatch (old=%d, new=%d)", oldSize, newSize)
	}

	// Update shape
	t.schema.Shape = newShape
	return t.save()
}

// ApplyOperation applies a mathematical operation to the tensor
func (t *tensorImpl) ApplyOperation(ctx context.Context, op Operation) (Tensor, error) {
	switch op.Type {
	case "add":
		return t.applyAddOperation(op)
	case "multiply":
		return t.applyMultiplyOperation(op)
	case "matrix_multiply":
		return t.applyMatrixMultiplyOperation(op)
	case "transpose":
		return t.applyTransposeOperation(op)
	case "sum":
		return t.applyReductionOperation(op, "sum")
	case "mean":
		return t.applyReductionOperation(op, "mean")
	case "max":
		return t.applyReductionOperation(op, "max")
	case "min":
		return t.applyReductionOperation(op, "min")
	case "conv1d":
		return t.applyConv1DOperation(op)
	case "conv2d":
		return t.applyConv2DOperation(op)
	case "relu":
		return t.applyActivationFunction(op, "relu")
	case "sigmoid":
		return t.applyActivationFunction(op, "sigmoid")
	case "tanh":
		return t.applyActivationFunction(op, "tanh")
	case "svd":
		return t.applySVDOperation(op)
	case "eigenvalues":
		return t.applyEigenvaluesOperation(op)
	case "cosine_similarity":
		return t.applyCosineSimilarity(op)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", op.Type)
	}
}

// Metadata returns the tensor metadata
func (t *tensorImpl) Metadata() map[string]interface{} {
	return t.schema.Metadata
}

// SetMetadata sets a metadata value
func (t *tensorImpl) SetMetadata(key string, value interface{}) error {
	if t.schema.Metadata == nil {
		t.schema.Metadata = make(map[string]interface{})
	}
	t.schema.Metadata[key] = value
	return t.save()
}

// Helper methods

func (t *tensorImpl) calculateFlatIndex(indices []int) int {
	if len(indices) != len(t.schema.Shape) {
		return 0
	}

	index := 0
	stride := 1
	for i := len(indices) - 1; i >= 0; i-- {
		index += indices[i] * stride
		stride *= t.schema.Shape[i]
	}

	return index
}

func (t *tensorImpl) flatToMultiDimIndex(flatIndex int, shape []int) []int {
	indices := make([]int, len(shape))

	for i := len(indices) - 1; i >= 0; i-- {
		indices[i] = flatIndex % shape[i]
		flatIndex /= shape[i]
	}

	return indices
}

func (t *tensorImpl) calculateChunkSize() int {
	if len(t.schema.ChunkSize) == 0 {
		return 1 // Default chunk size
	}

	size := 1
	for _, dim := range t.schema.ChunkSize {
		size *= dim
	}
	return size
}

func (t *tensorImpl) calculateChunkStartIndex(indices []int) int {
	// Convert chunk indices to flat index considering chunk size
	chunkIndices := make([]int, len(indices))
	for i := range indices {
		chunkStride := t.schema.ChunkSize[i]
		if chunkStride == 0 {
			chunkStride = 1
		}
		chunkIndices[i] = indices[i] * chunkStride
	}

	return t.calculateFlatIndex(chunkIndices)
}

// broadcastShapes determines the broadcast shape for two tensors
func broadcastShapes(shape1, shape2 []int) ([]int, error) {
	// Pad the shorter shape with leading 1s
	maxLen := max(len(shape1), len(shape2))
	paddedShape1 := make([]int, maxLen)
	paddedShape2 := make([]int, maxLen)

	for i := 0; i < maxLen; i++ {
		idx1 := len(shape1) - maxLen + i
		idx2 := len(shape2) - maxLen + i

		if idx1 >= 0 {
			paddedShape1[i] = shape1[idx1]
		} else {
			paddedShape1[i] = 1
		}

		if idx2 >= 0 {
			paddedShape2[i] = shape2[idx2]
		} else {
			paddedShape2[i] = 1
		}
	}

	// Calculate broadcast shape
	broadcastShape := make([]int, maxLen)
	for i := 0; i < maxLen; i++ {
		if paddedShape1[i] != paddedShape2[i] && paddedShape1[i] != 1 && paddedShape2[i] != 1 {
			return nil, fmt.Errorf("shapes %v and %v are not broadcastable", shape1, shape2)
		}
		broadcastShape[i] = max(paddedShape1[i], paddedShape2[i])
	}

	return broadcastShape, nil
}

// broadcastTensor expands a tensor to the target broadcast shape
func (t *tensorImpl) broadcastTensor(targetShape []int) (*tensorImpl, error) {
	broadcastShape, err := broadcastShapes(t.schema.Shape, targetShape)
	if err != nil {
		return nil, err
	}

	// Create broadcasted tensor
	broadcastSchema := TensorSchema{
		Shape:       broadcastShape,
		DType:       t.schema.DType,
		ChunkSize:   t.schema.ChunkSize,
		Compression: t.schema.Compression,
		Metadata:    t.schema.Metadata,
	}

	broadcasted := &tensorImpl{
		name:   fmt.Sprintf("%s_broadcast", t.name),
		schema: broadcastSchema,
		engine: t.engine,
		data:   make([]float32, t.calculateSize(broadcastShape)),
	}

	// Fill broadcasted data
	for i := range broadcasted.data {
		// Convert flat index to multi-dimensional indices in broadcasted tensor
		indices := t.flatToMultiDimIndex(i, broadcastShape)

		// Map to original tensor indices (handle broadcasting)
		originalIndices := make([]int, len(t.schema.Shape))
		for j := range originalIndices {
			broadcastIdx := len(broadcastShape) - len(t.schema.Shape) + j
			if broadcastIdx >= 0 {
				if t.schema.Shape[j] == 1 {
					originalIndices[j] = 0 // Broadcast dimension
				} else {
					originalIndices[j] = indices[broadcastIdx]
				}
			}
		}

		// Get value from original tensor
		originalFlatIdx := t.calculateFlatIndex(originalIndices)
		broadcasted.data[i] = t.data[originalFlatIdx]
	}

	return broadcasted, nil
}

func (t *tensorImpl) calculateSize(shape []int) int {
	size := 1
	for _, dim := range shape {
		size *= dim
	}
	return size
}

func (t *tensorImpl) getFilePath() string {
	return filepath.Join(t.engine.(*engineImpl).dataDir, "tensor_"+t.name+".bin")
}

func (t *tensorImpl) save() error {
	filePath := t.getFilePath()

	// Convert float32 slice to bytes
	data := float32SliceToBytes(t.data)

	return os.WriteFile(filePath, data, 0644)
}

func (t *tensorImpl) load() error {
	filePath := t.getFilePath()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, initialize with zeros
			return nil
		}
		return err
	}

	// Convert bytes to float32 slice
	t.data = bytesToFloat32Slice(data)
	return nil
}

// Operation implementations

func (t *tensorImpl) applyAddOperation(op Operation) (Tensor, error) {
	otherTensor, ok := op.Operand.(*tensorImpl)
	if !ok {
		return nil, fmt.Errorf("operand must be a tensor")
	}

	// Calculate broadcast shape
	broadcastShape, err := broadcastShapes(t.schema.Shape, otherTensor.schema.Shape)
	if err != nil {
		return nil, fmt.Errorf("cannot broadcast shapes: %w", err)
	}

	// Broadcast both tensors to the same shape
	broadcastedT, err := t.broadcastTensor(broadcastShape)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast first tensor: %w", err)
	}

	broadcastedOther, err := otherTensor.broadcastTensor(broadcastShape)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast second tensor: %w", err)
	}

	// Create result tensor
	resultSchema := TensorSchema{
		Shape:       broadcastShape,
		DType:       t.schema.DType,
		ChunkSize:   t.schema.ChunkSize,
		Compression: t.schema.Compression,
		Metadata:    map[string]interface{}{"operation": "add"},
	}

	result := &tensorImpl{
		name:   fmt.Sprintf("%s_plus_%s", t.name, otherTensor.name),
		schema: resultSchema,
		engine: t.engine,
		data:   make([]float32, len(broadcastedT.data)),
	}

	// Perform element-wise addition
	for i := range broadcastedT.data {
		result.data[i] = broadcastedT.data[i] + broadcastedOther.data[i]
	}

	return result, nil
}

func (t *tensorImpl) applyMultiplyOperation(op Operation) (Tensor, error) {
	otherTensor, ok := op.Operand.(*tensorImpl)
	if !ok {
		return nil, fmt.Errorf("operand must be a tensor")
	}

	// Calculate broadcast shape
	broadcastShape, err := broadcastShapes(t.schema.Shape, otherTensor.schema.Shape)
	if err != nil {
		return nil, fmt.Errorf("cannot broadcast shapes: %w", err)
	}

	// Broadcast both tensors to the same shape
	broadcastedT, err := t.broadcastTensor(broadcastShape)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast first tensor: %w", err)
	}

	broadcastedOther, err := otherTensor.broadcastTensor(broadcastShape)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast second tensor: %w", err)
	}

	// Create result tensor
	resultSchema := TensorSchema{
		Shape:       broadcastShape,
		DType:       t.schema.DType,
		ChunkSize:   t.schema.ChunkSize,
		Compression: t.schema.Compression,
		Metadata:    map[string]interface{}{"operation": "multiply"},
	}

	result := &tensorImpl{
		name:   fmt.Sprintf("%s_times_%s", t.name, otherTensor.name),
		schema: resultSchema,
		engine: t.engine,
		data:   make([]float32, len(broadcastedT.data)),
	}

	// Perform element-wise multiplication
	for i := range broadcastedT.data {
		result.data[i] = broadcastedT.data[i] * broadcastedOther.data[i]
	}

	return result, nil
}

func (t *tensorImpl) applyMatrixMultiplyOperation(op Operation) (Tensor, error) {
	otherTensor, ok := op.Operand.(*tensorImpl)
	if !ok {
		return nil, fmt.Errorf("operand must be a tensor")
	}

	// Check if both tensors are 2D matrices
	if len(t.schema.Shape) != 2 || len(otherTensor.schema.Shape) != 2 {
		return nil, fmt.Errorf("matrix multiplication requires 2D tensors")
	}

	// Check matrix dimensions: (m x n) * (n x p) = (m x p)
	m, n := t.schema.Shape[0], t.schema.Shape[1]
	n2, p := otherTensor.schema.Shape[0], otherTensor.schema.Shape[1]

	if n != n2 {
		return nil, fmt.Errorf("matrix dimensions incompatible: (%d x %d) * (%d x %d)", m, n, n2, p)
	}

	// Create result tensor (m x p)
	resultSchema := TensorSchema{
		Shape:       []int{m, p},
		DType:       t.schema.DType,
		ChunkSize:   t.schema.ChunkSize,
		Compression: t.schema.Compression,
		Metadata:    map[string]interface{}{"operation": "matrix_multiply"},
	}

	result := &tensorImpl{
		name:   fmt.Sprintf("%s_matmul_%s", t.name, otherTensor.name),
		schema: resultSchema,
		engine: t.engine,
		data:   make([]float32, m*p),
	}

	// Perform matrix multiplication
	for i := 0; i < m; i++ {
		for j := 0; j < p; j++ {
			sum := float32(0)
			for k := 0; k < n; k++ {
				// Get elements from both matrices
				aIdx := i*n + k
				bIdx := k*p + j
				sum += t.data[aIdx] * otherTensor.data[bIdx]
			}
			result.data[i*p+j] = sum
		}
	}

	return result, nil
}

func (t *tensorImpl) applyTransposeOperation(op Operation) (Tensor, error) {
	// Check if tensor is 2D
	if len(t.schema.Shape) != 2 {
		return nil, fmt.Errorf("transpose requires 2D tensor")
	}

	rows, cols := t.schema.Shape[0], t.schema.Shape[1]

	// Create transposed tensor (cols x rows)
	resultSchema := TensorSchema{
		Shape:       []int{cols, rows},
		DType:       t.schema.DType,
		ChunkSize:   t.schema.ChunkSize,
		Compression: t.schema.Compression,
		Metadata:    map[string]interface{}{"operation": "transpose"},
	}

	result := &tensorImpl{
		name:   fmt.Sprintf("%s_transpose", t.name),
		schema: resultSchema,
		engine: t.engine,
		data:   make([]float32, cols*rows),
	}

	// Perform transpose
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			srcIdx := i*cols + j
			dstIdx := j*rows + i
			result.data[dstIdx] = t.data[srcIdx]
		}
	}

	return result, nil
}

func (t *tensorImpl) applyReductionOperation(op Operation, reductionType string) (Tensor, error) {
	// Get axis from parameters (default: reduce all axes)
	axis := -1 // Default: reduce all dimensions
	if axisParam, ok := op.Params["axis"]; ok {
		if axisInt, ok := axisParam.(int); ok {
			axis = axisInt
		}
	}

	var resultShape []int
	var resultData []float32

	if axis == -1 {
		// Reduce all dimensions to scalar
		resultShape = []int{1}
		resultData = []float32{t.reduceAll(reductionType)}
	} else {
		// Reduce along specific axis
		if axis < 0 || axis >= len(t.schema.Shape) {
			return nil, fmt.Errorf("axis %d out of bounds for tensor with %d dimensions", axis, len(t.schema.Shape))
		}

		// Calculate result shape
		resultShape = make([]int, len(t.schema.Shape)-1)
		copy(resultShape, t.schema.Shape[:axis])
		copy(resultShape[axis:], t.schema.Shape[axis+1:])

		// Perform reduction along axis
		resultData = t.reduceAlongAxis(axis, reductionType)
	}

	// Create result tensor
	resultSchema := TensorSchema{
		Shape:       resultShape,
		DType:       t.schema.DType,
		ChunkSize:   t.schema.ChunkSize,
		Compression: t.schema.Compression,
		Metadata:    map[string]interface{}{"operation": reductionType, "axis": axis},
	}

	result := &tensorImpl{
		name:   fmt.Sprintf("%s_%s", t.name, reductionType),
		schema: resultSchema,
		engine: t.engine,
		data:   resultData,
	}

	return result, nil
}

func (t *tensorImpl) reduceAll(reductionType string) float32 {
	switch reductionType {
	case "sum":
		sum := float32(0)
		for _, v := range t.data {
			sum += v
		}
		return sum
	case "mean":
		if len(t.data) == 0 {
			return 0
		}
		sum := float32(0)
		for _, v := range t.data {
			sum += v
		}
		return sum / float32(len(t.data))
	case "max":
		if len(t.data) == 0 {
			return 0
		}
		max := t.data[0]
		for _, v := range t.data[1:] {
			if v > max {
				max = v
			}
		}
		return max
	case "min":
		if len(t.data) == 0 {
			return 0
		}
		min := t.data[0]
		for _, v := range t.data[1:] {
			if v < min {
				min = v
			}
		}
		return min
	default:
		return 0
	}
}

func (t *tensorImpl) reduceAlongAxis(axis int, reductionType string) []float32 {
	// Calculate the size of the result
	resultSize := 1
	for i, dim := range t.schema.Shape {
		if i != axis {
			resultSize *= dim
		}
	}

	result := make([]float32, resultSize)
	axisSize := t.schema.Shape[axis]

	// For each position in the result, reduce along the specified axis
	for resultIdx := 0; resultIdx < resultSize; resultIdx++ {
		// Convert result index to multi-dimensional indices
		resultIndices := make([]int, len(t.schema.Shape)-1)
		temp := resultIdx
		for i := len(resultIndices) - 1; i >= 0; i-- {
			dimIdx := i
			if i >= axis {
				dimIdx++
			}
			resultIndices[i] = temp % t.schema.Shape[dimIdx]
			temp /= t.schema.Shape[dimIdx]
		}

		// Build full indices for the original tensor
		var values []float32
		for axisPos := 0; axisPos < axisSize; axisPos++ {
			fullIndices := make([]int, len(t.schema.Shape))
			copy(fullIndices[:axis], resultIndices[:axis])
			fullIndices[axis] = axisPos
			copy(fullIndices[axis+1:], resultIndices[axis:])

			flatIdx := t.calculateFlatIndex(fullIndices)
			values = append(values, t.data[flatIdx])
		}

		// Apply reduction to the collected values
		result[resultIdx] = t.reduceValues(values, reductionType)
	}

	return result
}

func (t *tensorImpl) reduceValues(values []float32, reductionType string) float32 {
	switch reductionType {
	case "sum":
		sum := float32(0)
		for _, v := range values {
			sum += v
		}
		return sum
	case "mean":
		if len(values) == 0 {
			return 0
		}
		sum := float32(0)
		for _, v := range values {
			sum += v
		}
		return sum / float32(len(values))
	case "max":
		if len(values) == 0 {
			return 0
		}
		max := values[0]
		for _, v := range values[1:] {
			if v > max {
				max = v
			}
		}
		return max
	case "min":
		if len(values) == 0 {
			return 0
		}
		min := values[0]
		for _, v := range values[1:] {
			if v < min {
				min = v
			}
		}
		return min
	default:
		return 0
	}
}

func (t *tensorImpl) applyActivationFunction(op Operation, activationType string) (Tensor, error) {
	// Create result tensor with same shape
	resultSchema := TensorSchema{
		Shape:       t.schema.Shape,
		DType:       t.schema.DType,
		ChunkSize:   t.schema.ChunkSize,
		Compression: t.schema.Compression,
		Metadata:    map[string]interface{}{"operation": activationType},
	}

	result := &tensorImpl{
		name:   fmt.Sprintf("%s_%s", t.name, activationType),
		schema: resultSchema,
		engine: t.engine,
		data:   make([]float32, len(t.data)),
	}

	// Apply activation function element-wise
	for i, value := range t.data {
		switch activationType {
		case "relu":
			if value > 0 {
				result.data[i] = value
			} else {
				result.data[i] = 0
			}
		case "sigmoid":
			result.data[i] = float32(1.0 / (1.0 + math.Exp(-float64(value))))
		case "tanh":
			result.data[i] = float32(math.Tanh(float64(value)))
		}
	}

	return result, nil
}

func (t *tensorImpl) applyConv1DOperation(op Operation) (Tensor, error) {
	// Get kernel from operand
	kernel, ok := op.Operand.(*tensorImpl)
	if !ok {
		return nil, fmt.Errorf("kernel must be a tensor")
	}

	// Validate 1D convolution: input should be 1D, kernel should be 1D
	if len(t.schema.Shape) != 1 || len(kernel.schema.Shape) != 1 {
		return nil, fmt.Errorf("conv1d requires 1D input and kernel tensors")
	}

	inputSize := t.schema.Shape[0]
	kernelSize := kernel.schema.Shape[0]

	// Get stride and padding from parameters
	stride := 1
	if strideParam, ok := op.Params["stride"]; ok {
		if s, ok := strideParam.(int); ok {
			stride = s
		}
	}

	padding := 0
	if paddingParam, ok := op.Params["padding"]; ok {
		if p, ok := paddingParam.(int); ok {
			padding = p
		}
	}

	// Calculate output size
	outputSize := ((inputSize + 2*padding - kernelSize) / stride) + 1
	if outputSize <= 0 {
		return nil, fmt.Errorf("invalid output size: %d", outputSize)
	}

	// Create result tensor
	resultSchema := TensorSchema{
		Shape:       []int{outputSize},
		DType:       t.schema.DType,
		ChunkSize:   t.schema.ChunkSize,
		Compression: t.schema.Compression,
		Metadata:    map[string]interface{}{"operation": "conv1d", "kernel_size": kernelSize, "stride": stride, "padding": padding},
	}

	result := &tensorImpl{
		name:   fmt.Sprintf("%s_conv1d", t.name),
		schema: resultSchema,
		engine: t.engine,
		data:   make([]float32, outputSize),
	}

	// Perform 1D convolution
	for outIdx := 0; outIdx < outputSize; outIdx++ {
		sum := float32(0)
		for k := 0; k < kernelSize; k++ {
			inputIdx := outIdx*stride + k - padding
			if inputIdx >= 0 && inputIdx < inputSize {
				sum += t.data[inputIdx] * kernel.data[kernelSize-1-k] // Flip kernel
			}
		}
		result.data[outIdx] = sum
	}

	return result, nil
}

func (t *tensorImpl) applyConv2DOperation(op Operation) (Tensor, error) {
	// Get kernel from operand
	kernel, ok := op.Operand.(*tensorImpl)
	if !ok {
		return nil, fmt.Errorf("kernel must be a tensor")
	}

	// Validate 2D convolution: input should be 2D or 3D (H,W) or (C,H,W), kernel should be 3D or 4D
	if len(t.schema.Shape) < 2 || len(kernel.schema.Shape) < 2 {
		return nil, fmt.Errorf("conv2d requires at least 2D input and kernel tensors")
	}

	// Simplified implementation for 2D input (H,W) and 2D kernel (KH,KW)
	if len(t.schema.Shape) == 2 && len(kernel.schema.Shape) == 2 {
		return t.applyConv2DOperation2D(kernel, op)
	}

	return nil, fmt.Errorf("complex conv2d not yet implemented")
}

func (t *tensorImpl) applyConv2DOperation2D(kernel *tensorImpl, op Operation) (Tensor, error) {
	inputH, inputW := t.schema.Shape[0], t.schema.Shape[1]
	kernelH, kernelW := kernel.schema.Shape[0], kernel.schema.Shape[1]

	// Get stride and padding from parameters
	strideH, strideW := 1, 1
	if strideParam, ok := op.Params["stride"]; ok {
		if s, ok := strideParam.([]int); ok && len(s) == 2 {
			strideH, strideW = s[0], s[1]
		}
	}

	paddingH, paddingW := 0, 0
	if paddingParam, ok := op.Params["padding"]; ok {
		if p, ok := paddingParam.([]int); ok && len(p) == 2 {
			paddingH, paddingW = p[0], p[1]
		}
	}

	// Calculate output size
	outputH := ((inputH + 2*paddingH - kernelH) / strideH) + 1
	outputW := ((inputW + 2*paddingW - kernelW) / strideW) + 1

	if outputH <= 0 || outputW <= 0 {
		return nil, fmt.Errorf("invalid output size: %dx%d", outputH, outputW)
	}

	// Create result tensor
	resultSchema := TensorSchema{
		Shape:       []int{outputH, outputW},
		DType:       t.schema.DType,
		ChunkSize:   t.schema.ChunkSize,
		Compression: t.schema.Compression,
		Metadata:    map[string]interface{}{"operation": "conv2d", "kernel_size": []int{kernelH, kernelW}, "stride": []int{strideH, strideW}, "padding": []int{paddingH, paddingW}},
	}

	result := &tensorImpl{
		name:   fmt.Sprintf("%s_conv2d", t.name),
		schema: resultSchema,
		engine: t.engine,
		data:   make([]float32, outputH*outputW),
	}

	// Perform 2D convolution
	for outY := 0; outY < outputH; outY++ {
		for outX := 0; outX < outputW; outX++ {
			sum := float32(0)
			for ky := 0; ky < kernelH; ky++ {
				for kx := 0; kx < kernelW; kx++ {
					inputY := outY*strideH + ky - paddingH
					inputX := outX*strideW + kx - paddingW

					if inputY >= 0 && inputY < inputH && inputX >= 0 && inputX < inputW {
						inputIdx := inputY*inputW + inputX
						kernelIdx := (kernelH-1-ky)*kernelW + (kernelW - 1 - kx) // Flip kernel
						sum += t.data[inputIdx] * kernel.data[kernelIdx]
					}
				}
			}
			result.data[outY*outputW+outX] = sum
		}
	}

	return result, nil
}

func (t *tensorImpl) applySVDOperation(op Operation) (Tensor, error) {
	// Simplified SVD implementation for 2D matrices
	if len(t.schema.Shape) != 2 {
		return nil, fmt.Errorf("SVD requires 2D tensor")
	}

	m, n := t.schema.Shape[0], t.schema.Shape[1]

	// For now, return a simplified decomposition
	// In a real implementation, this would use a proper SVD algorithm like Golub-Reinsch

	// Create U, S, V matrices (simplified)
	// U: m x m, S: min(m,n) x 1, V: n x n
	k := min(m, n)

	// Create S tensor (singular values)
	sSchema := TensorSchema{
		Shape:       []int{k},
		DType:       t.schema.DType,
		ChunkSize:   []int{k},
		Compression: t.schema.Compression,
		Metadata:    map[string]interface{}{"operation": "svd_s"},
	}

	sTensor := &tensorImpl{
		name:   fmt.Sprintf("%s_svd_s", t.name),
		schema: sSchema,
		engine: t.engine,
		data:   make([]float32, k),
	}

	// Simplified singular values (just use diagonal elements)
	for i := 0; i < k; i++ {
		sTensor.data[i] = t.data[i*n+i] // Diagonal elements
	}

	return sTensor, nil
}

func (t *tensorImpl) applyEigenvaluesOperation(op Operation) (Tensor, error) {
	// Simplified eigenvalue computation for 2D matrices
	if len(t.schema.Shape) != 2 || t.schema.Shape[0] != t.schema.Shape[1] {
		return nil, fmt.Errorf("eigenvalues require square 2D tensor")
	}

	n := t.schema.Shape[0]

	// Create eigenvalues tensor
	eigenSchema := TensorSchema{
		Shape:       []int{n},
		DType:       t.schema.DType,
		ChunkSize:   []int{n},
		Compression: t.schema.Compression,
		Metadata:    map[string]interface{}{"operation": "eigenvalues"},
	}

	eigenTensor := &tensorImpl{
		name:   fmt.Sprintf("%s_eigenvalues", t.name),
		schema: eigenSchema,
		engine: t.engine,
		data:   make([]float32, n),
	}

	// Simplified eigenvalue computation for 2x2 case
	if n == 2 {
		a, b := t.data[0], t.data[1]
		c, d := t.data[2], t.data[3]

		trace := a + d
		det := a*d - b*c

		discriminant := trace*trace - 4*det
		if discriminant >= 0 {
			sqrtDisc := float32(math.Sqrt(float64(discriminant)))
			eigenTensor.data[0] = (trace + sqrtDisc) / 2
			eigenTensor.data[1] = (trace - sqrtDisc) / 2
		} else {
			// Complex eigenvalues - return real parts
			eigenTensor.data[0] = trace / 2
			eigenTensor.data[1] = trace / 2
		}
	} else {
		// For larger matrices, return diagonal elements as approximation
		for i := 0; i < n; i++ {
			eigenTensor.data[i] = t.data[i*n+i]
		}
	}

	return eigenTensor, nil
}

func (t *tensorImpl) applyCosineSimilarity(op Operation) (Tensor, error) {
	otherTensor, ok := op.Operand.(*tensorImpl)
	if !ok {
		return nil, fmt.Errorf("operand must be a tensor")
	}

	// Calculate cosine similarity
	similarity := cosineSimilarity(t.data, otherTensor.data)

	// Create result tensor (1x1 tensor)
	resultSchema := TensorSchema{
		Shape:       []int{1, 1},
		DType:       "float32",
		ChunkSize:   []int{1, 1},
		Compression: "none",
		Metadata:    map[string]interface{}{"operation": "cosine_similarity"},
	}

	result := &tensorImpl{
		name:   fmt.Sprintf("%s_cosine_%s", t.name, otherTensor.name),
		schema: resultSchema,
		engine: t.engine,
		data:   []float32{similarity},
	}

	return result, nil
}

// Utility functions

func bytesToFloat32Slice(data []byte) []float32 {
	if len(data)%4 != 0 {
		return nil
	}

	slice := (*[1 << 28]float32)(unsafe.Pointer(&data[0]))[:len(data)/4]
	return slice
}

func float32SliceToBytes(slice []float32) []byte {
	return (*[1 << 28]byte)(unsafe.Pointer(&slice[0]))[:len(slice)*4]
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = float32(math.Sqrt(float64(normA)))
	normB = float32(math.Sqrt(float64(normB)))

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}
