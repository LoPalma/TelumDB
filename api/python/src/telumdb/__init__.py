"""
TelumDB Python Client

The World's First Hybrid General-Purpose + AI Tensor Database

This package provides Python bindings for TelumDB, allowing you to:
- Execute SQL queries on traditional data
- Store and retrieve tensor data
- Perform mixed SQL + tensor operations
- Use familiar Python APIs with NumPy integration
"""

__version__ = "0.1.0"
__author__ = "TelumDB Contributors"
__email__ = "contributors@telumdb.io"

from .client import Client, Connection
from .tensor import Tensor, TensorSchema
from .exceptions import (
    TelumDBError,
    ConnectionError,
    QueryError,
    TensorError,
    ConfigurationError,
)

# Import core C++ extension if available
try:
    from . import _core
    _CORE_AVAILABLE = True
except ImportError:
    _CORE_AVAILABLE = False
    import warnings
    warnings.warn(
        "C++ core extension not available. Some features may be slower.",
        ImportWarning
    )

__all__ = [
    # Version info
    "__version__",
    "__author__",
    "__email__",
    
    # Main classes
    "Client",
    "Connection",
    "Tensor",
    "TensorSchema",
    
    # Exceptions
    "TelumDBError",
    "ConnectionError", 
    "QueryError",
    "TensorError",
    "ConfigurationError",
    
    # Core extension
    "_core",
    "_CORE_AVAILABLE",
]

def get_version():
    """Get the current TelumDB Python client version."""
    return __version__

def get_build_info():
    """Get build information including core extension status."""
    info = {
        "version": __version__,
        "core_extension": _CORE_AVAILABLE,
    }
    if _CORE_AVAILABLE:
        info["core_version"] = getattr(_core, "version", "unknown")
    return info