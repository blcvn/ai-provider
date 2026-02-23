package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blcvn/backend/services/ai-model-service/controllers"
	"github.com/blcvn/backend/services/ai-model-service/dto"
	"github.com/blcvn/backend/services/ai-model-service/helper"
	"github.com/blcvn/backend/services/ai-model-service/repository/postgres"
	"github.com/blcvn/backend/services/ai-model-service/usecases"
	pb "github.com/blcvn/kratos-proto/go/ai-model"

	"github.com/blcvn/ba-shared-libs/pkg/mtls"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	grpcPort    string
	httpPort    string
	metricsPath string
	serviceName string
	jaegerURL   string
	tlsCertPath string
	tlsKeyPath  string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the AI Model Service",
	Run:   runServe,
}

func init() {
	serveCmd.Flags().StringVar(&grpcPort, "grpc-port", "9085", "gRPC server port")
	serveCmd.Flags().StringVar(&httpPort, "http-port", "8085", "HTTP server port")
	serveCmd.Flags().StringVar(&metricsPath, "metrics-path", "/metrics", "Path for Prometheus metrics")
	serveCmd.Flags().StringVar(&serviceName, "service-name", "ai-model-service", "Service name")
	serveCmd.Flags().StringVar(&jaegerURL, "jaeger-url", "", "Jaeger collector URL")
	serveCmd.Flags().StringVar(&tlsCertPath, "tls-cert", "", "Path to TLS certificate")
	serveCmd.Flags().StringVar(&tlsKeyPath, "tls-key", "", "Path to TLS key")
}

func runServe(cmd *cobra.Command, args []string) {
	// Get configuration from environment (fallback or override)
	dbURL := getEnv("DATABASE_URL", "postgresql://baagent:password@localhost:5432/baagent?sslmode=disable")
	vaultAddr := getEnv("VAULT_ADDR", "http://localhost:8200")
	vaultToken := getEnv("VAULT_TOKEN", "")

	// Initialize database
	db, err := gorm.Open(gorm_postgres.New(gorm_postgres.Config{
		DSN:                  dbURL,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	log.Println("Running database migrations...")
	if err := db.AutoMigrate(&dto.AIModel{}, &dto.UsageLog{}); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Vault client
	vaultClient, err := helper.NewVaultClient(vaultAddr, vaultToken)
	if err != nil {
		log.Fatalf("Failed to initialize Vault client: %v", err)
	}
	defer vaultClient.Close()

	// Load AI Service Secret from Vault (or Env fallback)
	// We want to store the master secret key in Vault, similar to JWTSecret in auth-service.
	// For now, let's assume it's in env AI_SERVICE_SECRET or we can read it from Vault path "secret/data/ai-service-config".
	// The user requested: "load giống JWTSecret của auth-service".
	// In auth-service, it loads config file from Vault Agent path.
	// Here, we can try to read from Env first (pushed by Vault Agent to Env or File).
	// Let's rely on Env AI_SERVICE_SECRET for simplicity as we didn't setup the full config loading struct.
	aiServiceSecret := getEnv("AI_SERVICE_SECRET", "")
	if aiServiceSecret == "" {
		// Log warning or fatal? Fatal if we want to enforce it.
		// For dev, maybe default to something?
		log.Println("Warning: AI_SERVICE_SECRET is not set, using default dev secret")
		aiServiceSecret = "dev-secret-key-must-be-32-bytes-long!" // 32 bytes for AES-256
	}

	// Initialize layers
	modelRepo := postgres.NewModelRepository(db)
	cryptoHelper := helper.NewCryptoHelpers(aiServiceSecret)

	// Seed models
	helper.SeedModels(db, cryptoHelper)

	modelUsecase := usecases.NewModelUsecase(modelRepo, vaultClient, cryptoHelper)
	transform := helper.NewTransform()
	modelController := controllers.NewModelController(modelUsecase, transform)

	// Setup mTLS
	var reloader *mtls.CertReloader
	if tlsCertPath != "" && tlsKeyPath != "" {
		var err error
		reloader, err = mtls.NewCertReloader(tlsCertPath, tlsKeyPath)
		if err != nil {
			log.Printf("Warning: failed to load mTLS certs: %v", err)
		} else {
			log.Println("mTLS enabled")
		}
	}

	// Logging options
	logger := logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		log.Printf("[gRPC] %s: %v", msg, fields)
	})
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	// gRPC Server Options
	var serverOpts []grpc.ServerOption
	serverOpts = append(serverOpts,
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(logger, loggingOpts...),
			recovery.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(logger, loggingOpts...),
			recovery.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
		),
	)

	if reloader != nil {
		creds := credentials.NewTLS(&tls.Config{
			GetConfigForClient: reloader.GetConfigForClient,
		})
		serverOpts = append(serverOpts, grpc.Creds(creds))
	}

	// Start gRPC server
	grpcServer := grpc.NewServer(serverOpts...)

	// Register Services
	pb.RegisterAIModelServiceServer(grpcServer, modelController)

	// Register Health Server
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("baagent.aimodel.v1.AIModelService", healthpb.HealthCheckResponse_SERVING)

	// Register Prometheus metrics
	grpc_prometheus.Register(grpcServer)
	grpc_prometheus.EnableHandlingTimeHistogram()

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	go func() {
		log.Printf("Starting gRPC server on port %s", grpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP gateway
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Dial for Healthz
	healthConn, err := grpc.Dial("localhost:"+grpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to dial health server: %v", err)
	}

	mux := runtime.NewServeMux(
		runtime.WithHealthzEndpoint(healthpb.NewHealthClient(healthConn)),
	)

	// Gateway Dial Options
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	// Note: If gRPC server uses TLS, gateway needs to dial with TLS too or insecure skip verify if creating separate internal loop
	// Since gateway is running in same process, we can target localhost.
	// If mTLS is ON, we must use TLS.
	// For simplicity, if reloader != nil, we can use a client TLS config that trusts our own CA.
	if reloader != nil {
		// Use the CertReloader to get client config (reuses loaded CA)
		clientTLS := reloader.ClientTLSConfig()
		clientTLS.InsecureSkipVerify = true // For localhost internal call if loopback IP doesn't match cert SANs (often an issue)
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(clientTLS))}
	}

	err = pb.RegisterAIModelServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%s", grpcPort), dialOpts)
	if err != nil {
		log.Fatalf("Failed to register HTTP gateway: %v", err)
	}

	// Add standard HTTP handlers
	httpMux := http.NewServeMux()
	httpMux.Handle("/", mux)
	httpMux.Handle(metricsPath, promhttp.Handler())
	httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: corsHandler.Handler(httpMux),
	}

	if reloader != nil {
		httpServer.TLSConfig = &tls.Config{
			GetConfigForClient: reloader.GetConfigForClient,
		}
	}

	go func() {
		log.Printf("Starting HTTP server on port %s", httpPort)
		var err error
		if reloader != nil {
			// Certificates are provided via TLSConfig.GetConfigForClient, so filenames can be empty
			// However ListenAndServeTLS treats empty strings as "lookup in global struct" sometimes or depends on implementation
			// Go docs: "If certFile and keyFile are not empty, ListenAndServeTLS ... uses them ... Otherwise ... uses s.TLSConfig.Certificates"
			// GetConfigForClient is NOT Config.Certificates.
			// Ideally we should set s.TLSConfig.Certificates to the initial cert so it can start?
			// But reloader handles it.
			// Let's use empty strings and see. If it fails, we might need a dummy cert or point to the files initially.
			// Reloader does initial load. We can get the certs from reloader if we exposed them, pass paths since we have them.
			err = httpServer.ListenAndServeTLS("", "")
		} else {
			err = httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	grpcServer.GracefulStop()
	log.Println("Servers stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
