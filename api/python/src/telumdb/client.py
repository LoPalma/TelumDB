"""
TelumDB Python Client

Provides high-level client interface for connecting to and interacting with TelumDB.
"""

import socket
import json
import time
from typing import Optional, Dict, Any, List, Union, ContextManager
from contextlib import contextmanager

import numpy as np

from .exceptions import ConnectionError, QueryError, ConfigurationError
from .tensor import Tensor


class Client:
    """Main TelumDB client class."""
    
    def __init__(
        self,
        host: str = "localhost",
        port: int = 5432,
        database: Optional[str] = None,
        username: Optional[str] = None,
        password: Optional[str] = None,
        timeout: int = 30,
        use_ssl: bool = False,
    ):
        """
        Initialize TelumDB client.
        
        Args:
            host: Database server host
            port: Database server port
            database: Database name
            username: Username for authentication
            password: Password for authentication
            timeout: Connection timeout in seconds
            use_ssl: Whether to use SSL connection
        """
        self.host = host
        self.port = port
        self.database = database
        self.username = username
        self.password = password
        self.timeout = timeout
        self.use_ssl = use_ssl
        
        self._socket: Optional[socket.socket] = None
        self._connected = False
        self._session_id: Optional[str] = None
    
    def connect(self) -> None:
        """Connect to the database server."""
        if self._connected:
            return
            
        try:
            # Create socket connection
            self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self._socket.settimeout(self.timeout)
            
            # Connect to server
            self._socket.connect((self.host, self.port))
            
            # TODO: Implement authentication handshake
            self._session_id = f"session_{int(time.time())}"
            self._connected = True
            
        except socket.error as e:
            raise ConnectionError(f"Failed to connect to {self.host}:{self.port}: {e}")
    
    def disconnect(self) -> None:
        """Disconnect from the database server."""
        if self._socket:
            self._socket.close()
            self._socket = None
        self._connected = False
        self._session_id = None
    
    def is_connected(self) -> bool:
        """Check if client is connected to server."""
        return self._connected
    
    def execute(
        self, 
        query: str, 
        params: Optional[List[Any]] = None
    ) -> "Result":
        """
        Execute a SQL/TQL query.
        
        Args:
            query: SQL or TQL query string
            params: Query parameters for parameterized queries
            
        Returns:
            Result object containing query results
            
        Raises:
            ConnectionError: If not connected to server
            QueryError: If query execution fails
        """
        if not self._connected:
            self.connect()
        
        try:
            # TODO: Implement actual query execution
            # For now, return mock result
            return Result(
                columns=["result"],
                rows=[[f"Mock result for: {query}"]],
                affected=0
            )
            
        except Exception as e:
            raise QueryError(f"Query execution failed: {e}")
    
    def execute_many(
        self, 
        query: str, 
        params_list: List[List[Any]]
    ) -> "Result":
        """
        Execute a query multiple times with different parameter sets.
        
        Args:
            query: SQL or TQL query string
            params_list: List of parameter sets
            
        Returns:
            Result object containing combined results
        """
        # TODO: Implement batch execution
        return self.execute(query)
    
    def create_table(
        self, 
        name: str, 
        schema: Dict[str, str],
        if_not_exists: bool = False
    ) -> None:
        """
        Create a new table.
        
        Args:
            name: Table name
            schema: Dictionary mapping column names to data types
            if_not_exists: Don't raise error if table already exists
        """
        columns = []
        for col_name, col_type in schema.items():
            columns.append(f"{col_name} {col_type}")
        
        if_not_exists_clause = "IF NOT EXISTS " if if_not_exists else ""
        query = f"CREATE TABLE {if_not_exists_clause}{name} ({', '.join(columns)})"
        
        self.execute(query)
    
    def create_tensor(
        self,
        name: str,
        shape: List[int],
        dtype: str = "float32",
        chunk_size: Optional[List[int]] = None,
        compression: str = "lz4",
        metadata: Optional[Dict[str, Any]] = None
    ) -> Tensor:
        """
        Create a new tensor.
        
        Args:
            name: Tensor name
            shape: Tensor shape
            dtype: Data type
            chunk_size: Chunk size for storage
            compression: Compression algorithm
            metadata: Optional metadata dictionary
            
        Returns:
            Tensor object
        """
        if chunk_size is None:
            chunk_size = [64] * len(shape)
        
        # TODO: Implement actual tensor creation
        query = f"""
        CREATE TENSOR {name} (
            shape {shape},
            dtype {dtype},
            chunk_size {chunk_size},
            compression '{compression}'
        )
        """
        
        self.execute(query)
        
        return Tensor(
            name=name,
            shape=shape,
            dtype=dtype,
            chunk_size=chunk_size,
            compression=compression,
            metadata=metadata or {},
            client=self
        )
    
    def get_tensor(self, name: str) -> Tensor:
        """
        Get an existing tensor by name.
        
        Args:
            name: Tensor name
            
        Returns:
            Tensor object
            
        Raises:
            QueryError: If tensor doesn't exist
        """
        # TODO: Implement tensor retrieval
        return Tensor(
            name=name,
            shape=[],
            dtype="float32",
            chunk_size=[],
            compression="none",
            metadata={},
            client=self
        )
    
    def list_tables(self) -> List[str]:
        """List all tables in the database."""
        result = self.execute("SHOW TABLES")
        return [row[0] for row in result.rows]
    
    def list_tensors(self) -> List[str]:
        """List all tensors in the database."""
        result = self.execute("SHOW TENSORS")
        return [row[0] for row in result.rows]
    
    @contextmanager
    def transaction(self):
        """Context manager for database transactions."""
        # TODO: Implement transaction management
        self.execute("BEGIN")
        try:
            yield
            self.execute("COMMIT")
        except Exception:
            self.execute("ROLLBACK")
            raise
    
    def __enter__(self):
        """Context manager entry."""
        self.connect()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.disconnect()


class Result:
    """Query result object."""
    
    def __init__(
        self, 
        columns: List[str], 
        rows: List[List[Any]], 
        affected: int = 0
    ):
        self.columns = columns
        self.rows = rows
        self.affected = affected
    
    def __len__(self) -> int:
        """Return number of rows."""
        return len(self.rows)
    
    def __iter__(self):
        """Iterate over rows."""
        return iter(self.rows)
    
    def fetchone(self) -> Optional[List[Any]]:
        """Fetch one row."""
        if self.rows:
            return self.rows[0]
        return None
    
    def fetchall(self) -> List[List[Any]]:
        """Fetch all rows."""
        return self.rows
    
    def fetchmany(self, size: int) -> List[List[Any]]:
        """Fetch multiple rows."""
        return self.rows[:size]
    
    def to_pandas(self):
        """Convert result to pandas DataFrame."""
        try:
            import pandas as pd
            return pd.DataFrame(self.rows, columns=self.columns)
        except ImportError:
            raise ImportError("pandas is required for to_pandas() method")
    
    def to_numpy(self) -> np.ndarray:
        """Convert result to NumPy array."""
        return np.array(self.rows)


# Convenience function for quick connections
def connect(
    host: str = "localhost",
    port: int = 5432,
    database: Optional[str] = None,
    username: Optional[str] = None,
    password: Optional[str] = None,
    **kwargs
) -> Client:
    """
    Create and return a connected TelumDB client.
    
    Args:
        host: Database server host
        port: Database server port
        database: Database name
        username: Username for authentication
        password: Password for authentication
        **kwargs: Additional client arguments
        
    Returns:
        Connected Client instance
    """
    client = Client(
        host=host,
        port=port,
        database=database,
        username=username,
        password=password,
        **kwargs
    )
    client.connect()
    return client