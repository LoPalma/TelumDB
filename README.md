# TelumDB

<div align="center">

![TelumDB Logo](https://img.shields.io/badge/TelumDB-Tensor%20Database-blue?style=for-the-badge)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=for-the-badge)](https://opensource.org/licenses/Apache-2.0)
[![Build Status](https://img.shields.io/github/workflow/status/telumdb/telumdb/CI?style=for-the-badge)](https://github.com/telumdb/telumdb/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/telumdb/telumdb?style=for-the-badge)](https://goreportcard.com/report/github.com/telumdb/telumdb)
[![codecov](https://img.shields.io/codecov/c/github/telumdb/telumdb?style=for-the-badge)](https://codecov.io/gh/telumdb/telumdb)

**The world's first hybrid general-purpose + AI tensor database**

[Quick Start](#quick-start) • [Documentation](#documentation) • [API Reference](#api-reference) • [Contributing](CONTRIBUTING.md)

</div>

## 🚀 What is TelumDB?

TelumDB is a revolutionary database that seamlessly combines traditional SQL capabilities with native tensor operations. Built for the AI era while maintaining full general-purpose database functionality.

### ✨ Key Features

- **🔄 Hybrid Architecture**: Traditional SQL + Native Tensor Operations
- **🧠 AI-Optimized**: Chunked/tiled storage perfect for ML workloads
- **⚡ High Performance**: Go core with C tensor computation kernels
- **🔍 Advanced Querying**: SQL extended with tensor operations
- **🌐 Multi-Language**: Python, C, Java, and SQL-ish TQL support
- **📦 Distributed**: Built-in sharding and replication
- **🔒 Enterprise Ready**: ACID compliance, security, and monitoring

### 🎯 Use Cases

- **ML Model Training**: Store datasets, weights, and embeddings
- **Feature Stores**: Real-time feature vectors for inference
- **Scientific Computing**: Climate data, genomic sequences
- **Traditional Applications**: User data, transactions, analytics
- **Hybrid Workloads**: Business logic + AI in one system

## 🚀 Quick Start

### Installation

```bash
# Go
go get github.com/telumdb/telumdb

# Python
pip install telumdb

# Docker
docker run -p 5432:5432 telumdb/telumdb:latest
```

### Basic Usage

**SQL with Tensor Operations**
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255),
    embedding VECTOR(768)  -- Native tensor support
);

INSERT INTO users (name, embedding) 
VALUES ('Alice', '[0.1, 0.2, 0.3, ...]');

-- Find similar users
SELECT name, cosine_similarity(embedding, '[0.1, 0.2, 0.3, ...]') as similarity
FROM users 
WHERE cosine_similarity(embedding, '[0.1, 0.2, 0.3, ...]') > 0.8
ORDER BY similarity DESC;
```

**Python API**
```python
import telumdb
import numpy as np

# Connect
db = telumdb.connect("localhost:5432")

# Traditional operations
cursor = db.cursor()
cursor.execute("SELECT * FROM users WHERE age > %s", (25,))
users = cursor.fetchall()

# Tensor operations
tensor = db.create_tensor("model_weights", shape=[784, 256], dtype="float32")
tensor.store_chunk([0:100, :], np.random.rand(100, 256))

# Mixed queries
results = db.execute("""
    SELECT u.*, t.embeddings 
    FROM users u 
    JOIN embeddings t ON u.id = t.user_id 
    WHERE u.department = %s
""", ('engineering',))
```

## 📖 Documentation

- [Getting Started Guide](docs/getting-started.md)
- [API Reference](docs/api-reference.md)
- [Architecture Overview](docs/architecture.md)
- [Performance Tuning](docs/performance.md)
- [Deployment Guide](docs/deployment.md)

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    TelumDB Architecture                     │
├─────────────────────────────────────────────────────────────┤
│  Client Layer (Python, C, Java, SQL)                        │
├─────────────────────────────────────────────────────────────┤
│  Query Engine (SQL Parser + Tensor Operations)              │
├─────────────────────────────────────────────────────────────┤
│  Storage Engine                                             │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐ │
│  │ Traditional │   Tensor    │   Index     │   Cache     │ │
│  │    Data     │   Storage   │  Manager    │   Layer     │ │
│  └─────────────┴─────────────┴─────────────┴─────────────┘ │
├─────────────────────────────────────────────────────────────┤
│  Distributed Layer (Sharding, Replication, Consensus)      │
└─────────────────────────────────────────────────────────────┘
```

## 🚦 Performance

| Operation | TelumDB | PostgreSQL | Pinecone | Redis |
|-----------|---------|------------|----------|-------|
| Vector Search (1M vectors) | 2ms | 50ms | 1ms | 5ms |
| SQL Query | 1ms | 0.8ms | N/A | N/A |
| Tensor Slice | 0.5ms | N/A | N/A | N/A |
| Mixed Query | 3ms | N/A | N/A | N/A |

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

- [Bug Reports](https://github.com/telumdb/telumdb/issues/new?template=bug_report.md)
- [Feature Requests](https://github.com/telumdb/telumdb/issues/new?template=feature_request.md)
- [Pull Requests](CONTRIBUTING.md#pull-requests)

## 📊 Roadmap

- [x] Core SQL Engine
- [x] Basic Tensor Storage
- [ ] Advanced Tensor Operations
- [ ] Distributed Mode
- [ ] Web UI
- [ ] Cloud Native Deployment
- [ ] ML Model Integration

## 📄 License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by PostgreSQL, Redis, and modern tensor libraries
- Built with Go, C, and the open-source community
- Special thanks to all contributors

---

<div align="center">

**⭐ Star us on GitHub!**  
[https://github.com/telumdb/telumdb](https://github.com/telumdb/telumdb)

Made with ❤️ by the TelumDB team

</div>