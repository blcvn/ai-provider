# AI Model Service

AI Model Service manages AI model configurations, credentials (via Vault), and usage tracking with quota management.

## Features

- **Model Management**: CRUD operations for AI model configurations
- **Vault Integration**: Secure API key retrieval with 5-minute TTL caching
- **Quota Management**: Daily and monthly token usage limits
- **Usage Tracking**: Comprehensive logging for analytics and cost tracking
- **Clean Architecture**: Layered design (controllers, usecases, repository)

## Architecture

```
controllers/    # gRPC handlers (presentation layer)
usecases/       # Business logic (Vault, quota)
repository/     # Data access (PostgreSQL)
entities/       # Domain models
dto/            # Database models
helper/         # Utilities (Vault client, transformations)
```

## API Endpoints

### HTTP (via gRPC Gateway)
- `POST /ai/models` - Create model
- `GET /ai/models/{id}` - Get model
- `GET /ai/models` - List models
- `PUT /ai/models/{id}` - Update model
- `DELETE /ai/models/{id}` - Delete model
- `GET /ai/models/stats` - Usage statistics

### gRPC Only (Internal)
- `GetCredentials` - Retrieve API keys from Vault
- `LogUsage` - Log AI usage
- `CheckQuota` - Check quota limits

## Environment Variables

```bash
DATABASE_URL=postgresql://baagent:password@postgres:5432/baagent?sslmode=disable
VAULT_ADDR=http://vault:8200
VAULT_TOKEN=your-vault-token
GRPC_PORT=9085
HTTP_PORT=8085
```

## Database Migrations

```bash
migrate -path migrations -database "${DATABASE_URL}" up
```

## Running

### Development
```bash
go run main.go serve
```

### Production
```bash
go build -o ai-model-service
./ai-model-service serve
```

### Docker
```bash
docker build -t ai-model-service .
docker run -p 8085:8085 -p 9085:9085 \
  -e DATABASE_URL=... \
  -e VAULT_ADDR=... \
  -e VAULT_TOKEN=... \
  ai-model-service
```

## Vault Setup

Store API keys in Vault:

```bash
vault kv put secret/ai-services/openai \
  api_key="sk-..." \
  base_url="https://api.openai.com/v1"

vault kv put secret/ai-services/anthropic \
  api_key="sk-ant-..." \
  base_url="https://api.anthropic.com/v1"
```

## Example Usage

### Create Model
```bash
curl -X POST http://localhost:8085/ai/models \
  -H "Content-Type: application/json" \
  -d '{
    "payload": {
      "name": "gpt-4",
      "provider": "openai",
      "model_id": "gpt-4-turbo-preview",
      "base_url": "https://api.openai.com/v1",
      "vault_path": "secret/ai-services/openai",
      "quota_daily": 100000,
      "quota_monthly": 3000000,
      "cost_per_1k_tokens": 0.01
    }
  }'
```

### Get Credentials (Internal gRPC)
```go
resp, err := client.GetCredentials(ctx, &pb.GetCredentialsRequest{
    ModelId: "model-uuid",
})
// Returns: api_key, base_url, headers
```

### Check Quota
```go
resp, err := client.CheckQuota(ctx, &pb.CheckQuotaRequest{
    ModelId: "model-uuid",
})
// Returns: exceeded, daily_used, daily_limit, monthly_used, monthly_limit
```

## Testing

```bash
go test ./...
```

## Dependencies

- **gRPC/Protobuf**: Service communication
- **GORM**: Database ORM
- **Vault API**: Secrets management
- **TTL Cache**: Credentials caching
- **Decimal**: Precise cost calculations
