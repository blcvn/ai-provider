# AI Proxy Service

Unified gateway for all LLM requests with intelligent routing, fallback, caching, and monitoring.

## Features

- **Multi-Provider Support**: Anthropic Claude, OpenAI GPT, Ollama (local LLMs)
- **LangChainGo Integration**: Standardized LLM interactions using `github.com/tmc/langchaingo`
- **Circuit Breaker**: Automatic failover with `sony/gobreaker`
- **Redis Caching**: SHA256-based caching for deterministic requests (1h TTL)
- **Load Balancing**: Round-robin API key distribution
- **Prometheus Metrics**: Comprehensive observability
- **Cost Tracking**: Automatic cost calculation per model

## Architecture

```
Client → AI Proxy Service → [Cache] → [Router] → [Circuit Breaker] → LLM Provider
                                                                      ├─ Anthropic Claude
                                                                      ├─ OpenAI GPT
                                                                      └─ Ollama (Local)
```

## Quick Start

### Prerequisites

- Go 1.23+
- Redis
- AI Model Service running on `:8085`

### Environment Variables

```bash
cp .env.example .env
# Edit .env with your configuration
```

### Run Locally

```bash
go run cmd/server.go
```

### Build

```bash
go build -o ai-proxy-service ./cmd/server.go
./ai-proxy-service
```

### Docker

```bash
docker build -t ai-proxy-service .
docker run -p 8087:8087 -p 9090:9090 --env-file .env ai-proxy-service
```

## API

### gRPC Endpoints

- `Complete(CompleteRequest) → CompleteResponse` - Synchronous completion
- `StreamComplete(CompleteRequest) → stream StreamCompleteResponse` - Streaming (future)
- `HealthCheck(HealthCheckRequest) → HealthCheckResponse` - Health status
- `GetProviderStatus(GetProviderStatusRequest) → GetProviderStatusResponse` - Circuit breaker states

### HTTP Endpoints (via gRPC-Gateway)

- `POST /v1/complete` - Completion request
- `GET /v1/health` - Health check
- `GET /v1/providers/status` - Provider status

## Metrics

Prometheus metrics available at `:9090/metrics`:

- `ai_proxy_requests_total` - Total requests by model/provider/status
- `ai_proxy_request_duration_seconds` - Request latency histogram
- `ai_proxy_tokens_used_total` - Token consumption by type
- `ai_proxy_cost_total` - Cumulative cost in USD
- `ai_proxy_cache_hits_total` - Cache hit/miss counts
- `ai_proxy_circuit_breaker_state` - Circuit breaker states

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8087` | gRPC server port |
| `METRICS_PORT` | `9090` | Prometheus metrics port |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `AI_MODEL_SERVICE_ADDR` | `localhost:8085` | AI Model Service address |
| `CIRCUIT_BREAKER_MAX_REQUESTS` | `5` | Failures before opening |
| `CIRCUIT_BREAKER_TIMEOUT` | `60` | Seconds to stay open |
| `CACHE_TTL` | `3600` | Cache TTL in seconds |

## Provider Adapters

### Anthropic Claude

```go
provider, _ := anthropic.NewClaudeProvider(apiKey, "claude-3-5-sonnet-20241022")
response, _ := provider.Complete(ctx, &CompletionRequest{
    Prompt: "Hello",
    MaxTokens: 100,
})
```

### OpenAI GPT

```go
provider, _ := openai.NewGPTProvider(apiKey, "gpt-4")
response, _ := provider.Complete(ctx, &CompletionRequest{
    Prompt: "Hello",
    MaxTokens: 100,
})
```

### Ollama (Local LLM)

```go
provider, _ := local.NewOllamaProvider("http://localhost:11434", "llama2")
response, _ := provider.Complete(ctx, &CompletionRequest{
    Prompt: "Hello",
    MaxTokens: 100,
})
```

## Circuit Breaker

- **Closed**: Normal operation
- **Open**: Provider failing, requests fail fast (60s timeout)
- **Half-Open**: Testing recovery (1 request allowed)

Configuration:
- Opens after 5 consecutive failures
- Stays open for 60 seconds
- Allows 1 request in half-open state

## Caching Strategy

- **Deterministic requests** (temperature=0) are cached
- **Cache key**: SHA256(model_id + prompt + temperature + max_tokens + top_p)
- **TTL**: 1 hour
- **Storage**: Redis

## Cost Calculation

| Provider | Input (per 1M tokens) | Output (per 1M tokens) |
|----------|----------------------|------------------------|
| Claude 3.5 Sonnet | $3.00 | $15.00 |
| GPT-4 | $10.00 | $30.00 |
| Ollama (Local) | $0.00 | $0.00 |

## Development

### Project Structure

```
ai-proxy-service/
├── cmd/
│   └── server.go          # Main entrypoint
├── providers/
│   ├── interface.go       # LLMProvider interface
│   ├── anthropic/
│   │   └── claude.go      # Claude adapter
│   ├── openai/
│   │   └── gpt.go         # GPT adapter
│   └── local/
│       └── ollama.go      # Ollama adapter
├── cache/
│   └── redis_cache.go     # Redis caching
├── resilience/
│   └── circuit_breaker.go # Circuit breaker
├── router/
│   └── router.go          # Provider routing
├── usecases/
│   └── proxy_usecase.go   # Business logic
├── controllers/
│   └── proxy_controller.go # gRPC handlers
├── metrics/
│   └── metrics.go         # Prometheus metrics
├── config/
│   └── config.go          # Configuration
└── Dockerfile             # Docker build
```

### Testing

```bash
# Unit tests
go test ./...

# Integration tests (requires Redis and AI Model Service)
go test -tags=integration ./...
```

## License

MIT
