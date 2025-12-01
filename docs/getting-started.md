# Getting Started with TelumDB

Welcome to TelumDB - The World's First Hybrid General-Purpose + AI Tensor Database!

This guide will help you get up and running with TelumDB quickly.

## Installation

### Binary Installation

Download the latest binary for your platform:

```bash
# Linux (AMD64)
wget https://github.com/telumdb/telumdb/releases/latest/download/telumdb-linux-amd64
chmod +x telumdb-linux-amd64
sudo mv telumdb-linux-amd64 /usr/local/bin/telumdb

# macOS (AMD64)
wget https://github.com/telumdb/telumdb/releases/latest/download/telumdb-darwin-amd64
chmod +x telumdb-darwin-amd64
sudo mv telumdb-darwin-amd64 /usr/local/bin/telumdb

# Windows (AMD64)
# Download telumdb-windows-amd64.exe and add to PATH
```

### Docker Installation

```bash
# Pull the latest image
docker pull telumdb/telumdb:latest

# Run the container
docker run -d \
  --name telumdb \
  -p 5432:5432 \
  -p 8080:8080 \
  -v telumdb_data:/app/data \
  telumdb/telumdb:latest
```

### From Source

```bash
# Clone the repository
git clone https://github.com/telumdb/telumdb.git
cd telumdb

# Build the binary
make build

# Install to system (optional)
make install
```

## Quick Start

### 1. Start the Server

```bash
# Using default configuration
telumdb

# Using custom configuration
telumdb -config /path/to/config.yaml

# Check version
telumdb -version
```

The server will start on:
- Database protocol: `localhost:5432`
- HTTP API: `localhost:8080`
- Metrics: `localhost:9000`

### 2. Connect with CLI

```bash
# Interactive mode
telumdb-cli

# Execute single command
telumdb-cli -c "SHOW TABLES"

# Connect to remote server
telumdb-cli -url telumdb://remote-host:5432
```

### 3. Connect with Python

```python
import telumdb
import numpy as np

# Connect to database
db = telumdb.connect(host="localhost", port=5432)

# Create a traditional table
db.create_table("users", {
    "id": "INTEGER PRIMARY KEY",
    "name": "VARCHAR(255)",
    "email": "VARCHAR(255)",
    "age": "INTEGER"
})

# Insert data
db.execute("""
    INSERT INTO users (id, name, email, age) VALUES
    (1, 'Alice', 'alice@example.com', 30),
    (2, 'Bob', 'bob@example.com', 25)
""")

# Create a tensor for embeddings
embeddings = db.create_tensor(
    name="user_embeddings",
    shape=[1000, 768],  # 1000 users, 768-dimensional embeddings
    dtype="float32"
)

# Store embedding data
sample_embeddings = np.random.rand(10, 768).astype(np.float32)
embeddings.store_chunk([0, 0], sample_embeddings)

# Mixed query: Find users with similar embeddings
result = db.execute("""
    SELECT u.name, u.email, 
           cosine_similarity(e.embeddings, [0.1, 0.2, 0.3, ...]) as similarity
    FROM users u
    JOIN user_embeddings e ON u.id = e.user_id
    WHERE u.age > 25
      AND cosine_similarity(e.embeddings, [0.1, 0.2, 0.3, ...]) > 0.8
    ORDER BY similarity DESC
""")

for row in result.rows:
    print(f"User: {row[0]}, Email: {row[1]}, Similarity: {row[2]}")
```

## Basic Concepts

### Tables vs Tensors

TelumDB stores two types of data:

**Tables**: Traditional relational data with rows and columns
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255)
);
```

**Tensors**: Multi-dimensional arrays optimized for AI/ML workloads
```sql
CREATE TENSOR embeddings (
    shape [1000, 768],
    dtype float32,
    chunk_size [100, 768]
);
```

### Mixed Queries

Combine traditional SQL with tensor operations:

```sql
-- Find similar users based on embeddings
SELECT u.name, u.age,
       cosine_similarity(e.embeddings, query_vector) as similarity
FROM users u
JOIN user_embeddings e ON u.id = e.user_id
WHERE cosine_similarity(e.embeddings, query_vector) > 0.8
ORDER BY similarity DESC;
```

## Configuration

### Basic Configuration

Create a `config.yaml` file:

```yaml
server:
  host: "0.0.0.0"
  port: 5432
  http_port: 8080
  max_connections: 1000

storage:
  data_dir: "./data"
  engine: "hybrid"
  cache_size: 1073741824  # 1GB
  tensor:
    chunk_size: [64, 64, 64]
    default_dtype: "float32"
    memory_limit: 4294967296  # 4GB

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  port: 9000
```

### Environment Variables

```bash
export TELUMDB_CONFIG_FILE=/path/to/config.yaml
export TELUMDB_DATA_DIR=/path/to/data
export TELUMDB_LOG_LEVEL=debug
```

## Next Steps

- [API Reference](api-reference.md) - Detailed API documentation
- [Architecture Overview](architecture.md) - Understanding TelumDB internals
- [Performance Tuning](performance.md) - Optimize your TelumDB deployment
- [Deployment Guide](deployment.md) - Production deployment strategies

## Examples

Check out the [examples](../examples/) directory for more comprehensive examples:

- [Basic Usage](../examples/basic_usage.py) - Core operations
- [ML Pipeline](../examples/ml_pipeline.py) - End-to-end ML workflow
- [Time Series](../examples/time_series.py) - Time series data analysis
- [Image Processing](../examples/image_processing.py) - Image tensor operations

## Getting Help

- **Documentation**: https://telumdb.readthedocs.io/
- **GitHub Issues**: https://github.com/telumdb/telumdb/issues
- **Discussions**: https://github.com/telumdb/telumdb/discussions
- **Email**: support@telumdb.io

## Community

- **Discord**: [Join our Discord server](https://discord.gg/telumdb)
- **Twitter**: [@TelumDB](https://twitter.com/telumdb)
- **Blog**: https://blog.telumdb.io/

Happy coding with TelumDB! 🚀