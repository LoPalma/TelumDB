"""
TelumDB Connection Module

Handles low-level connection management to TelumDB servers.
"""

import socket
import json
import time
from typing import Optional, Dict, Any, Union
from contextlib import contextmanager

from .exceptions import ConnectionError, ConfigurationError


class Connection:
    """Low-level connection to TelumDB server."""
    
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
        Initialize connection parameters.
        
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
        """Establish connection to the server."""
        if self._connected:
            return
        
        try:
            # Create socket
            self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self._socket.settimeout(self.timeout)
            
            # Connect to server
            self._socket.connect((self.host, self.port))
            
            # Send authentication handshake
            self._authenticate()
            
            self._connected = True
            
        except socket.timeout:
            raise ConnectionError(f"Connection timeout to {self.host}:{self.port}")
        except socket.error as e:
            raise ConnectionError(f"Failed to connect to {self.host}:{self.port}: {e}")
    
    def disconnect(self) -> None:
        """Close the connection."""
        if self._socket:
            try:
                self._socket.close()
            except socket.error:
                pass
            finally:
                self._socket = None
                self._connected = False
                self._session_id = None
    
    def is_connected(self) -> bool:
        """Check if connection is active."""
        return self._connected and self._socket is not None
    
    def execute(self, query: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Execute a query and return the result.
        
        Args:
            query: SQL/TQL query string
            params: Query parameters
            
        Returns:
            Dictionary containing query results
        """
        if not self.is_connected():
            raise ConnectionError("Not connected to server")
        
        try:
            # Prepare request
            request = {
                "type": "query",
                "query": query,
                "params": params or {},
                "session_id": self._session_id,
            }
            
            # Send request
            self._send_request(request)
            
            # Receive response
            response = self._receive_response()
            
            # Handle errors
            if response.get("error"):
                raise ConnectionError(f"Query error: {response['error']}")
            
            return response
            
        except socket.timeout:
            raise ConnectionError("Query timeout")
        except socket.error as e:
            raise ConnectionError(f"Connection error during query: {e}")
    
    def _authenticate(self) -> None:
        """Perform authentication handshake."""
        auth_request = {
            "type": "auth",
            "username": self.username,
            "password": self.password,
            "database": self.database,
            "client_version": "0.1.0",
        }
        
        self._send_request(auth_request)
        response = self._receive_response()
        
        if response.get("error"):
            raise ConnectionError(f"Authentication failed: {response['error']}")
        
        self._session_id = response.get("session_id")
        if not self._session_id:
            raise ConnectionError("No session ID received from server")
    
    def _send_request(self, request: Dict[str, Any]) -> None:
        """Send a request to the server."""
        if not self._socket:
            raise ConnectionError("No active connection")
        
        try:
            # Convert request to JSON and send length-prefixed
            data = json.dumps(request).encode('utf-8')
            length = len(data).to_bytes(4, byteorder='big')
            
            self._socket.sendall(length + data)
            
        except socket.error as e:
            raise ConnectionError(f"Failed to send request: {e}")
    
    def _receive_response(self) -> Dict[str, Any]:
        """Receive a response from the server."""
        if not self._socket:
            raise ConnectionError("No active connection")
        
        try:
            # Read length prefix
            length_bytes = self._recv_exact(4)
            length = int.from_bytes(length_bytes, byteorder='big')
            
            # Read response data
            data = self._recv_exact(length)
            response = json.loads(data.decode('utf-8'))
            
            return response
            
        except socket.error as e:
            raise ConnectionError(f"Failed to receive response: {e}")
        except json.JSONDecodeError as e:
            raise ConnectionError(f"Invalid JSON response: {e}")
    
    def _recv_exact(self, length: int) -> bytes:
        """Receive exactly the specified number of bytes."""
        if not self._socket:
            raise ConnectionError("No active connection")
            
        data = b''
        while len(data) < length:
            chunk = self._socket.recv(length - len(data))
            if not chunk:
                raise ConnectionError("Connection closed by server")
            data += chunk
        return data
    
    def ping(self) -> bool:
        """Check if server is responsive."""
        try:
            response = self.execute("SELECT 1")
            return response.get("success", False)
        except:
            return False
    
    def get_server_info(self) -> Dict[str, Any]:
        """Get server information."""
        try:
            response = self.execute("SELECT version()")
            return response.get("server_info", {})
        except:
            return {}
    
    def __enter__(self):
        """Context manager entry."""
        self.connect()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.disconnect()


@contextmanager
def connect(
    host: str = "localhost",
    port: int = 5432,
    database: Optional[str] = None,
    username: Optional[str] = None,
    password: Optional[str] = None,
    timeout: int = 30,
    use_ssl: bool = False,
):
    """
    Context manager for creating and managing connections.
    
    Args:
        host: Database server host
        port: Database server port
        database: Database name
        username: Username for authentication
        password: Password for authentication
        timeout: Connection timeout in seconds
        use_ssl: Whether to use SSL connection
        
    Yields:
        Connection: Active connection to the server
    """
    conn = Connection(
        host=host,
        port=port,
        database=database,
        username=username,
        password=password,
        timeout=timeout,
        use_ssl=use_ssl,
    )
    
    try:
        conn.connect()
        yield conn
    finally:
        conn.disconnect()