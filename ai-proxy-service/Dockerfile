# Build stage
FROM golang:1.24 as builder
WORKDIR /app

# Copy go.mod files for dependency resolution
COPY services/pkg/go.mod ./services/pkg/
COPY services/ai-proxy-service/go.mod services/ai-proxy-service/go.sum ./services/ai-proxy-service/

# Download dependencies
RUN cd services/ai-proxy-service && go mod download

# Copy source code
COPY . .

# Build the application
RUN cd services/ai-proxy-service && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /bin/ai-proxy-service ./main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates curl tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Create necessary directories
RUN mkdir -p /app/config /vault/secrets && \
    chown -R appuser:appuser /app

# Copy binary from builder
COPY --from=builder /bin/ai-proxy-service /bin/ai-proxy-service

# Copy config file from builder
COPY --from=builder /app/services/ai-proxy-service/config/config.yaml /app/config/config.yaml

# Set ownership
RUN chown appuser:appuser /bin/ai-proxy-service && \
    chown -R appuser:appuser /app/config

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /app

# Environment variables with defaults (NATIVE PORTS)
ENV SERVICE_NAME=ai-proxy-service \
    JAEGER_URL=localhost:4317 \
    METRICS_PATH=/metrics \
    GRPC_PORT=9087 \
    HTTP_PORT=8087 \
    CONFIG_FILE=/app/config/config.yaml \
    SECRET_FILE=/vault/secrets/config.json

# Expose ports
EXPOSE 9087 8087

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${HTTP_PORT}/metrics || exit 1

# Run the application
CMD ["/bin/ai-proxy-service", "serve", \
    "--service-name=${SERVICE_NAME}", \
    "--jaeger-url=${JAEGER_URL}", \
    "--metrics-path=${METRICS_PATH}", \
    "--grpc-port=${GRPC_PORT}", \
    "--http-port=${HTTP_PORT}"]
