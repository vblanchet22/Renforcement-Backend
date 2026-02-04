package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

// CategoryService handles category business logic
type CategoryService struct {
	repo         *postgres.CategoryRepository
	colocationRepo *postgres.ColocationRepository
}

// NewCategoryService creates a new CategoryService
func NewCategoryService(repo *postgres.CategoryRepository, colocationRepo *postgres.ColocationRepository) *CategoryService {
	return &CategoryService{
		repo:         repo,
		colocationRepo: colocationRepo,
	}
}

// List returns all categories for a colocation (global + custom)
func (s *CategoryService) List(ctx context.Context, colocationID string) ([]domain.ExpenseCategory, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user is member of the colocation
	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	return s.repo.ListByColocation(ctx, colocationID)
}

// Create creates a new custom category for a colocation
func (s *CategoryService) Create(ctx context.Context, colocationID, name string, icon, color *string) (*domain.ExpenseCategory, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user is member of the colocation
	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	category := &domain.ExpenseCategory{
		Name:         name,
		Icon:         icon,
		Color:        color,
		ColocationID: &colocationID,
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("erreur lors de la creation de la categorie: %w", err)
	}

	return category, nil
}

// Update updates a custom category
func (s *CategoryService) Update(ctx context.Context, colocationID, categoryID string, name, icon, color *string) (*domain.ExpenseCategory, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user is member of the colocation
	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	// Get existing category
	category, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("categorie introuvable: %w", err)
	}

	// Check if it's a custom category for this colocation
	if category.ColocationID == nil || *category.ColocationID != colocationID {
		return nil, fmt.Errorf("impossible de modifier une categorie globale ou d'une autre colocation")
	}

	// Update fields
	if name != nil {
		category.Name = *name
	}
	if icon != nil {
		category.Icon = icon
	}
	if color != nil {
		category.Color = color
	}

	if err := s.repo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise a jour: %w", err)
	}

	return category, nil
}

// Delete deletes a custom category
func (s *CategoryService) Delete(ctx context.Context, colocationID, categoryID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Check if user is member of the colocation
	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	// Get existing category
	category, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("categorie introuvable: %w", err)
	}

	// Check if it's a custom category for this colocation
	if category.ColocationID == nil || *category.ColocationID != colocationID {
		return fmt.Errorf("impossible de supprimer une categorie globale ou d'une autre colocation")
	}

	return s.repo.Delete(ctx, categoryID)
}

// GetStats returns category statistics for a colocation
func (s *CategoryService) GetStats(ctx context.Context, colocationID string, startDate, endDate *time.Time) ([]domain.CategoryStat, float64, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Check if user is member of the colocation
	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, 0, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	return s.repo.GetStats(ctx, colocationID, startDate, endDate)
}
