package handler

import (
	"context"

	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/service"
	"github.com/vblanchet22/back_coloc/internal/utils"
	pb "github.com/vblanchet22/back_coloc/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthHandler implements the AuthService gRPC server
type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	service *service.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register creates a new user account
func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	// Validate request
	if req.Email == "" || req.Password == "" || req.Nom == "" || req.Prenom == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email, mot de passe, nom et prenom sont obligatoires")
	}

	var telephone *string
	if req.Telephone != nil {
		telephone = req.Telephone
	}

	result, err := h.service.Register(ctx, req.Email, req.Password, req.Nom, req.Prenom, telephone)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return toAuthResponse(result), nil
}

// Login authenticates a user
func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email et mot de passe sont obligatoires")
	}

	result, err := h.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%v", err)
	}

	return toAuthResponse(result), nil
}

// RefreshToken generates new tokens
func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.AuthResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh token obligatoire")
	}

	result, err := h.service.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "%v", err)
	}

	return toAuthResponse(result), nil
}

// Logout invalidates a refresh token
func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh token obligatoire")
	}

	err := h.service.Logout(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.LogoutResponse{Success: true}, nil
}

func toAuthResponse(result *service.AuthResult) *pb.AuthResponse {
	user := result.User
	userInfo := &pb.UserInfo{
		Id:     user.ID,
		Email:  user.Email,
		Nom:    user.Nom,
		Prenom: user.Prenom,
	}

	if user.Telephone != nil {
		userInfo.Telephone = user.Telephone
	}
	if user.AvatarURL != nil {
		userInfo.AvatarUrl = user.AvatarURL
	}

	return &pb.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		User:         userInfo,
	}
}

// DomainUserToProto converts a domain User to protobuf User
func DomainUserToProto(user *domain.User) *pb.User {
	// Extract creation timestamp from ULID
	createdAt := ""
	if createdTime, err := utils.ULIDToTime(user.ID); err == nil {
		createdAt = utils.FormatFrenchDateTime(createdTime)
	}

	return &pb.User{
		Id:        user.ID,
		Email:     user.Email,
		Nom:       user.Nom,
		Prenom:    user.Prenom,
		Telephone: user.Telephone,
		AvatarUrl: user.AvatarURL,
		IsActive:  user.IsActive,
		CreatedAt: createdAt,
		UpdatedAt: utils.FormatFrenchDateTime(user.UpdatedAt),
	}
}
