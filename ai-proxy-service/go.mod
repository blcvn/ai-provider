module github.com/blcvn/backend/services/ai-proxy-service

go 1.24.4

require (
	github.com/blcvn/kratos-proto/go/ai-model v0.0.0
	github.com/blcvn/kratos-proto/go/ai-proxy v0.0.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/prometheus/client_golang v1.20.5
	github.com/sony/gobreaker v0.5.0
	github.com/tmc/langchaingo v0.1.14
	google.golang.org/grpc v1.78.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.7 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkoukk/tiktoken-go v0.1.6 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260128011058-8636f8732409 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260128011058-8636f8732409 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/blcvn/kratos-proto/go/ai-model => ../../../protos/go/ai-model
	github.com/blcvn/kratos-proto/go/ai-proxy => ../../../protos/go/ai-proxy
)
