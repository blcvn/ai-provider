FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ai-model-service .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/ai-model-service .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Expose ports
EXPOSE 8085 9085

# Run the service
CMD ["./ai-model-service", "serve"]
