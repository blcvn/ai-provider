package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/blcvn/backend/services/ai-proxy-service/controllers"
	"github.com/blcvn/backend/services/ai-proxy-service/helper"
	"github.com/blcvn/backend/services/ai-proxy-service/providers/anthropic"
	"github.com/blcvn/backend/services/ai-proxy-service/usecases"
	pb "github.com/blcvn/kratos-proto/go/ai-proxy"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the AI Proxy Service",
	Run:   runServe,
}

func init() {
	serveCmd.Flags().String("service-name", "ai-proxy-service", "Service name")
	serveCmd.Flags().String("jaeger-url", "localhost:4317", "Jaeger URL")
	serveCmd.Flags().String("metrics-path", "/metrics", "Metrics path")
	serveCmd.Flags().String("grpc-port", "9087", "gRPC port")
	serveCmd.Flags().String("http-port", "8087", "HTTP port")
}

func runServe(cmd *cobra.Command, args []string) {

	modelSvcAddr := getEnv("AI_MODEL_SERVICE_ADDR", "localhost:9085")
	grpcPort := getEnv("GRPC_PORT", "9087")
	httpPort := getEnv("HTTP_PORT", "8087")

	modelClient, err := helper.NewAIModelClient(modelSvcAddr)
	if err != nil {
		log.Fatalf("Failed to connect to AI Model Service: %v", err)
	}

	usecase := usecases.NewProxyUsecase(modelClient)

	// Register Providers
	// Note: API Key and Model ID are dynamic per request, but the factory needs initial dummy or changing the provider signature.
	// However, current implementation of providers (NewClaudeProvider) takes args.
	// But in `Complete` method of providers, we re-initialize client with request's API Key!
	// So passing empty strings here is fine for the "base" provider structure/factory logic if it existed,
	// BUT `usecases.RegisterProvider` expects an instance.
	// For now, we register instances with dummy data, assuming `Complete` overrides them.

	anthropicProvider, err := anthropic.NewClaudeProvider()
	if err != nil {
		log.Printf("Warning: Failed to init Anthropic provider template: %v", err)
	} else {
		usecase.RegisterProvider("anthropic", anthropicProvider)
	}

	// openaiProvider, err := openai.NewGPTProvider("dummy-key", "dummy-model")
	// if err != nil {
	// 	log.Printf("Warning: Failed to init OpenAI provider template: %v", err)
	// } else {
	// 	usecase.RegisterProvider("openai", openaiProvider)
	// }

	controller := controllers.NewProxyController(usecase)

	grpcServer := grpc.NewServer()
	pb.RegisterAIProxyServiceServer(grpcServer, controller)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed on grpc port: %v", err)
	}

	go func() {
		log.Printf("Starting gRPC on %s", grpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	ctx := context.Background()
	gwMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = pb.RegisterAIProxyServiceHandlerFromEndpoint(ctx, gwMux, fmt.Sprintf("localhost:%s", grpcPort), opts)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", gwMux)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: mux,
	}

	go func() {
		log.Printf("Starting HTTP on %s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	grpcServer.GracefulStop()
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
