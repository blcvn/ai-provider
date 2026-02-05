module github.com/blcvn/backend/services/ai-proxy-service

go 1.24.0

require (
	github.com/blcvn/kratos-proto/go/ai-model v0.0.0
	github.com/blcvn/kratos-proto/go/ai-proxy v0.0.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.7
	github.com/prometheus/client_golang v1.23.2
	github.com/sony/gobreaker v1.0.0
	github.com/spf13/cobra v1.10.2
	github.com/tmc/langchaingo v0.1.5
	google.golang.org/grpc v1.78.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.8.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkoukk/tiktoken-go v0.1.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260128011058-8636f8732409 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260128011058-8636f8732409 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/blcvn/kratos-proto/go/ai-proxy => /home/anhdt/vnpay/BA-Agentic/Agentic/protos/go/ai-proxy

replace github.com/blcvn/kratos-proto/go/ai-model => /home/anhdt/vnpay/BA-Agentic/Agentic/protos/go/ai-model
