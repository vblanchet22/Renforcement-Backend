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

// ColocationHandler implements the ColocationService gRPC server
type ColocationHandler struct {
	pb.UnimplementedColocationServiceServer
	service *service.ColocationService
}

// NewColocationHandler creates a new ColocationHandler
func NewColocationHandler(service *service.ColocationService) *ColocationHandler {
	return &ColocationHandler{service: service}
}

// CreateColocation creates a new colocation
func (h *ColocationHandler) CreateColocation(ctx context.Context, req *pb.CreateColocationRequest) (*pb.Colocation, error) {
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nom obligatoire")
	}

	result, err := h.service.Create(ctx, req.Name, req.Description, req.Address)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return colocationWithRoleToProto(result), nil
}

// GetColocation retrieves a colocation by ID
func (h *ColocationHandler) GetColocation(ctx context.Context, req *pb.GetColocationRequest) (*pb.Colocation, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id obligatoire")
	}

	result, err := h.service.GetByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}

	return colocationWithRoleToProto(result), nil
}

// ListColocations lists all colocations for the current user
func (h *ColocationHandler) ListColocations(ctx context.Context, req *pb.ListColocationsRequest) (*pb.ListColocationsResponse, error) {
	results, err := h.service.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var colocations []*pb.Colocation
	for _, result := range results {
		colocations = append(colocations, colocationWithRoleToProto(&result))
	}

	return &pb.ListColocationsResponse{Colocations: colocations}, nil
}

// UpdateColocation updates a colocation
func (h *ColocationHandler) UpdateColocation(ctx context.Context, req *pb.UpdateColocationRequest) (*pb.Colocation, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id obligatoire")
	}

	result, err := h.service.Update(ctx, req.Id, req.Name, req.Description, req.Address)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return colocationWithRoleToProto(result), nil
}

// DeleteColocation deletes a colocation
func (h *ColocationHandler) DeleteColocation(ctx context.Context, req *pb.DeleteColocationRequest) (*pb.DeleteColocationResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id obligatoire")
	}

	if err := h.service.Delete(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.DeleteColocationResponse{Success: true}, nil
}

// JoinColocation joins a colocation with an invite code
func (h *ColocationHandler) JoinColocation(ctx context.Context, req *pb.JoinColocationRequest) (*pb.Colocation, error) {
	if req.InviteCode == "" {
		return nil, status.Errorf(codes.InvalidArgument, "code d'invitation obligatoire")
	}

	result, err := h.service.Join(ctx, req.InviteCode)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return colocationWithRoleToProto(result), nil
}

// LeaveColocation leaves a colocation
func (h *ColocationHandler) LeaveColocation(ctx context.Context, req *pb.LeaveColocationRequest) (*pb.LeaveColocationResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id obligatoire")
	}

	if err := h.service.Leave(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.LeaveColocationResponse{Success: true}, nil
}

// GetMembers retrieves all members of a colocation
func (h *ColocationHandler) GetMembers(ctx context.Context, req *pb.GetMembersRequest) (*pb.GetMembersResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	members, err := h.service.GetMembers(ctx, req.ColocationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbMembers []*pb.ColocationMember
	for _, member := range members {
		pbMembers = append(pbMembers, memberToProto(&member))
	}

	return &pb.GetMembersResponse{Members: pbMembers}, nil
}

// RemoveMember removes a member from a colocation
func (h *ColocationHandler) RemoveMember(ctx context.Context, req *pb.RemoveMemberRequest) (*pb.RemoveMemberResponse, error) {
	if req.ColocationId == "" || req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et user_id obligatoires")
	}

	if err := h.service.RemoveMember(ctx, req.ColocationId, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.RemoveMemberResponse{Success: true}, nil
}

// UpdateMemberRole updates a member's role
func (h *ColocationHandler) UpdateMemberRole(ctx context.Context, req *pb.UpdateMemberRoleRequest) (*pb.ColocationMember, error) {
	if req.ColocationId == "" || req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et user_id obligatoires")
	}

	role := protoRoleToString(req.Role)
	member, err := h.service.UpdateMemberRole(ctx, req.ColocationId, req.UserId, role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return memberToProto(member), nil
}

// RegenerateInviteCode regenerates the invite code
func (h *ColocationHandler) RegenerateInviteCode(ctx context.Context, req *pb.RegenerateInviteCodeRequest) (*pb.RegenerateInviteCodeResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id obligatoire")
	}

	code, err := h.service.RegenerateInviteCode(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.RegenerateInviteCodeResponse{InviteCode: code}, nil
}

// SendInvitation sends an invitation
func (h *ColocationHandler) SendInvitation(ctx context.Context, req *pb.SendInvitationRequest) (*pb.Invitation, error) {
	if req.ColocationId == "" || req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et email obligatoires")
	}

	inv, err := h.service.SendInvitation(ctx, req.ColocationId, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return invitationToProto(inv), nil
}

// ListInvitations lists pending invitations
func (h *ColocationHandler) ListInvitations(ctx context.Context, req *pb.ListInvitationsRequest) (*pb.ListInvitationsResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	invitations, err := h.service.ListInvitations(ctx, req.ColocationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbInvitations []*pb.Invitation
	for _, inv := range invitations {
		pbInvitations = append(pbInvitations, invitationToProto(&inv))
	}

	return &pb.ListInvitationsResponse{Invitations: pbInvitations}, nil
}

// CancelInvitation cancels an invitation
func (h *ColocationHandler) CancelInvitation(ctx context.Context, req *pb.CancelInvitationRequest) (*pb.CancelInvitationResponse, error) {
	if req.ColocationId == "" || req.InvitationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et invitation_id obligatoires")
	}

	if err := h.service.CancelInvitation(ctx, req.ColocationId, req.InvitationId); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.CancelInvitationResponse{Success: true}, nil
}

// Helper functions

func colocationWithRoleToProto(c *service.ColocationWithRole) *pb.Colocation {
	return &pb.Colocation{
		Id:              c.ID,
		Name:            c.Name,
		Description:     c.Description,
		Address:         c.Address,
		CreatedBy:       c.CreatedBy,
		InviteCode:      c.InviteCode,
		CreatedAt:       utils.FormatFrenchDateTime(c.CreatedAt),
		UpdatedAt:       utils.FormatFrenchDateTime(c.UpdatedAt),
		CurrentUserRole: stringToProtoRole(c.CurrentUserRole),
		MemberCount:     int32(c.MemberCount),
	}
}

func memberToProto(m *domain.ColocationMember) *pb.ColocationMember {
	return &pb.ColocationMember{
		Id:           m.ID,
		UserId:       m.UserID,
		ColocationId: m.ColocationID,
		Role:         stringToProtoRole(m.Role),
		JoinedAt:     utils.FormatFrenchDateTime(m.JoinedAt),
		Email:        m.Email,
		Nom:          m.Nom,
		Prenom:       m.Prenom,
		AvatarUrl:    m.AvatarURL,
	}
}

func invitationToProto(i *domain.ColocationInvitation) *pb.Invitation {
	return &pb.Invitation{
		Id:           i.ID,
		ColocationId: i.ColocationID,
		InvitedBy:    i.InvitedBy,
		InvitedEmail: i.InvitedEmail,
		Status:       stringToProtoInvitationStatus(i.Status),
		ExpiresAt:    utils.FormatFrenchDateTime(i.ExpiresAt),
		CreatedAt:    utils.FormatFrenchDateTime(i.CreatedAt),
	}
}

func stringToProtoRole(role string) pb.MemberRole {
	switch role {
	case domain.RoleAdmin:
		return pb.MemberRole_MEMBER_ROLE_ADMIN
	case domain.RoleMember:
		return pb.MemberRole_MEMBER_ROLE_MEMBER
	default:
		return pb.MemberRole_MEMBER_ROLE_UNSPECIFIED
	}
}

func protoRoleToString(role pb.MemberRole) string {
	switch role {
	case pb.MemberRole_MEMBER_ROLE_ADMIN:
		return domain.RoleAdmin
	case pb.MemberRole_MEMBER_ROLE_MEMBER:
		return domain.RoleMember
	default:
		return domain.RoleMember
	}
}

func stringToProtoInvitationStatus(status string) pb.InvitationStatus {
	switch status {
	case domain.InvitationStatusPending:
		return pb.InvitationStatus_INVITATION_STATUS_PENDING
	case domain.InvitationStatusAccepted:
		return pb.InvitationStatus_INVITATION_STATUS_ACCEPTED
	case domain.InvitationStatusRejected:
		return pb.InvitationStatus_INVITATION_STATUS_REJECTED
	case domain.InvitationStatusExpired:
		return pb.InvitationStatus_INVITATION_STATUS_EXPIRED
	default:
		return pb.InvitationStatus_INVITATION_STATUS_UNSPECIFIED
	}
}
