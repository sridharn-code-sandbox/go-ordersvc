# Build stage
FROM --platform=linux/arm64 golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimizations
ARG VERSION=dev
RUN CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o ordersvc \
    ./cmd/ordersvc

# Runtime stage
FROM --platform=linux/arm64 alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Copy binary from builder
COPY --from=builder /build/ordersvc /usr/local/bin/ordersvc

# Set ownership
RUN chown appuser:appuser /usr/local/bin/ordersvc

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080 9090

# Run application
ENTRYPOINT ["/usr/local/bin/ordersvc"]
