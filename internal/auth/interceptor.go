package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// EmailKey is the context key for user email
	EmailKey ContextKey = "email"
)

// AuthInterceptor handles authentication for gRPC requests
type AuthInterceptor struct {
	jwtManager      *JWTManager
	publicMethods   map[string]bool
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(jwtManager *JWTManager) *AuthInterceptor {
	// Methods that don't require authentication
	publicMethods := map[string]bool{
		"/coloc.AuthService/Register":     true,
		"/coloc.AuthService/Login":        true,
		"/coloc.AuthService/RefreshToken": true,
	}

	return &AuthInterceptor{
		jwtManager:    jwtManager,
		publicMethods: publicMethods,
	}
}

// Unary returns a unary interceptor for authentication
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip authentication for public methods
		if i.publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Authenticate the request
		newCtx, err := i.authenticate(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// Stream returns a stream interceptor for authentication
func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip authentication for public methods
		if i.publicMethods[info.FullMethod] {
			return handler(srv, stream)
		}

		// Authenticate the request
		ctx := stream.Context()
		newCtx, err := i.authenticate(ctx)
		if err != nil {
			return err
		}

		// Wrap the stream with the new context
		wrapped := &wrappedStream{
			ServerStream: stream,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

func (i *AuthInterceptor) authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata manquante")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "token d'autorisation manquant")
	}

	token := values[0]
	// Remove "Bearer " prefix if present
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	claims, err := i.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "token invalide: %v", err)
	}

	// Add user info to context
	newCtx := context.WithValue(ctx, UserIDKey, claims.UserID)
	newCtx = context.WithValue(newCtx, EmailKey, claims.Email)

	return newCtx, nil
}

// wrappedStream wraps a grpc.ServerStream with a custom context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// GetUserIDFromContext extracts the user ID from context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", status.Errorf(codes.Unauthenticated, "utilisateur non authentifie")
	}
	return userID, nil
}

// GetEmailFromContext extracts the email from context
func GetEmailFromContext(ctx context.Context) (string, error) {
	email, ok := ctx.Value(EmailKey).(string)
	if !ok || email == "" {
		return "", status.Errorf(codes.Unauthenticated, "utilisateur non authentifie")
	}
	return email, nil
}
