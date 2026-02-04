package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/config"
	handler "github.com/vblanchet22/back_coloc/internal/grpc"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
	"github.com/vblanchet22/back_coloc/internal/service"
	pb "github.com/vblanchet22/back_coloc/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type server struct {
	cfg               *config.Config
	pool              *pgxpool.Pool
	jwtManager        *auth.JWTManager
	authHandler       *handler.AuthHandler
	userHandler       *handler.UserHandler
	colocationHandler *handler.ColocationHandler
	categoryHandler   *handler.CategoryHandler
	expenseHandler    *handler.ExpenseHandler
	balanceHandler    *handler.BalanceHandler
	paymentHandler    *handler.PaymentHandler
	decisionHandler   *handler.DecisionHandler
	fundHandler          *handler.FundHandler
	notificationHandler  *handler.NotificationHandler
}

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

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	// Initialize repositories
	authRepo := postgres.NewAuthRepository(pool)
	colocationRepo := postgres.NewColocationRepository(pool)
	categoryRepo := postgres.NewCategoryRepository(pool)
	expenseRepo := postgres.NewExpenseRepository(pool)
	balanceRepo := postgres.NewBalanceRepository(pool)
	paymentRepo := postgres.NewPaymentRepository(pool)
	decisionRepo := postgres.NewDecisionRepository(pool)
	fundRepo := postgres.NewFundRepository(pool)
	notificationRepo := postgres.NewNotificationRepository(pool)

	// Initialize services
	authService := service.NewAuthService(authRepo, jwtManager)
	userService := service.NewUserService(authRepo)
	colocationService := service.NewColocationService(colocationRepo)
	categoryService := service.NewCategoryService(categoryRepo, colocationRepo)
	expenseService := service.NewExpenseService(expenseRepo, colocationRepo, categoryRepo)
	balanceService := service.NewBalanceService(balanceRepo, colocationRepo)
	paymentService := service.NewPaymentService(paymentRepo, colocationRepo)
	decisionService := service.NewDecisionService(decisionRepo, colocationRepo)
	fundService := service.NewFundService(fundRepo, colocationRepo)
	notificationService := service.NewNotificationService(notificationRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	colocationHandler := handler.NewColocationHandler(colocationService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	balanceHandler := handler.NewBalanceHandler(balanceService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	decisionHandler := handler.NewDecisionHandler(decisionService)
	fundHandler := handler.NewFundHandler(fundService)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	srv := &server{
		cfg:               cfg,
		pool:              pool,
		jwtManager:        jwtManager,
		authHandler:       authHandler,
		userHandler:       userHandler,
		colocationHandler: colocationHandler,
		categoryHandler:   categoryHandler,
		expenseHandler:    expenseHandler,
		balanceHandler:    balanceHandler,
		paymentHandler:    paymentHandler,
		decisionHandler:   decisionHandler,
		fundHandler:          fundHandler,
		notificationHandler:  notificationHandler,
	}

	// Start gRPC server in goroutine
	go func() {
		if err := srv.runGRPCServer(); err != nil {
			log.Fatalf("Erreur serveur gRPC: %v", err)
		}
	}()

	// Start HTTP gateway
	if err := srv.runHTTPGateway(); err != nil {
		log.Fatalf("Erreur gateway HTTP: %v", err)
	}
}

func (s *server) runGRPCServer() error {
	lis, err := net.Listen("tcp", ":"+s.cfg.Server.GRPCPort)
	if err != nil {
		return err
	}

	// Create auth interceptor
	authInterceptor := auth.NewAuthInterceptor(s.jwtManager)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)

	// Register gRPC services
	pb.RegisterAuthServiceServer(grpcServer, s.authHandler)
	pb.RegisterUserServiceServer(grpcServer, s.userHandler)
	pb.RegisterColocationServiceServer(grpcServer, s.colocationHandler)
	pb.RegisterCategoryServiceServer(grpcServer, s.categoryHandler)
	pb.RegisterExpenseServiceServer(grpcServer, s.expenseHandler)
	pb.RegisterBalanceServiceServer(grpcServer, s.balanceHandler)
	pb.RegisterPaymentServiceServer(grpcServer, s.paymentHandler)
	pb.RegisterDecisionServiceServer(grpcServer, s.decisionHandler)
	pb.RegisterFundServiceServer(grpcServer, s.fundHandler)
	pb.RegisterNotificationServiceServer(grpcServer, s.notificationHandler)

	// Enable reflection for grpcurl/grpcui
	reflection.Register(grpcServer)

	log.Printf("Serveur gRPC demarre sur le port %s", s.cfg.Server.GRPCPort)
	return grpcServer.Serve(lis)
}

func (s *server) runHTTPGateway() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := "localhost:" + s.cfg.Server.GRPCPort

	// Register HTTP handlers
	if err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterColocationServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterCategoryServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterExpenseServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterBalanceServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterPaymentServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterDecisionServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterFundServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return err
	}
	if err := pb.RegisterNotificationServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return err
	}

	// Wrap with CORS
	handler := corsMiddleware(mux)

	// Add health check
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	httpMux.Handle("/", handler)

	log.Printf("Gateway REST demarree sur le port %s", s.cfg.Server.HTTPPort)
	log.Printf("API disponible sur http://localhost:%s/api/", s.cfg.Server.HTTPPort)

	return http.ListenAndServe(":"+s.cfg.Server.HTTPPort, httpMux)
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
