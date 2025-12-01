# TelumDB Quick Start

## 🚀 5-Minute Setup

### 1. Build from Source
```bash
git clone https://github.com/telumdb/telumdb.git
cd telumdb
make build
```

### 2. Start Server
```bash
./build/telumdb
```

### 3. Connect with CLI
```bash
./build/telumdb-cli
telumdb> CREATE TABLE users (id INTEGER, name VARCHAR(255));
telumdb> CREATE TENSOR embeddings (shape [1000, 768], dtype float32);
telumdb> \q
```

### 4. Use Python API
```python
import telumdb
db = telumdb.connect()
db.create_table("users", {"id": "INTEGER", "name": "VARCHAR(255)"})
tensor = db.create_tensor("embeddings", shape=[1000, 768])
```

## 🐳 Docker Quick Start
```bash
docker run -p 5432:5432 telumdb/telumdb:latest
```

## 📚 What Makes TelumDB Special?

✅ **Hybrid Database**: Traditional SQL + Native Tensor Operations  
✅ **AI-Optimized**: Chunked storage perfect for ML workloads  
✅ **Multi-Language**: Python, C, Java, and SQL support  
✅ **Production Ready**: ACID transactions, distributed, secure  

## 🎯 Key Features

```sql
-- Traditional SQL
CREATE TABLE users (id INTEGER, name VARCHAR(255));

-- Native Tensors
CREATE TENSOR embeddings (shape [1000, 768], dtype float32);

-- Mixed Queries
SELECT u.name, cosine_similarity(e.embeddings, query_vector) as similarity
FROM users u JOIN embeddings e ON u.id = e.user_id
WHERE similarity > 0.8;
```

## 🔗 Links

- 📖 [Full Documentation](docs/getting-started.md)
- 🐙 [GitHub Repository](https://github.com/telumdb/telumdb)
- 💬 [Discussions](https://github.com/telumdb/telumdb/discussions)
- 🐛 [Issues](https://github.com/telumdb/telumdb/issues)

## 🌟 Star Us!

If you find TelumDB interesting, please give us a star on GitHub!

[![GitHub stars](https://img.shields.io/github/stars/telumdb/telumdb.svg?style=social&label=Star)](https://github.com/telumdb/telumdb)