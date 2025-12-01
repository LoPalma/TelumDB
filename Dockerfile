# Multi-stage build for TelumDB
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o telumdb ./cmd/telumdb

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S telumdb && \
    adduser -u 1001 -S telumdb -G telumdb

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/telumdb /app/telumdb

# Copy configuration files
COPY --chown=telumdb:telumdb config/default.yaml /app/config.yaml

# Create data directory
RUN mkdir -p /app/data && chown telumdb:telumdb /app/data

# Switch to non-root user
USER telumdb

# Expose ports
EXPOSE 5432 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD telumdb health || exit 1

# Set environment variables
ENV TELUMDB_CONFIG_FILE=/app/config.yaml
ENV TELUMDB_DATA_DIR=/app/data

# Run the binary
CMD ["./telumdb"]