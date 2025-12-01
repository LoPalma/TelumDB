"""
TelumDB Tensor Operations

Provides tensor storage and manipulation capabilities with NumPy integration.
"""

from typing import List, Dict, Any, Optional, Union, Tuple
import warnings

try:
    import numpy as np
    _NUMPY_AVAILABLE = True
except ImportError:
    _NUMPY_AVAILABLE = False
    warnings.warn("NumPy not available. Tensor operations will be limited.", ImportWarning)

from .exceptions import TensorError, ValidationError


class TensorSchema:
    """Schema definition for tensors."""
    
    def __init__(
        self,
        shape: List[int],
        dtype: str = "float32",
        chunk_size: Optional[List[int]] = None,
        compression: str = "lz4",
        metadata: Optional[Dict[str, Any]] = None
    ):
        self.shape = shape
        self.dtype = dtype
        self.chunk_size = chunk_size or [64] * len(shape)
        self.compression = compression
        self.metadata = metadata or {}
    
    def validate(self) -> None:
        """Validate tensor schema."""
        if not self.shape:
            raise ValidationError("Tensor shape cannot be empty")
        
        if any(dim <= 0 for dim in self.shape):
            raise ValidationError("All dimensions must be positive")
        
        if len(self.chunk_size) != len(self.shape):
            raise ValidationError("Chunk size must match tensor dimensions")
        
        if any(chunk <= 0 for chunk in self.chunk_size):
            raise ValidationError("All chunk sizes must be positive")
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert schema to dictionary."""
        return {
            "shape": self.shape,
            "dtype": self.dtype,
            "chunk_size": self.chunk_size,
            "compression": self.compression,
            "metadata": self.metadata
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "TensorSchema":
        """Create schema from dictionary."""
        return cls(
            shape=data["shape"],
            dtype=data.get("dtype", "float32"),
            chunk_size=data.get("chunk_size"),
            compression=data.get("compression", "lz4"),
            metadata=data.get("metadata", {})
        )


class Tensor:
    """Tensor object for storing and manipulating multi-dimensional data."""
    
    def __init__(
        self,
        name: str,
        shape: List[int],
        dtype: str = "float32",
        chunk_size: Optional[List[int]] = None,
        compression: str = "lz4",
        metadata: Optional[Dict[str, Any]] = None,
        client: Optional[Any] = None
    ):
        self.name = name
        self.shape = shape
        self.dtype = dtype
        self.chunk_size = chunk_size or [64] * len(shape)
        self.compression = compression
        self.metadata = metadata or {}
        self.client = client
        
        # Validate tensor parameters
        self._validate_parameters()
    
    def _validate_parameters(self) -> None:
        """Validate tensor parameters."""
        if not self.shape:
            raise ValidationError("Tensor shape cannot be empty")
        
        if any(dim <= 0 for dim in self.shape):
            raise ValidationError("All dimensions must be positive")
        
        if len(self.chunk_size) != len(self.shape):
            raise ValidationError("Chunk size must match tensor dimensions")
    
    def store_chunk(
        self, 
        indices: List[int], 
        data: Union["np.ndarray", List[Any]]
    ) -> None:
        """
        Store a chunk of tensor data.
        
        Args:
            indices: Chunk indices
            data: Chunk data as NumPy array or list
        """
        if not self.client:
            raise TensorError("No client available for tensor operations", self.name)
        
        # Convert data to NumPy array if needed
        if _NUMPY_AVAILABLE and not isinstance(data, np.ndarray):
            data = np.array(data, dtype=self.dtype)
        
        # TODO: Implement actual chunk storage
        print(f"Storing chunk at indices {indices} with shape {getattr(data, 'shape', len(data))}")
    
    def get_chunk(self, indices: List[int]) -> Union["np.ndarray", List[Any]]:
        """
        Get a chunk of tensor data.
        
        Args:
            indices: Chunk indices
            
        Returns:
            Chunk data as NumPy array or list
        """
        if not self.client:
            raise TensorError("No client available for tensor operations", self.name)
        
        # TODO: Implement actual chunk retrieval
        if _NUMPY_AVAILABLE:
            return np.zeros(self.chunk_size, dtype=self.dtype)
        else:
            return [[0] * self.chunk_size[-1] for _ in range(self.chunk_size[-2])]
    
    def slice(self, ranges: List[Union[int, Tuple[int, int]]]) -> "Tensor":
        """
        Slice tensor along specified dimensions.
        
        Args:
            ranges: List of slices, each can be an int or (start, end) tuple
            
        Returns:
            New tensor representing the slice
        """
        # TODO: Implement tensor slicing
        return Tensor(
            name=f"{self.name}_slice",
            shape=self.shape,  # Would be updated based on slice
            dtype=self.dtype,
            chunk_size=self.chunk_size,
            compression=self.compression,
            metadata={"parent": self.name, "slice": ranges},
            client=self.client
        )
    
    def reshape(self, new_shape: List[int]) -> "Tensor":
        """
        Reshape tensor to new dimensions.
        
        Args:
            new_shape: New tensor shape
            
        Returns:
            New tensor with reshaped dimensions
        """
        # Validate new shape
        total_elements = 1
        for dim in self.shape:
            total_elements *= dim
        
        new_total = 1
        for dim in new_shape:
            new_total *= dim
        
        if total_elements != new_total:
            raise ValidationError(f"Cannot reshape: {total_elements} elements != {new_total} elements")
        
        return Tensor(
            name=f"{self.name}_reshaped",
            shape=new_shape,
            dtype=self.dtype,
            chunk_size=self.chunk_size,
            compression=self.compression,
            metadata={"parent": self.name, "original_shape": self.shape},
            client=self.client
        )
    
    def apply_operation(self, operation: str, *args) -> "Tensor":
        """
        Apply mathematical operation to tensor.
        
        Args:
            operation: Operation name (e.g., 'add', 'multiply', 'transpose')
            *args: Operation arguments
            
        Returns:
            New tensor with operation applied
        """
        # TODO: Implement tensor operations
        return Tensor(
            name=f"{self.name}_{operation}",
            shape=self.shape,
            dtype=self.dtype,
            chunk_size=self.chunk_size,
            compression=self.compression,
            metadata={"parent": self.name, "operation": operation, "args": args},
            client=self.client
        )
    
    def to_numpy(self) -> "np.ndarray":
        """
        Convert tensor to NumPy array.
        
        Returns:
            NumPy array representation of tensor
        """
        if not _NUMPY_AVAILABLE:
            raise ImportError("NumPy is required for to_numpy() method")
        
        # TODO: Implement actual conversion
        return np.zeros(self.shape, dtype=self.dtype)
    
    def from_numpy(self, array: "np.ndarray") -> None:
        """
        Load tensor data from NumPy array.
        
        Args:
            array: NumPy array to load from
        """
        if not _NUMPY_AVAILABLE:
            raise ImportError("NumPy is required for from_numpy() method")
        
        if not isinstance(array, np.ndarray):
            raise ValidationError("Input must be a NumPy array")
        
        # Update shape if different
        if list(array.shape) != self.shape:
            self.shape = list(array.shape)
        
        # TODO: Implement actual data loading
        print(f"Loading NumPy array with shape {array.shape} into tensor {self.name}")
    
    def get_metadata(self, key: str) -> Any:
        """Get metadata value by key."""
        return self.metadata.get(key)
    
    def set_metadata(self, key: str, value: Any) -> None:
        """Set metadata value by key."""
        self.metadata[key] = value
    
    def get_schema(self) -> TensorSchema:
        """Get tensor schema."""
        return TensorSchema(
            shape=self.shape,
            dtype=self.dtype,
            chunk_size=self.chunk_size,
            compression=self.compression,
            metadata=self.metadata
        )
    
    def __repr__(self) -> str:
        """String representation of tensor."""
        return (
            f"Tensor(name='{self.name}', shape={self.shape}, "
            f"dtype='{self.dtype}', chunk_size={self.chunk_size})"
        )
    
    def __len__(self) -> int:
        """Return size of first dimension."""
        return self.shape[0] if self.shape else 0
    
    def __getitem__(self, key) -> "Tensor":
        """Support slicing syntax."""
        if isinstance(key, slice):
            # Convert slice to range format
            start = key.start if key.start is not None else 0
            stop = key.stop if key.stop is not None else self.shape[0]
            return self.slice([(start, stop)])
        elif isinstance(key, tuple):
            # Multi-dimensional slicing
            ranges = []
            for k in key:
                if isinstance(k, slice):
                    start = k.start if k.start is not None else 0
                    stop = k.stop if k.stop is not None else self.shape[0]
                    ranges.append((start, stop))
                else:
                    ranges.append(k)
            return self.slice(ranges)
        else:
            return self.slice([key])
    
    # Mathematical operations
    def __add__(self, other) -> "Tensor":
        """Addition operation."""
        return self.apply_operation("add", other)
    
    def __sub__(self, other) -> "Tensor":
        """Subtraction operation."""
        return self.apply_operation("subtract", other)
    
    def __mul__(self, other) -> "Tensor":
        """Multiplication operation."""
        return self.apply_operation("multiply", other)
    
    def __truediv__(self, other) -> "Tensor":
        """Division operation."""
        return self.apply_operation("divide", other)