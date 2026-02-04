package service

import (
	"context"
	"fmt"

	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

// UserService handles user business logic
type UserService struct {
	repo *postgres.AuthRepository
}

// NewUserService creates a new UserService
func NewUserService(repo *postgres.AuthRepository) *UserService {
	return &UserService{repo: repo}
}

// GetCurrentUser retrieves the current authenticated user
func (s *UserService) GetCurrentUser(ctx context.Context) (*domain.User, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation de l'utilisateur: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("utilisateur introuvable")
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation de l'utilisateur: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("utilisateur introuvable")
	}

	return user, nil
}

// UpdateUserInput contains fields to update
type UpdateUserInput struct {
	Nom             *string
	Prenom          *string
	Telephone       *string
	AvatarURL       *string
	CurrentPassword *string
	NewPassword     *string
}

// UpdateCurrentUser updates the current user's profile
func (s *UserService) UpdateCurrentUser(ctx context.Context, input UpdateUserInput) (*domain.User, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation de l'utilisateur: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("utilisateur introuvable")
	}

	// Update fields if provided
	if input.Nom != nil {
		user.Nom = *input.Nom
	}
	if input.Prenom != nil {
		user.Prenom = *input.Prenom
	}
	if input.Telephone != nil {
		user.Telephone = input.Telephone
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}

	// Handle password change
	if input.NewPassword != nil && *input.NewPassword != "" {
		if input.CurrentPassword == nil || *input.CurrentPassword == "" {
			return nil, fmt.Errorf("mot de passe actuel requis pour changer le mot de passe")
		}

		// Verify current password
		if user.PasswordHash == nil || !auth.CheckPassword(*input.CurrentPassword, *user.PasswordHash) {
			return nil, fmt.Errorf("mot de passe actuel incorrect")
		}

		// Hash new password
		hash, err := auth.HashPassword(*input.NewPassword)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = &hash
	}

	// Update in database
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteCurrentUser deactivates the current user's account
func (s *UserService) DeleteCurrentUser(ctx context.Context) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Soft delete: set is_active to false
	if err := s.repo.DeactivateUser(ctx, userID); err != nil {
		return fmt.Errorf("erreur lors de la suppression du compte: %w", err)
	}

	// Delete all refresh tokens
	if err := s.repo.DeleteUserRefreshTokens(ctx, userID); err != nil {
		return fmt.Errorf("erreur lors de la suppression des sessions: %w", err)
	}

	return nil
}
