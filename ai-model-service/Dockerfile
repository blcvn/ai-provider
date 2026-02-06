# Build stage
FROM golang:1.24 as builder
WORKDIR /app

# Copy go.mod files for dependency resolution
COPY services/pkg/go.mod ./services/pkg/
COPY services/ai-model-service/go.mod services/ai-model-service/go.sum ./services/ai-model-service/

# Download dependencies
RUN cd services/ai-model-service && go mod download

# Copy source code
COPY . .

# Build the application
RUN cd services/ai-model-service && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /bin/ai-model-service ./main.go

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
COPY --from=builder /bin/ai-model-service /bin/ai-model-service

# Copy config file from builder (Optional, if exists in source)
# Verifying path: services/ai-model-service/configs/config.yaml or ./config/config.yaml?
# User's previous docker-compose mapped: ../../services/ai-model-service/configs/config.yaml
# We'll assume the standard location or verify it. 
# Copying from source if it exists helps defaults. 
# Safe approach: COPY if it exists pattern, or just rely on bind mount in docker-compose.
# auth-service Dockerfile copies it. I will try to copy expected path.
COPY --from=builder /app/services/ai-model-service/configs/config.yaml /app/config/config.yaml

# Set ownership
RUN chown appuser:appuser /bin/ai-model-service && \
    chown -R appuser:appuser /app/config

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /app

# Environment variables with defaults (NATIVE PORTS)
ENV SERVICE_NAME=ai-model-service \
    JAEGER_URL=localhost:4317 \
    METRICS_PATH=/metrics \
    GRPC_PORT=9085 \
    HTTP_PORT=8085 \
    CONFIG_FILE=/app/config/config.yaml \
    SECRET_FILE=/vault/secrets/config.json

# Expose ports
EXPOSE 9085 8085

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${HTTP_PORT}/metrics || exit 1

# Run the application
CMD ["/bin/ai-model-service", "serve", \
    "--service-name=${SERVICE_NAME}", \
    "--jaeger-url=${JAEGER_URL}", \
    "--metrics-path=${METRICS_PATH}", \
    "--grpc-port=${GRPC_PORT}", \
    "--http-port=${HTTP_PORT}"]
