# TelumDB Development Guidelines

## Build/Test Commands

### Go
- `make build` - Build main binary and CLI
- `make test` - Run all Go tests
- `make test ./pkg/storage/...` - Run tests for specific package
- `go test -v ./pkg/storage/engine_test.go` - Run single test file
- `make test-coverage` - Run tests with coverage report
- `make fmt` - Format Go code
- `make lint` - Run golangci-lint
- `make vet` - Run go vet
- `make quality` - Run fmt, lint, vet, test

### Python
- `cd api/python && python -m pytest tests/` - Run Python tests
- `cd api/python && python -m pytest tests/test_client.py` - Run single test
- `cd api/python && python setup.py build_ext --inplace` - Build Python bindings

## Code Style Guidelines

### Go
- Use `gofmt` for formatting (enforced by CI)
- Exported functions must have godoc comments
- Use meaningful variable names, avoid abbreviations
- Keep functions small and focused (<50 lines when possible)
- Handle errors explicitly with `if err != nil`
- Use table-driven tests for multiple test cases
- Import organization: stdlib, third-party, internal (separated by blank lines)

### Python
- Follow PEP 8 style (enforced by black/flake8)
- Use type hints for all function signatures
- Docstrings for all public functions/classes
- Import organization: stdlib, third-party, local (separated by blank lines)
- Use context managers for resource management
- Exception handling should be specific and meaningful

### Error Handling
- Go: Return errors as last return value, use fmt.Errorf with %w for wrapping
- Python: Use custom exceptions from telumdb.exceptions, include context in error messages

### Naming Conventions
- Go: PascalCase for exported, camelCase for unexported
- Python: snake_case for variables/functions, PascalCase for classes
- Use descriptive names that indicate purpose

### Testing
- Aim for >80% code coverage
- Unit tests in same package as code
- Integration tests in test/integration/
- Use table-driven tests in Go for multiple scenarios