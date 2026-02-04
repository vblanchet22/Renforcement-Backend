package service

import (
	"context"
	"fmt"

	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

// ColocationService handles colocation business logic
type ColocationService struct {
	repo *postgres.ColocationRepository
}

// NewColocationService creates a new ColocationService
func NewColocationService(repo *postgres.ColocationRepository) *ColocationService {
	return &ColocationService{repo: repo}
}

// ColocationWithRole contains colocation data with the current user's role
type ColocationWithRole struct {
	*domain.Colocation
	CurrentUserRole string
	MemberCount     int
}

// Create creates a new colocation and adds the creator as admin
func (s *ColocationService) Create(ctx context.Context, name string, description, address *string) (*ColocationWithRole, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	coloc := &domain.Colocation{
		Name:        name,
		Description: description,
		Address:     address,
		CreatedBy:   userID,
	}

	if err := s.repo.Create(ctx, coloc); err != nil {
		return nil, err
	}

	// Add creator as admin
	if err := s.repo.AddMember(ctx, coloc.ID, userID, domain.RoleAdmin); err != nil {
		return nil, err
	}

	return &ColocationWithRole{
		Colocation:      coloc,
		CurrentUserRole: domain.RoleAdmin,
		MemberCount:     1,
	}, nil
}

// GetByID retrieves a colocation by ID
func (s *ColocationService) GetByID(ctx context.Context, id string) (*ColocationWithRole, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	coloc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if coloc == nil {
		return nil, fmt.Errorf("colocation introuvable")
	}

	// Check if user is a member
	member, err := s.repo.GetMember(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	memberCount, err := s.repo.CountMembers(ctx, id)
	if err != nil {
		return nil, err
	}

	return &ColocationWithRole{
		Colocation:      coloc,
		CurrentUserRole: member.Role,
		MemberCount:     memberCount,
	}, nil
}

// List retrieves all colocations for the current user
func (s *ColocationService) List(ctx context.Context) ([]ColocationWithRole, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	colocations, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []ColocationWithRole
	for _, coloc := range colocations {
		member, err := s.repo.GetMember(ctx, coloc.ID, userID)
		if err != nil {
			return nil, err
		}

		memberCount, err := s.repo.CountMembers(ctx, coloc.ID)
		if err != nil {
			return nil, err
		}

		c := coloc // Create new variable to avoid pointer issues
		result = append(result, ColocationWithRole{
			Colocation:      &c,
			CurrentUserRole: member.Role,
			MemberCount:     memberCount,
		})
	}

	return result, nil
}

// Update updates a colocation (admin only)
func (s *ColocationService) Update(ctx context.Context, id string, name, description, address *string) (*ColocationWithRole, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user is admin
	member, err := s.repo.GetMember(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if member == nil || member.Role != domain.RoleAdmin {
		return nil, fmt.Errorf("seuls les administrateurs peuvent modifier la colocation")
	}

	coloc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if coloc == nil {
		return nil, fmt.Errorf("colocation introuvable")
	}

	if name != nil {
		coloc.Name = *name
	}
	if description != nil {
		coloc.Description = description
	}
	if address != nil {
		coloc.Address = address
	}

	if err := s.repo.Update(ctx, coloc); err != nil {
		return nil, err
	}

	memberCount, err := s.repo.CountMembers(ctx, id)
	if err != nil {
		return nil, err
	}

	return &ColocationWithRole{
		Colocation:      coloc,
		CurrentUserRole: member.Role,
		MemberCount:     memberCount,
	}, nil
}

// Delete deletes a colocation (admin only)
func (s *ColocationService) Delete(ctx context.Context, id string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Check if user is admin
	member, err := s.repo.GetMember(ctx, id, userID)
	if err != nil {
		return err
	}
	if member == nil || member.Role != domain.RoleAdmin {
		return fmt.Errorf("seuls les administrateurs peuvent supprimer la colocation")
	}

	return s.repo.Delete(ctx, id)
}

// Join joins a colocation using an invite code
func (s *ColocationService) Join(ctx context.Context, inviteCode string) (*ColocationWithRole, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	coloc, err := s.repo.GetByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, err
	}
	if coloc == nil {
		return nil, fmt.Errorf("code d'invitation invalide")
	}

	// Check if already a member
	existing, err := s.repo.GetMember(ctx, coloc.ID, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("vous etes deja membre de cette colocation")
	}

	// Add as member
	if err := s.repo.AddMember(ctx, coloc.ID, userID, domain.RoleMember); err != nil {
		return nil, err
	}

	memberCount, err := s.repo.CountMembers(ctx, coloc.ID)
	if err != nil {
		return nil, err
	}

	return &ColocationWithRole{
		Colocation:      coloc,
		CurrentUserRole: domain.RoleMember,
		MemberCount:     memberCount,
	}, nil
}

// Leave leaves a colocation
func (s *ColocationService) Leave(ctx context.Context, id string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	member, err := s.repo.GetMember(ctx, id, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	// Check if user is the only admin
	if member.Role == domain.RoleAdmin {
		members, err := s.repo.ListMembers(ctx, id)
		if err != nil {
			return err
		}

		adminCount := 0
		for _, m := range members {
			if m.Role == domain.RoleAdmin {
				adminCount++
			}
		}

		if adminCount == 1 && len(members) > 1 {
			return fmt.Errorf("vous devez nommer un autre administrateur avant de quitter")
		}
	}

	return s.repo.RemoveMember(ctx, id, userID)
}

// GetMembers retrieves all members of a colocation
func (s *ColocationService) GetMembers(ctx context.Context, colocationID string) ([]domain.ColocationMember, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user is a member
	member, err := s.repo.GetMember(ctx, colocationID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	return s.repo.ListMembers(ctx, colocationID)
}

// RemoveMember removes a member (admin only)
func (s *ColocationService) RemoveMember(ctx context.Context, colocationID, targetUserID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Check if user is admin
	member, err := s.repo.GetMember(ctx, colocationID, userID)
	if err != nil {
		return err
	}
	if member == nil || member.Role != domain.RoleAdmin {
		return fmt.Errorf("seuls les administrateurs peuvent retirer des membres")
	}

	// Cannot remove yourself
	if targetUserID == userID {
		return fmt.Errorf("utilisez la fonction quitter pour vous retirer")
	}

	return s.repo.RemoveMember(ctx, colocationID, targetUserID)
}

// UpdateMemberRole updates a member's role (admin only)
func (s *ColocationService) UpdateMemberRole(ctx context.Context, colocationID, targetUserID, role string) (*domain.ColocationMember, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user is admin
	member, err := s.repo.GetMember(ctx, colocationID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil || member.Role != domain.RoleAdmin {
		return nil, fmt.Errorf("seuls les administrateurs peuvent modifier les roles")
	}

	// Validate role
	if role != domain.RoleAdmin && role != domain.RoleMember {
		return nil, fmt.Errorf("role invalide")
	}

	if err := s.repo.UpdateMemberRole(ctx, colocationID, targetUserID, role); err != nil {
		return nil, err
	}

	return s.repo.GetMember(ctx, colocationID, targetUserID)
}

// RegenerateInviteCode regenerates the invite code (admin only)
func (s *ColocationService) RegenerateInviteCode(ctx context.Context, id string) (string, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return "", err
	}

	// Check if user is admin
	member, err := s.repo.GetMember(ctx, id, userID)
	if err != nil {
		return "", err
	}
	if member == nil || member.Role != domain.RoleAdmin {
		return "", fmt.Errorf("seuls les administrateurs peuvent regenerer le code")
	}

	return s.repo.RegenerateInviteCode(ctx, id)
}

// SendInvitation sends an invitation by email
func (s *ColocationService) SendInvitation(ctx context.Context, colocationID, email string) (*domain.ColocationInvitation, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user is a member
	member, err := s.repo.GetMember(ctx, colocationID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	inv := &domain.ColocationInvitation{
		ColocationID: colocationID,
		InvitedBy:    userID,
		InvitedEmail: email,
	}

	if err := s.repo.CreateInvitation(ctx, inv); err != nil {
		return nil, err
	}

	return inv, nil
}

// ListInvitations lists pending invitations
func (s *ColocationService) ListInvitations(ctx context.Context, colocationID string) ([]domain.ColocationInvitation, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user is a member
	member, err := s.repo.GetMember(ctx, colocationID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	return s.repo.ListInvitations(ctx, colocationID)
}

// CancelInvitation cancels an invitation
func (s *ColocationService) CancelInvitation(ctx context.Context, colocationID, invitationID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Check if user is a member
	member, err := s.repo.GetMember(ctx, colocationID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	return s.repo.DeleteInvitation(ctx, colocationID, invitationID)
}
