"""
TelumDB Cursor Module

Provides cursor-like interface for executing queries and fetching results.
"""

from typing import List, Dict, Any, Optional, Iterator


class Cursor:
    """Database cursor for executing queries and fetching results."""
    
    def __init__(self, connection):
        """
        Initialize cursor.
        
        Args:
            connection: Active database connection
        """
        self._connection = connection
        self._last_result: Optional[Dict[str, Any]] = None
        self._row_index = 0
        self._closed = False
    
    def execute(self, query: str, params: Optional[Dict[str, Any]] = None) -> None:
        """
        Execute a query.
        
        Args:
            query: SQL/TQL query string
            params: Query parameters
        """
        if self._closed:
            raise Exception("Cursor is closed")
        
        try:
            self._last_result = self._connection.execute(query, params)
            self._row_index = 0
        except Exception as e:
            raise Exception(f"Failed to execute query: {e}")
    
    def executemany(self, query: str, params_list: List[Dict[str, Any]]) -> None:
        """
        Execute a query multiple times with different parameters.
        
        Args:
            query: SQL/TQL query string
            params_list: List of parameter dictionaries
        """
        if self._closed:
            raise Exception("Cursor is closed")
        
        for params in params_list:
            self.execute(query, params)
    
    def fetchone(self) -> Optional[List[Any]]:
        """
        Fetch one row from the result set.
        
        Returns:
            Single row as list, or None if no more rows
        """
        if not self._last_result:
            return None
        
        rows = self._last_result.get("rows", [])
        if self._row_index >= len(rows):
            return None
        
        row = rows[self._row_index]
        self._row_index += 1
        return row
    
    def fetchmany(self, size: int = 1) -> List[List[Any]]:
        """
        Fetch multiple rows from the result set.
        
        Args:
            size: Number of rows to fetch
            
        Returns:
            List of rows
        """
        if not self._last_result:
            return []
        
        rows = self._last_result.get("rows", [])
        result = []
        
        for _ in range(size):
            row = self.fetchone()
            if row is None:
                break
            result.append(row)
        
        return result
    
    def fetchall(self) -> List[List[Any]]:
        """
        Fetch all remaining rows from the result set.
        
        Returns:
            List of all remaining rows
        """
        if not self._last_result:
            return []
        
        rows = self._last_result.get("rows", [])
        if self._row_index >= len(rows):
            return []
        
        result = rows[self._row_index:]
        self._row_index = len(rows)
        return result
    
    def __iter__(self) -> Iterator[List[Any]]:
        """
        Iterate over result rows.
        
        Yields:
            Each row in the result set
        """
        while True:
            row = self.fetchone()
            if row is None:
                break
            yield row
    
    def rowcount(self) -> int:
        """
        Get the number of rows affected by the last query.
        
        Returns:
            Number of affected rows
        """
        if not self._last_result:
            return -1
        
        return self._last_result.get("affected_rows", -1)
    
    def description(self) -> Optional[List[Dict[str, Any]]]:
        """
        Get column information for the last query.
        
        Returns:
            List of column descriptions
        """
        if not self._last_result:
            return None
        
        columns = self._last_result.get("columns", [])
        return [{"name": col, "type": "unknown"} for col in columns]
    
    def close(self) -> None:
        """Close the cursor."""
        self._closed = True
        self._last_result = None
    
    def __enter__(self):
        """Context manager entry."""
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.close()


class DictCursor:
    """Cursor that returns rows as dictionaries."""
    
    def __init__(self, connection):
        """
        Initialize cursor.
        
        Args:
            connection: Active database connection
        """
        self._connection = connection
        self._last_result: Optional[Dict[str, Any]] = None
        self._row_index = 0
        self._closed = False
    
    def execute(self, query: str, params: Optional[Dict[str, Any]] = None) -> None:
        """
        Execute a query.
        
        Args:
            query: SQL/TQL query string
            params: Query parameters
        """
        if self._closed:
            raise Exception("Cursor is closed")
        
        try:
            self._last_result = self._connection.execute(query, params)
            self._row_index = 0
        except Exception as e:
            raise Exception(f"Failed to execute query: {e}")
    
    def fetchone(self) -> Optional[Dict[str, Any]]:
        """
        Fetch one row as a dictionary.
        
        Returns:
            Single row as dictionary, or None if no more rows
        """
        if not self._last_result:
            return None
        
        rows = self._last_result.get("rows", [])
        if self._row_index >= len(rows):
            return None
        
        row = rows[self._row_index]
        self._row_index += 1
        
        columns = self._last_result.get("columns", [])
        return dict(zip(columns, row))
    
    def fetchmany(self, size: int = 1) -> List[Dict[str, Any]]:
        """
        Fetch multiple rows as dictionaries.
        
        Args:
            size: Number of rows to fetch
            
        Returns:
            List of row dictionaries
        """
        result = []
        for _ in range(size):
            row = self.fetchone()
            if row is None:
                break
            result.append(row)
        return result
    
    def fetchall(self) -> List[Dict[str, Any]]:
        """
        Fetch all remaining rows as dictionaries.
        
        Returns:
            List of row dictionaries
        """
        result = []
        while True:
            row = self.fetchone()
            if row is None:
                break
            result.append(row)
        return result
    
    def __iter__(self) -> Iterator[Dict[str, Any]]:
        """
        Iterate over result rows as dictionaries.
        
        Yields:
            Each row as a dictionary
        """
        while True:
            row = self.fetchone()
            if row is None:
                break
            yield row
    
    def rowcount(self) -> int:
        """
        Get the number of rows affected by the last query.
        
        Returns:
            Number of affected rows
        """
        if not self._last_result:
            return -1
        
        return self._last_result.get("affected_rows", -1)
    
    def description(self) -> Optional[List[Dict[str, Any]]]:
        """
        Get column information for the last query.
        
        Returns:
            List of column descriptions
        """
        if not self._last_result:
            return None
        
        columns = self._last_result.get("columns", [])
        return [{"name": col, "type": "unknown"} for col in columns]
    
    def close(self) -> None:
        """Close the cursor."""
        self._closed = True
        self._last_result = None
    
    def __enter__(self):
        """Context manager entry."""
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.close()