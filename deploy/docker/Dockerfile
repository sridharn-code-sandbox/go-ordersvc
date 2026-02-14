# Build stage
FROM --platform=linux/arm64 golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary with optimizations
ARG VERSION=dev
RUN CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o ordersvc \
    ./cmd/ordersvc

# Runtime stage - using distroless for minimal attack surface
FROM --platform=linux/arm64 gcr.io/distroless/static-debian12:nonroot

# Copy binary from builder
COPY --from=builder /build/ordersvc /usr/local/bin/ordersvc

# Expose ports (HTTP and gRPC)
EXPOSE 8080 9090

# Run as nonroot user (distroless default)
USER nonroot:nonroot

# Run application
ENTRYPOINT ["/usr/local/bin/ordersvc"]
