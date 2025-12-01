# TelumDB C Library Test Makefile

CC = gcc
CFLAGS = -Wall -Wextra -std=c99 -g -DDEBUG
INCLUDES = -I../include
LDFLAGS = -L../build -ltelumdb

# Directories
SRC_DIR = .
BUILD_DIR = ../build
LIB_DIR = ../build

# Files
SOURCES = $(wildcard $(SRC_DIR)/*.c)
OBJECTS = $(SOURCES:$(SRC_DIR)/%.c=$(BUILD_DIR)/test_%.o)
TESTS = $(SOURCES:$(SRC_DIR)/%.c=$(BUILD_DIR)/test_%)

# Default target
all: $(TESTS)

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Compile test files
$(BUILD_DIR)/test_%: $(SRC_DIR)/%.c | $(BUILD_DIR)
	$(CC) $(CFLAGS) $(INCLUDES) -o $@ $< $(LDFLAGS)

# Run tests
test: $(TESTS)
	@echo "Running TelumDB C Library Tests..."
	@echo "=================================="
	@for test in $(TESTS); do \
		echo "Running $$test..."; \
		./$$test; \
		if [ $$? -eq 0 ]; then \
			echo "✓ $$test passed"; \
		else \
			echo "✗ $$test failed"; \
			exit 1; \
		fi; \
		echo ""; \
	done
	@echo "🎉 All tests passed!"

# Run specific test
test_connection: $(BUILD_DIR)/test_connection
	@echo "Running connection tests..."
	./$(BUILD_DIR)/test_connection

# Clean
clean:
	rm -rf $(BUILD_DIR)

# Help
help:
	@echo "TelumDB C Library Test Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all           - Build all tests"
	@echo "  test          - Run all tests"
	@echo "  test_connection - Run connection tests only"
	@echo "  clean         - Remove build artifacts"
	@echo "  help          - Show this help message"

.PHONY: all test test_connection clean help