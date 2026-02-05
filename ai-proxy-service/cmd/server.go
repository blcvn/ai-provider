package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/blcvn/backend/services/ai-proxy-service/cache"
	"github.com/blcvn/backend/services/ai-proxy-service/config"
	"github.com/blcvn/backend/services/ai-proxy-service/controllers"
	"github.com/blcvn/backend/services/ai-proxy-service/resilience"
	"github.com/blcvn/backend/services/ai-proxy-service/router"
	"github.com/blcvn/backend/services/ai-proxy-service/usecases"
	aimodel "github.com/blcvn/kratos-proto/go/ai-model"
	aiproxy "github.com/blcvn/kratos-proto/go/ai-proxy"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Starting AI Proxy Service on port %s", cfg.Port)

	// Connect to AI Model Service
	conn, err := grpc.NewClient(cfg.AIModelServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to AI Model Service: %v", err)
	}
	defer conn.Close()

	modelServiceClient := aimodel.NewAIModelServiceClient(conn)

	// Initialize components
	redisCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	circuitBreaker := resilience.NewCircuitBreakerManager()
	routerInstance := router.NewRouter(modelServiceClient)

	// Initialize usecase
	usecase := usecases.NewProxyUsecase(routerInstance, redisCache, circuitBreaker, modelServiceClient)

	// Initialize controller
	controller := controllers.NewProxyController(usecase)

	// Start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Metrics server listening on :%s", cfg.MetricsPort)
		if err := http.ListenAndServe(":"+cfg.MetricsPort, nil); err != nil {
			log.Fatalf("failed to start metrics server: %v", err)
		}
	}()

	// gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	aiproxy.RegisterAIProxyServiceServer(s, controller)

	log.Printf("AI Proxy Service listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
