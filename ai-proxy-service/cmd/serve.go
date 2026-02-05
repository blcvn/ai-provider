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
	"github.com/blcvn/backend/services/ai-proxy-service/usecases"
	pb "github.com/blcvn/kratos-proto/go/ai-proxy"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the AI Proxy Service",
	Run:   runServe,
}

func runServe(cmd *cobra.Command, args []string) {
	modelSvcAddr := getEnv("AI_MODEL_SERVICE_ADDR", "localhost:9085")
	grpcPort := getEnv("GRPC_PORT", "9087")
	httpPort := getEnv("HTTP_PORT", "8087")

	modelClient, err := helper.NewAIModelClient(modelSvcAddr)
	if err != nil {
		log.Fatalf("Failed to connect to AI Model Service: %v", err)
	}

	usecase := usecases.NewAIProxyUsecase(modelClient)

	// Register Default OpenAI Provider (or load from config)
	// In production, we'd load these from a DB or Config
	// For now, usecase dynamically gets creds and should probably
	// have a factory for providers based on creds.Provider

	// Wrap usecase with a logic to create providers on the fly or keep a map
	// I'll update the usecase to take a ProviderFactory

	controller := controllers.NewAIProxyController(usecase)

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
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = pb.RegisterAIProxyServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%s", grpcPort), opts)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

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
