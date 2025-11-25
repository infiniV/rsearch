# Stage 1: Build stage
FROM golang:alpine AS builder

# Build arguments for version information
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# Build the binary with optimizations
# CGO_ENABLED=0 for static binary
# -ldflags="-s -w" strips debug info for smaller size
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}" \
    -o rsearch \
    ./cmd/rsearch/main.go

# Stage 2: Runtime stage
FROM alpine:3.19

# Build arguments for labels
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Add labels
LABEL maintainer="rsearch" \
      version="${VERSION}" \
      description="Production-grade query translation service for OpenSearch/Elasticsearch to SQL" \
      org.opencontainers.image.source="https://github.com/infiniv/rsearch" \
      org.opencontainers.image.title="rsearch" \
      org.opencontainers.image.description="Query translation service" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}"

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1000 rsearch && \
    adduser -D -u 1000 -G rsearch rsearch

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/rsearch /app/rsearch

# Copy example config for reference
COPY config.example.yaml /app/config.example.yaml

# Create directory for schemas if needed
RUN mkdir -p /app/schemas && chown -R rsearch:rsearch /app

# Switch to non-root user
USER rsearch

# Expose application port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
ENTRYPOINT ["/app/rsearch"]
