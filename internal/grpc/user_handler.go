package handler

import (
	"context"

	"github.com/vblanchet22/back_coloc/internal/service"
	pb "github.com/vblanchet22/back_coloc/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserHandler implements the UserService gRPC server
type UserHandler struct {
	pb.UnimplementedUserServiceServer
	service *service.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetCurrentUser retrieves the current authenticated user
func (h *UserHandler) GetCurrentUser(ctx context.Context, req *pb.GetCurrentUserRequest) (*pb.User, error) {
	user, err := h.service.GetCurrentUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return DomainUserToProto(user), nil
}

// UpdateCurrentUser updates the current user's profile
func (h *UserHandler) UpdateCurrentUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	input := service.UpdateUserInput{}

	if req.Nom != nil {
		input.Nom = req.Nom
	}
	if req.Prenom != nil {
		input.Prenom = req.Prenom
	}
	if req.Telephone != nil {
		input.Telephone = req.Telephone
	}
	if req.AvatarUrl != nil {
		input.AvatarURL = req.AvatarUrl
	}
	if req.CurrentPassword != nil {
		input.CurrentPassword = req.CurrentPassword
	}
	if req.NewPassword != nil {
		input.NewPassword = req.NewPassword
	}

	user, err := h.service.UpdateCurrentUser(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return DomainUserToProto(user), nil
}

// DeleteCurrentUser deletes the current user's account
func (h *UserHandler) DeleteCurrentUser(ctx context.Context, req *pb.DeleteCurrentUserRequest) (*pb.DeleteUserResponse, error) {
	err := h.service.DeleteCurrentUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.DeleteUserResponse{Success: true}, nil
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id obligatoire")
	}

	user, err := h.service.GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}

	return DomainUserToProto(user), nil
}
