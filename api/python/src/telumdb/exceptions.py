"""
TelumDB Python Client Exceptions

Custom exception classes for different types of errors that can occur
when interacting with TelumDB.
"""


class TelumDBError(Exception):
    """Base exception class for all TelumDB errors."""
    
    def __init__(self, message: str, error_code: Optional[str] = None):
        super().__init__(message)
        self.message = message
        self.error_code = error_code
    
    def __str__(self) -> str:
        if self.error_code:
            return f"[{self.error_code}] {self.message}"
        return self.message


class ConnectionError(TelumDBError):
    """Raised when connection to database fails."""
    
    def __init__(self, message: str, host: Optional[str] = None, port: Optional[int] = None):
        super().__init__(message, "CONNECTION_ERROR")
        self.host = host
        self.port = port


class QueryError(TelumDBError):
    """Raised when query execution fails."""
    
    def __init__(self, message: str, query: Optional[str] = None):
        super().__init__(message, "QUERY_ERROR")
        self.query = query


class TensorError(TelumDBError):
    """Raised when tensor operations fail."""
    
    def __init__(self, message: str, tensor_name: Optional[str] = None):
        super().__init__(message, "TENSOR_ERROR")
        self.tensor_name = tensor_name


class ConfigurationError(TelumDBError):
    """Raised when configuration is invalid."""
    
    def __init__(self, message: str, config_key: Optional[str] = None):
        super().__init__(message, "CONFIG_ERROR")
        self.config_key = config_key


class AuthenticationError(TelumDBError):
    """Raised when authentication fails."""
    
    def __init__(self, message: str):
        super().__init__(message, "AUTH_ERROR")


class PermissionError(TelumDBError):
    """Raised when operation is not permitted."""
    
    def __init__(self, message: str, operation: Optional[str] = None):
        super().__init__(message, "PERMISSION_ERROR")
        self.operation = operation


class TimeoutError(TelumDBError):
    """Raised when operation times out."""
    
    def __init__(self, message: str, timeout: Optional[float] = None):
        super().__init__(message, "TIMEOUT_ERROR")
        self.timeout = timeout


class ValidationError(TelumDBError):
    """Raised when data validation fails."""
    
    def __init__(self, message: str, field: Optional[str] = None):
        super().__init__(message, "VALIDATION_ERROR")
        self.field = field


class ResourceError(TelumDBError):
    """Raised when resource limits are exceeded."""
    
    def __init__(self, message: str, resource_type: Optional[str] = None):
        super().__init__(message, "RESOURCE_ERROR")
        self.resource_type = resource_type


class TransactionError(TelumDBError):
    """Raised when transaction operations fail."""
    
    def __init__(self, message: str, transaction_id: Optional[str] = None):
        super().__init__(message, "TRANSACTION_ERROR")
        self.transaction_id = transaction_id


# Import Optional for type hints
try:
    from typing import Optional
except ImportError:
    # Fallback for older Python versions
    Optional = type