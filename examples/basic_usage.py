#!/usr/bin/env python3
"""
TelumDB Basic Usage Example

This example demonstrates basic TelumDB operations including:
- Connecting to the database
- Creating tables and tensors
- Inserting and querying data
- Mixed SQL + tensor operations
"""

import sys
import os

# Add the Python API to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'api', 'python', 'src'))

try:
    import telumdb
    import numpy as np
except ImportError as e:
    print(f"Import error: {e}")
    print("Please ensure TelumDB is installed and running")
    sys.exit(1)


def main():
    print("TelumDB Basic Usage Example")
    print("=" * 40)
    
    # Connect to TelumDB
    print("\n1. Connecting to TelumDB...")
    try:
        with telumdb.connect(host="localhost", port=5432) as db:
            print("✓ Connected successfully")
            
            # Create a traditional table
            print("\n2. Creating a users table...")
            db.create_table("users", {
                "id": "INTEGER PRIMARY KEY",
                "name": "VARCHAR(255)",
                "email": "VARCHAR(255)",
                "age": "INTEGER"
            })
            print("✓ Table created")
            
            # Insert some data
            print("\n3. Inserting user data...")
            db.execute("""
                INSERT INTO users (id, name, email, age) VALUES
                (1, 'Alice', 'alice@example.com', 30),
                (2, 'Bob', 'bob@example.com', 25),
                (3, 'Charlie', 'charlie@example.com', 35)
            """)
            print("✓ Data inserted")
            
            # Query the table
            print("\n4. Querying user data...")
            result = db.execute("SELECT * FROM users WHERE age > 25")
            print("Results:")
            for row in result.rows:
                print(f"  {row}")
            
            # Create a tensor for embeddings
            print("\n5. Creating embeddings tensor...")
            embeddings = db.create_tensor(
                name="user_embeddings",
                shape=[1000, 768],  # 1000 users, 768-dimensional embeddings
                dtype="float32",
                chunk_size=[100, 768],
                metadata={"description": "User embedding vectors"}
            )
            print(f"✓ Tensor created: {embeddings}")
            
            # Store some embedding data
            print("\n6. Storing embedding data...")
            if telumdb._NUMPY_AVAILABLE:
                # Generate random embeddings
                sample_embeddings = np.random.rand(10, 768).astype(np.float32)
                embeddings.store_chunk([0, 0], sample_embeddings)
                print("✓ Embedding chunk stored")
            else:
                print("⚠ NumPy not available, skipping embedding storage")
            
            # Mixed query: Find users with similar embeddings
            print("\n7. Mixed SQL + Tensor query...")
            query = """
                SELECT u.name, u.email, 
                       cosine_similarity(e.embeddings, [0.1, 0.2, 0.3, ...]) as similarity
                FROM users u
                JOIN user_embeddings e ON u.id = e.user_id
                WHERE u.age > 25
                  AND cosine_similarity(e.embeddings, [0.1, 0.2, 0.3, ...]) > 0.8
                ORDER BY similarity DESC
                LIMIT 5
            """
            
            # This would work when the full implementation is ready
            print("Query (when fully implemented):")
            print(query)
            
            # List all objects
            print("\n8. Listing database objects...")
            tables = db.list_tables()
            tensors = db.list_tensors()
            
            print(f"Tables: {tables}")
            print(f"Tensors: {tensors}")
            
            print("\n✓ Example completed successfully!")
            
    except Exception as e:
        print(f"✗ Error: {e}")
        return 1
    
    return 0


def demonstrate_tensor_operations():
    """Demonstrate advanced tensor operations."""
    print("\nTensor Operations Demo")
    print("-" * 30)
    
    if not telumdb._NUMPY_AVAILABLE:
        print("NumPy not available, skipping tensor operations demo")
        return
    
    # Create a tensor
    tensor = telumdb.Tensor(
        name="demo_tensor",
        shape=[100, 200, 3],
        dtype="float32",
        chunk_size=[50, 100, 3]
    )
    
    print(f"Created tensor: {tensor}")
    
    # Create some sample data
    data = np.random.rand(10, 20, 3).astype(np.float32)
    tensor.store_chunk([0, 0, 0], data)
    print("✓ Stored chunk")
    
    # Slice operations
    sliced = tensor.slice([(0, 50), (0, 100), (0, 3)])
    print(f"✓ Sliced tensor: {sliced}")
    
    # Reshape operations
    reshaped = tensor.reshape([100 * 200 * 3])
    print(f"✓ Reshaped tensor: {reshaped}")
    
    # Mathematical operations
    added = tensor + 1.0
    print(f"✓ Added 1.0: {added}")
    
    multiplied = tensor * 2.0
    print(f"✓ Multiplied by 2.0: {multiplied}")


if __name__ == "__main__":
    exit_code = main()
    
    # Run tensor operations demo if main succeeded
    if exit_code == 0:
        try:
            demonstrate_tensor_operations()
        except Exception as e:
            print(f"Tensor demo error: {e}")
    
    sys.exit(exit_code)