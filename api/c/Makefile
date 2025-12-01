# TelumDB C Library Makefile

CC = gcc
CFLAGS = -Wall -Wextra -std=c99 -fPIC -O2
INCLUDES = -Iinclude
LDFLAGS = -shared

# Directories
SRC_DIR = src
INCLUDE_DIR = include
BUILD_DIR = build
TEST_DIR = test

# Files
SOURCES = $(wildcard $(SRC_DIR)/*.c)
OBJECTS = $(SOURCES:$(SRC_DIR)/%.c=$(BUILD_DIR)/%.o)
LIBRARY = $(BUILD_DIR)/libtelumdb.so
STATIC_LIB = $(BUILD_DIR)/libtelumdb.a

# Default target
all: $(LIBRARY) $(STATIC_LIB)

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Compile source files
$(BUILD_DIR)/%.o: $(SRC_DIR)/%.c | $(BUILD_DIR)
	$(CC) $(CFLAGS) $(INCLUDES) -c $< -o $@

# Create shared library
$(LIBRARY): $(OBJECTS)
	$(CC) $(LDFLAGS) -o $@ $^

# Create static library
$(STATIC_LIB): $(OBJECTS)
	ar rcs $@ $^

# Install library
install: $(LIBRARY) $(STATIC_LIB)
	sudo cp $(LIBRARY) /usr/local/lib/
	sudo cp $(STATIC_LIB) /usr/local/lib/
	sudo cp $(INCLUDE_DIR)/telumdb.h /usr/local/include/
	sudo ldconfig

# Uninstall library
uninstall:
	sudo rm -f /usr/local/lib/libtelumdb.so
	sudo rm -f /usr/local/lib/libtelumdb.a
	sudo rm -f /usr/local/include/telumdb.h

# Test
test: $(LIBRARY)
	$(MAKE) -C $(TEST_DIR)

# Clean
clean:
	rm -rf $(BUILD_DIR)
	$(MAKE) -C $(TEST_DIR) clean

# Debug build
debug: CFLAGS += -g -DDEBUG
debug: $(LIBRARY) $(STATIC_LIB)

# Help
help:
	@echo "TelumDB C Library Build System"
	@echo ""
	@echo "Targets:"
	@echo "  all       - Build shared and static libraries"
	@echo "  install   - Install library to system"
	@echo "  uninstall - Remove library from system"
	@echo "  test      - Build and run tests"
	@echo "  clean     - Remove build artifacts"
	@echo "  debug     - Build with debug symbols"
	@echo "  help      - Show this help message"

.PHONY: all install uninstall test clean debug help