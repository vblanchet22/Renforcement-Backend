package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/domain"
)

// ColocationRepository manages colocation database operations
type ColocationRepository struct {
	pool *pgxpool.Pool
}

// NewColocationRepository creates a new ColocationRepository instance
func NewColocationRepository(pool *pgxpool.Pool) *ColocationRepository {
	return &ColocationRepository{pool: pool}
}

// generateInviteCode generates a random 8-character invite code
func generateInviteCode() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Create creates a new colocation
func (r *ColocationRepository) Create(ctx context.Context, coloc *domain.Colocation) error {
	coloc.InviteCode = generateInviteCode()

	query := `
		INSERT INTO colocations (name, description, address, created_by, invite_code)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		coloc.Name,
		coloc.Description,
		coloc.Address,
		coloc.CreatedBy,
		coloc.InviteCode,
	).Scan(&coloc.ID, &coloc.CreatedAt, &coloc.UpdatedAt)

	if err != nil {
		return fmt.Errorf("erreur lors de la creation de la colocation: %w", err)
	}

	return nil
}

// GetByID retrieves a colocation by ID
func (r *ColocationRepository) GetByID(ctx context.Context, id string) (*domain.Colocation, error) {
	query := `
		SELECT id, name, description, address, created_by, invite_code, created_at, updated_at
		FROM colocations
		WHERE id = $1
	`

	var coloc domain.Colocation
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&coloc.ID,
		&coloc.Name,
		&coloc.Description,
		&coloc.Address,
		&coloc.CreatedBy,
		&coloc.InviteCode,
		&coloc.CreatedAt,
		&coloc.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation de la colocation: %w", err)
	}

	return &coloc, nil
}

// GetByInviteCode retrieves a colocation by invite code
func (r *ColocationRepository) GetByInviteCode(ctx context.Context, code string) (*domain.Colocation, error) {
	query := `
		SELECT id, name, description, address, created_by, invite_code, created_at, updated_at
		FROM colocations
		WHERE invite_code = $1
	`

	var coloc domain.Colocation
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&coloc.ID,
		&coloc.Name,
		&coloc.Description,
		&coloc.Address,
		&coloc.CreatedBy,
		&coloc.InviteCode,
		&coloc.CreatedAt,
		&coloc.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation de la colocation: %w", err)
	}

	return &coloc, nil
}

// ListByUserID retrieves all colocations for a user
func (r *ColocationRepository) ListByUserID(ctx context.Context, userID string) ([]domain.Colocation, error) {
	query := `
		SELECT c.id, c.name, c.description, c.address, c.created_by, c.invite_code, c.created_at, c.updated_at
		FROM colocations c
		INNER JOIN colocation_members cm ON c.id = cm.colocation_id
		WHERE cm.user_id = $1
		ORDER BY c.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation des colocations: %w", err)
	}
	defer rows.Close()

	var colocations []domain.Colocation
	for rows.Next() {
		var coloc domain.Colocation
		err := rows.Scan(
			&coloc.ID,
			&coloc.Name,
			&coloc.Description,
			&coloc.Address,
			&coloc.CreatedBy,
			&coloc.InviteCode,
			&coloc.CreatedAt,
			&coloc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du scan de la colocation: %w", err)
		}
		colocations = append(colocations, coloc)
	}

	return colocations, nil
}

// Update updates a colocation
func (r *ColocationRepository) Update(ctx context.Context, coloc *domain.Colocation) error {
	query := `
		UPDATE colocations
		SET name = $1, description = $2, address = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		coloc.Name,
		coloc.Description,
		coloc.Address,
		coloc.ID,
	).Scan(&coloc.UpdatedAt)

	if err != nil {
		return fmt.Errorf("erreur lors de la mise a jour de la colocation: %w", err)
	}

	return nil
}

// Delete deletes a colocation
func (r *ColocationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM colocations WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression de la colocation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("colocation introuvable")
	}

	return nil
}

// RegenerateInviteCode regenerates the invite code
func (r *ColocationRepository) RegenerateInviteCode(ctx context.Context, id string) (string, error) {
	newCode := generateInviteCode()

	query := `UPDATE colocations SET invite_code = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.pool.Exec(ctx, query, newCode, id)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la regeneration du code: %w", err)
	}

	if result.RowsAffected() == 0 {
		return "", fmt.Errorf("colocation introuvable")
	}

	return newCode, nil
}

// AddMember adds a member to a colocation
func (r *ColocationRepository) AddMember(ctx context.Context, colocationID, userID, role string) error {
	query := `
		INSERT INTO colocation_members (colocation_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (colocation_id, user_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, colocationID, userID, role)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ajout du membre: %w", err)
	}

	return nil
}

// RemoveMember removes a member from a colocation
func (r *ColocationRepository) RemoveMember(ctx context.Context, colocationID, userID string) error {
	query := `DELETE FROM colocation_members WHERE colocation_id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, colocationID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du membre: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("membre introuvable")
	}

	return nil
}

// GetMember retrieves a specific member
func (r *ColocationRepository) GetMember(ctx context.Context, colocationID, userID string) (*domain.ColocationMember, error) {
	query := `
		SELECT cm.id, cm.colocation_id, cm.user_id, cm.role, cm.joined_at,
		       u.email, u.nom, u.prenom, u.avatar_url
		FROM colocation_members cm
		INNER JOIN users u ON cm.user_id = u.id
		WHERE cm.colocation_id = $1 AND cm.user_id = $2
	`

	var member domain.ColocationMember
	err := r.pool.QueryRow(ctx, query, colocationID, userID).Scan(
		&member.ID,
		&member.ColocationID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
		&member.Email,
		&member.Nom,
		&member.Prenom,
		&member.AvatarURL,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation du membre: %w", err)
	}

	return &member, nil
}

// ListMembers retrieves all members of a colocation
func (r *ColocationRepository) ListMembers(ctx context.Context, colocationID string) ([]domain.ColocationMember, error) {
	query := `
		SELECT cm.id, cm.colocation_id, cm.user_id, cm.role, cm.joined_at,
		       u.email, u.nom, u.prenom, u.avatar_url
		FROM colocation_members cm
		INNER JOIN users u ON cm.user_id = u.id
		WHERE cm.colocation_id = $1
		ORDER BY cm.joined_at
	`

	rows, err := r.pool.Query(ctx, query, colocationID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation des membres: %w", err)
	}
	defer rows.Close()

	var members []domain.ColocationMember
	for rows.Next() {
		var member domain.ColocationMember
		err := rows.Scan(
			&member.ID,
			&member.ColocationID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&member.Email,
			&member.Nom,
			&member.Prenom,
			&member.AvatarURL,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du scan du membre: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

// UpdateMemberRole updates a member's role
func (r *ColocationRepository) UpdateMemberRole(ctx context.Context, colocationID, userID, role string) error {
	query := `UPDATE colocation_members SET role = $1 WHERE colocation_id = $2 AND user_id = $3`

	result, err := r.pool.Exec(ctx, query, role, colocationID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise a jour du role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("membre introuvable")
	}

	return nil
}

// CountMembers counts the number of members in a colocation
func (r *ColocationRepository) CountMembers(ctx context.Context, colocationID string) (int, error) {
	query := `SELECT COUNT(*) FROM colocation_members WHERE colocation_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, colocationID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur lors du comptage des membres: %w", err)
	}

	return count, nil
}

// IsMember checks if a user is a member of a colocation
func (r *ColocationRepository) IsMember(ctx context.Context, colocationID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM colocation_members WHERE colocation_id = $1 AND user_id = $2)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, colocationID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("erreur lors de la verification du membre: %w", err)
	}

	return exists, nil
}

// CreateInvitation creates a new invitation
func (r *ColocationRepository) CreateInvitation(ctx context.Context, inv *domain.ColocationInvitation) error {
	query := `
		INSERT INTO colocation_invitations (colocation_id, invited_by, invited_email)
		VALUES ($1, $2, $3)
		RETURNING id, status, expires_at, created_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		inv.ColocationID,
		inv.InvitedBy,
		inv.InvitedEmail,
	).Scan(&inv.ID, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt)

	if err != nil {
		return fmt.Errorf("erreur lors de la creation de l'invitation: %w", err)
	}

	return nil
}

// ListInvitations lists pending invitations for a colocation
func (r *ColocationRepository) ListInvitations(ctx context.Context, colocationID string) ([]domain.ColocationInvitation, error) {
	query := `
		SELECT id, colocation_id, invited_by, invited_email, status, expires_at, created_at
		FROM colocation_invitations
		WHERE colocation_id = $1 AND status = 'pending' AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, colocationID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation des invitations: %w", err)
	}
	defer rows.Close()

	var invitations []domain.ColocationInvitation
	for rows.Next() {
		var inv domain.ColocationInvitation
		err := rows.Scan(
			&inv.ID,
			&inv.ColocationID,
			&inv.InvitedBy,
			&inv.InvitedEmail,
			&inv.Status,
			&inv.ExpiresAt,
			&inv.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du scan de l'invitation: %w", err)
		}
		invitations = append(invitations, inv)
	}

	return invitations, nil
}

// DeleteInvitation deletes an invitation
func (r *ColocationRepository) DeleteInvitation(ctx context.Context, colocationID, invitationID string) error {
	query := `DELETE FROM colocation_invitations WHERE id = $1 AND colocation_id = $2`

	result, err := r.pool.Exec(ctx, query, invitationID, colocationID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression de l'invitation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("invitation introuvable")
	}

	return nil
}
