package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"github.com/vblanchet22/back_coloc/internal/config"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aucun fichier .env trouve, utilisation des valeurs par defaut")
	}

	cfg := config.Load()

	// Connect to database
	pool, err := postgres.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Erreur de connexion a la base de donnees: %v", err)
	}
	defer pool.Close()

	// Start gRPC server in goroutine
	go func() {
		if err := runGRPCServer(cfg); err != nil {
			log.Fatalf("Erreur serveur gRPC: %v", err)
		}
	}()

	// Start HTTP gateway
	if err := runHTTPGateway(cfg); err != nil {
		log.Fatalf("Erreur gateway HTTP: %v", err)
	}
}

func runGRPCServer(cfg *config.Config) error {
	lis, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	// TODO: Register gRPC services here
	// pb.RegisterAuthServiceServer(grpcServer, authHandler)
	// pb.RegisterUserServiceServer(grpcServer, userHandler)
	// pb.RegisterColocationServiceServer(grpcServer, colocationHandler)
	// etc.

	// Enable reflection for grpcurl/grpcui
	reflection.Register(grpcServer)

	log.Printf("Serveur gRPC demarre sur le port %s", cfg.Server.GRPCPort)
	return grpcServer.Serve(lis)
}

func runHTTPGateway(cfg *config.Config) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := "localhost:" + cfg.Server.GRPCPort

	// TODO: Register HTTP handlers here
	// if err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
	// 	return err
	// }
	// if err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
	// 	return err
	// }
	// etc.

	_ = grpcEndpoint
	_ = opts

	// Wrap with CORS
	handler := corsMiddleware(mux)

	// Add health check
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	httpMux.Handle("/", handler)

	log.Printf("Gateway REST demarree sur le port %s", cfg.Server.HTTPPort)
	log.Printf("API disponible sur http://localhost:%s/api/", cfg.Server.HTTPPort)

	return http.ListenAndServe(":"+cfg.Server.HTTPPort, httpMux)
}

// customHeaderMatcher allows Authorization header to pass through
func customHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Authorization":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// corsMiddleware adds CORS headers for frontend access
func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}
