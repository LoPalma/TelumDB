# Contributing to TelumDB

Thank you for your interest in contributing to TelumDB! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Docker (optional, for development)
- Make

### Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/telumdb.git
   cd telumdb
   ```

2. **Install Dependencies**
   ```bash
   make deps
   ```

3. **Run Tests**
   ```bash
   make test
   ```

4. **Build Project**
   ```bash
   make build
   ```

## 📁 Project Structure

```
telumdb/
├── cmd/                    # Command-line applications
│   ├── telumdb/           # Main database server
│   └── telumdb-cli/       # Command-line client
├── pkg/                   # Core packages
│   ├── storage/          # Storage engine
│   ├── tensor/           # Tensor operations
│   ├── sql/              # SQL parser and executor
│   ├── api/              # Client APIs
│   └── distributed/      # Distributed components
├── internal/              # Internal packages
├── api/                   # API definitions
│   ├── python/           # Python bindings
│   ├── c/                # C library
│   └── java/             # Java module
├── docs/                  # Documentation
├── scripts/               # Build and deployment scripts
├── test/                  # Test files and benchmarks
└── examples/              # Example applications
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make bench

# Run integration tests
make test-integration
```

### Writing Tests

- Unit tests should be in the same package as the code they test
- Integration tests should be in `test/integration/`
- Use table-driven tests for multiple test cases
- Aim for >80% code coverage

Example test structure:
```go
func TestTensorStorage(t *testing.T) {
    tests := []struct {
        name     string
        shape    []int
        dtype    string
        wantErr  bool
    }{
        {
            name:    "valid tensor",
            shape:   []int{10, 20},
            dtype:   "float32",
            wantErr: false,
        },
        // more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## 📝 Code Style

### Go Code Style

We follow the standard Go conventions and use `gofmt` and `golint`:

```bash
make fmt
make lint
```

Key guidelines:
- Use `gofmt` for formatting
- Exported functions should have comments
- Use meaningful variable names
- Keep functions small and focused
- Handle errors properly

### C Code Style

For C code in tensor operations:
- Use 4 spaces for indentation
- Follow Linux kernel coding style
- Include proper error handling
- Document complex algorithms

## 🔀 Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Changes

- Write clean, well-documented code
- Add tests for new functionality
- Ensure all tests pass
- Update documentation if needed

### 3. Commit Changes

Use conventional commit messages:
```
feat: add tensor slicing operation
fix: resolve memory leak in tensor storage
docs: update API documentation
test: add integration tests for distributed mode
```

### 4. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Create a Pull Request with:
- Clear description of changes
- Link to relevant issues
- Test results
- Screenshots if applicable

## 🐛 Bug Reports

When reporting bugs, please include:

- **Environment**: OS, Go version, TelumDB version
- **Reproduction Steps**: Clear steps to reproduce the issue
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Error Messages**: Any error logs or stack traces
- **Additional Context**: Any other relevant information

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md).

## 💡 Feature Requests

For feature requests:

- **Problem Description**: What problem are you trying to solve?
- **Proposed Solution**: How do you envision the solution?
- **Alternatives Considered**: What other approaches did you consider?
- **Additional Context**: Any other relevant information

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md).

## 🏗️ Architecture Decisions

Major architectural changes should be documented in ADRs (Architecture Decision Records):

1. Create a new ADR in `docs/architecture/adr-XXX-description.md`
2. Follow the ADR template
3. Submit for review before implementation

## 📚 Documentation

### Types of Documentation

- **API Documentation**: Code comments and generated docs
- **User Documentation**: Guides, tutorials, examples
- **Developer Documentation**: Architecture, contributing guide
- **Architecture Documentation**: ADRs, design documents

### Writing Documentation

- Use clear, concise language
- Include code examples
- Add diagrams where helpful
- Keep documentation up to date

## 🚀 Release Process

Releases are managed using semantic versioning:

1. **Patch releases** (X.Y.Z): Bug fixes
2. **Minor releases** (X.Y+1.0): New features
3. **Major releases** (X+1.0.0): Breaking changes

Release checklist:
- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Version bumped
- [ ] Tag created
- [ ] Release notes written

## 🤝 Community

### Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please read our [Code of Conduct](CODE_OF_CONDUCT.md).

### Getting Help

- **GitHub Issues**: For bug reports and feature requests
- **Discussions**: For general questions and ideas
- **Discord/Slack**: For real-time conversation (coming soon)

### Recognition

Contributors are recognized in:
- README.md contributors section
- Release notes
- Annual contributor highlights

## 📋 Development Checklist

Before submitting a PR, ensure:

- [ ] Code follows style guidelines
- [ ] All tests pass
- [ ] New tests added for new functionality
- [ ] Documentation updated
- [ ] CHANGELOG updated (if applicable)
- [ ] No breaking changes without version bump
- [ ] Performance impact considered
- [ ] Security implications considered

## 🏆 Recognition System

We value all contributions! Contributors can earn:

- **🌟 Contributor**: First merged PR
- **💎 Regular Contributor**: 5+ merged PRs
- **🚀 Core Contributor**: 20+ merged PRs
- **👑 Maintainer**: Long-term commitment and expertise

## 📞 Contact

- **Maintainers**: @telumdb/core-team
- **Email**: maintainers@telumdb.io
- **Website**: https://telumdb.io

---

Thank you for contributing to TelumDB! Your contributions help make the project better for everyone. 🙏