package storage

import (
	"context"
	"testing"
)

func TestTensorOperationsDirect(t *testing.T) {
	ctx := context.Background()

	// Create test tensor directly
	tensor := &tensorImpl{
		name: "test",
		schema: TensorSchema{
			Shape:       []int{2, 3},
			DType:       "float32",
			ChunkSize:   []int{1, 1},
			Compression: "none",
			Metadata:    make(map[string]interface{}),
		},
		data: []float32{1, 2, 3, 4, 5, 6},
	}

	// Test all operations
	t.Run("AddOperation", testAddOperationDirect(tensor, ctx))
	t.Run("MultiplyOperation", testMultiplyOperationDirect(tensor, ctx))
	t.Run("MatrixMultiply", testMatrixMultiplyDirect(tensor, ctx))
	t.Run("Transpose", testTransposeDirect(tensor, ctx))
	t.Run("Reductions", testReductionsDirect(tensor, ctx))
	t.Run("ActivationFunctions", testActivationFunctionsDirect(tensor, ctx))
	t.Run("Convolution1D", testConvolution1DDirect(tensor, ctx))
	t.Run("Convolution2D", testConvolution2DDirect(tensor, ctx))
	t.Run("SVD", testSVDDirect(tensor, ctx))
	t.Run("Eigenvalues", testEigenvaluesDirect(tensor, ctx))
	t.Run("Slicing", testSlicingDirect(tensor, ctx))
	t.Run("Broadcasting", testBroadcastingDirect(tensor, ctx))
}

func testAddOperationDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		otherTensor := &tensorImpl{
			name: "other",
			schema: TensorSchema{
				Shape:       []int{2, 3},
				DType:       "float32",
				ChunkSize:   []int{1, 1},
				Compression: "none",
			},
			data: []float32{1, 2, 3, 4, 5, 6},
		}

		op := Operation{
			Type:    "add",
			Operand: otherTensor,
		}

		result, err := tensor.ApplyOperation(ctx, op)
		if err != nil {
			t.Fatalf("Add operation failed: %v", err)
		}

		if result.Shape()[0] != 2 || result.Shape()[1] != 3 {
			t.Errorf("Expected shape [2,3], got %v", result.Shape())
		}
	}
}

func testMultiplyOperationDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		otherTensor := &tensorImpl{
			name: "other",
			schema: TensorSchema{
				Shape:       []int{2, 3},
				DType:       "float32",
				ChunkSize:   []int{1, 1},
				Compression: "none",
			},
			data: []float32{2, 2, 2, 2, 2, 2},
		}

		op := Operation{
			Type:    "multiply",
			Operand: otherTensor,
		}

		result, err := tensor.ApplyOperation(ctx, op)
		if err != nil {
			t.Fatalf("Multiply operation failed: %v", err)
		}

		if result.Shape()[0] != 2 || result.Shape()[1] != 3 {
			t.Errorf("Expected shape [2,3], got %v", result.Shape())
		}
	}
}

func testMatrixMultiplyDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		matrixA := &tensorImpl{
			name: "matrixA",
			schema: TensorSchema{
				Shape:       []int{2, 3},
				DType:       "float32",
				ChunkSize:   []int{1, 1},
				Compression: "none",
			},
			data: []float32{1, 2, 3, 4, 5, 6},
		}

		matrixB := &tensorImpl{
			name: "matrixB",
			schema: TensorSchema{
				Shape:       []int{3, 2},
				DType:       "float32",
				ChunkSize:   []int{1, 1},
				Compression: "none",
			},
			data: []float32{7, 8, 9, 10, 11, 12},
		}

		op := Operation{
			Type:    "matrix_multiply",
			Operand: matrixB,
		}

		result, err := matrixA.ApplyOperation(ctx, op)
		if err != nil {
			t.Fatalf("Matrix multiply operation failed: %v", err)
		}

		if result.Shape()[0] != 2 || result.Shape()[1] != 2 {
			t.Errorf("Expected shape [2,2], got %v", result.Shape())
		}
	}
}

func testTransposeDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		op := Operation{
			Type: "transpose",
		}

		result, err := tensor.ApplyOperation(ctx, op)
		if err != nil {
			t.Fatalf("Transpose operation failed: %v", err)
		}

		if result.Shape()[0] != 3 || result.Shape()[1] != 2 {
			t.Errorf("Expected shape [3,2], got %v", result.Shape())
		}
	}
}

func testReductionsDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		tests := []struct {
			name     string
			opType   string
			axis     int
			expected int
		}{
			{"SumAll", "sum", -1, 1},
			{"MeanAll", "mean", -1, 1},
			{"MaxAll", "max", -1, 1},
			{"MinAll", "min", -1, 1},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				op := Operation{
					Type: tt.opType,
					Params: map[string]interface{}{
						"axis": tt.axis,
					},
				}

				result, err := tensor.ApplyOperation(ctx, op)
				if err != nil {
					t.Fatalf("%s operation failed: %v", tt.name, err)
				}

				if len(result.Shape()) != tt.expected {
					t.Errorf("%s: expected %d dimensions, got %d", tt.name, tt.expected, len(result.Shape()))
				}
			})
		}
	}
}

func testActivationFunctionsDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		activations := []string{"relu", "sigmoid", "tanh"}

		for _, activation := range activations {
			t.Run(activation, func(t *testing.T) {
				op := Operation{
					Type: activation,
				}

				result, err := tensor.ApplyOperation(ctx, op)
				if err != nil {
					t.Fatalf("%s activation failed: %v", activation, err)
				}

				if len(result.Shape()) != len(tensor.Shape()) {
					t.Errorf("%s: shape changed from %v to %v", activation, tensor.Shape(), result.Shape())
				}
			})
		}
	}
}

func testConvolution1DDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		input1D := &tensorImpl{
			name: "input1d",
			schema: TensorSchema{
				Shape:       []int{10},
				DType:       "float32",
				ChunkSize:   []int{5},
				Compression: "none",
			},
			data: []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		}

		kernel1D := &tensorImpl{
			name: "kernel1d",
			schema: TensorSchema{
				Shape:       []int{3},
				DType:       "float32",
				ChunkSize:   []int{3},
				Compression: "none",
			},
			data: []float32{1, 0, -1},
		}

		op := Operation{
			Type:    "conv1d",
			Operand: kernel1D,
			Params: map[string]interface{}{
				"stride":  1,
				"padding": 0,
			},
		}

		result, err := input1D.ApplyOperation(ctx, op)
		if err != nil {
			t.Fatalf("Conv1D operation failed: %v", err)
		}

		if result.Shape()[0] != 8 {
			t.Errorf("Expected output size 8, got %d", result.Shape()[0])
		}
	}
}

func testConvolution2DDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		input2D := &tensorImpl{
			name: "input2d",
			schema: TensorSchema{
				Shape:       []int{5, 5},
				DType:       "float32",
				ChunkSize:   []int{2, 2},
				Compression: "none",
			},
			data: make([]float32, 25),
		}

		for i := range input2D.data {
			input2D.data[i] = float32(i + 1)
		}

		kernel2D := &tensorImpl{
			name: "kernel2d",
			schema: TensorSchema{
				Shape:       []int{3, 3},
				DType:       "float32",
				ChunkSize:   []int{3, 3},
				Compression: "none",
			},
			data: []float32{1, 0, -1, 0, 0, 0, -1, 0, 1},
		}

		op := Operation{
			Type:    "conv2d",
			Operand: kernel2D,
			Params: map[string]interface{}{
				"stride":  []int{1, 1},
				"padding": []int{0, 0},
			},
		}

		result, err := input2D.ApplyOperation(ctx, op)
		if err != nil {
			t.Fatalf("Conv2D operation failed: %v", err)
		}

		if result.Shape()[0] != 3 || result.Shape()[1] != 3 {
			t.Errorf("Expected output shape [3,3], got %v", result.Shape())
		}
	}
}

func testSVDDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		op := Operation{
			Type: "svd",
		}

		result, err := tensor.ApplyOperation(ctx, op)
		if err != nil {
			t.Fatalf("SVD operation failed: %v", err)
		}

		if len(result.Shape()) != 1 || result.Shape()[0] != 2 {
			t.Errorf("Expected singular values shape [2], got %v", result.Shape())
		}
	}
}

func testEigenvaluesDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		squareMatrix := &tensorImpl{
			name: "square",
			schema: TensorSchema{
				Shape:       []int{2, 2},
				DType:       "float32",
				ChunkSize:   []int{1, 1},
				Compression: "none",
			},
			data: []float32{4, 2, 1, 3},
		}

		op := Operation{
			Type: "eigenvalues",
		}

		result, err := squareMatrix.ApplyOperation(ctx, op)
		if err != nil {
			t.Fatalf("Eigenvalues operation failed: %v", err)
		}

		if len(result.Shape()) != 1 || result.Shape()[0] != 2 {
			t.Errorf("Expected eigenvalues shape [2], got %v", result.Shape())
		}
	}
}

func testSlicingDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		ranges := []Range{
			{Start: 0, End: 1},
			{Start: 1, End: 3},
		}

		result, err := tensor.Slice(ctx, ranges)
		if err != nil {
			t.Fatalf("Slicing failed: %v", err)
		}

		if result.Shape()[0] != 1 || result.Shape()[1] != 2 {
			t.Errorf("Expected shape [1,2], got %v", result.Shape())
		}
	}
}

func testBroadcastingDirect(tensor *tensorImpl, ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		smallTensor := &tensorImpl{
			name: "small",
			schema: TensorSchema{
				Shape:       []int{1, 3},
				DType:       "float32",
				ChunkSize:   []int{1, 1},
				Compression: "none",
			},
			data: []float32{1, 2, 3},
		}

		op := Operation{
			Type:    "add",
			Operand: smallTensor,
		}

		result, err := tensor.ApplyOperation(ctx, op)
		if err != nil {
			t.Fatalf("Broadcasting add failed: %v", err)
		}

		if result.Shape()[0] != 2 || result.Shape()[1] != 3 {
			t.Errorf("Expected broadcasted shape [2,3], got %v", result.Shape())
		}
	}
}
