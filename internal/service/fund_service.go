package service

import (
	"context"
	"fmt"

	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

// FundService handles fund business logic
type FundService struct {
	repo           *postgres.FundRepository
	colocationRepo *postgres.ColocationRepository
}

// NewFundService creates a new FundService
func NewFundService(repo *postgres.FundRepository, colocationRepo *postgres.ColocationRepository) *FundService {
	return &FundService{
		repo:           repo,
		colocationRepo: colocationRepo,
	}
}

// Create creates a new fund
func (s *FundService) Create(ctx context.Context, colocationID, name string, description *string, targetAmount *float64) (*domain.CommonFund, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	fund := &domain.CommonFund{
		ColocationID: colocationID,
		Name:         name,
		Description:  description,
		TargetAmount: targetAmount,
		CreatedBy:    userID,
	}

	if err := s.repo.Create(ctx, fund); err != nil {
		return nil, fmt.Errorf("erreur lors de la creation: %w", err)
	}

	return s.repo.GetByID(ctx, fund.ID)
}

// GetByID retrieves a fund by ID
func (s *FundService) GetByID(ctx context.Context, colocationID, fundID string) (*domain.CommonFund, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	fund, err := s.repo.GetByID(ctx, fundID)
	if err != nil {
		return nil, err
	}
	if fund == nil || fund.ColocationID != colocationID {
		return nil, fmt.Errorf("fonds introuvable")
	}

	return fund, nil
}

// List lists funds for a colocation
func (s *FundService) List(ctx context.Context, colocationID string, isActive *bool) ([]domain.CommonFund, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	return s.repo.ListByColocation(ctx, colocationID, isActive)
}

// Update updates a fund
func (s *FundService) Update(ctx context.Context, colocationID, fundID string, name *string, description *string, targetAmount *float64, isActive *bool) (*domain.CommonFund, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fund, err := s.repo.GetByID(ctx, fundID)
	if err != nil {
		return nil, err
	}
	if fund == nil || fund.ColocationID != colocationID {
		return nil, fmt.Errorf("fonds introuvable")
	}

	if fund.CreatedBy != userID {
		return nil, fmt.Errorf("seul le createur peut modifier ce fonds")
	}

	if name != nil {
		fund.Name = *name
	}
	if description != nil {
		fund.Description = description
	}
	if targetAmount != nil {
		fund.TargetAmount = targetAmount
	}
	if isActive != nil {
		fund.IsActive = *isActive
	}

	if err := s.repo.Update(ctx, fund); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise a jour: %w", err)
	}

	return s.repo.GetByID(ctx, fundID)
}

// Delete deletes a fund
func (s *FundService) Delete(ctx context.Context, colocationID, fundID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	fund, err := s.repo.GetByID(ctx, fundID)
	if err != nil {
		return err
	}
	if fund == nil || fund.ColocationID != colocationID {
		return fmt.Errorf("fonds introuvable")
	}

	if fund.CreatedBy != userID {
		return fmt.Errorf("seul le createur peut supprimer ce fonds")
	}

	return s.repo.Delete(ctx, fundID)
}

// AddContribution adds a contribution to a fund
func (s *FundService) AddContribution(ctx context.Context, colocationID, fundID string, amount float64, note *string) (*domain.FundContribution, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	fund, err := s.repo.GetByID(ctx, fundID)
	if err != nil {
		return nil, err
	}
	if fund == nil || fund.ColocationID != colocationID {
		return nil, fmt.Errorf("fonds introuvable")
	}

	if !fund.IsActive {
		return nil, fmt.Errorf("ce fonds n'est plus actif")
	}

	if amount <= 0 {
		return nil, fmt.Errorf("le montant doit etre positif")
	}

	contribution := &domain.FundContribution{
		FundID: fundID,
		UserID: userID,
		Amount: amount,
		Note:   note,
	}

	if err := s.repo.AddContribution(ctx, contribution); err != nil {
		return nil, fmt.Errorf("erreur lors de l'ajout de la contribution: %w", err)
	}

	return s.repo.GetContribution(ctx, contribution.ID)
}

// ListContributions lists contributions for a fund
func (s *FundService) ListContributions(ctx context.Context, colocationID, fundID string) ([]domain.FundContribution, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	return s.repo.ListContributions(ctx, fundID)
}

// DeleteContribution deletes a contribution
func (s *FundService) DeleteContribution(ctx context.Context, colocationID, fundID, contributionID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	contribution, err := s.repo.GetContribution(ctx, contributionID)
	if err != nil {
		return err
	}
	if contribution == nil || contribution.FundID != fundID {
		return fmt.Errorf("contribution introuvable")
	}

	if contribution.UserID != userID {
		return fmt.Errorf("seul le contributeur peut supprimer sa contribution")
	}

	return s.repo.DeleteContribution(ctx, contributionID, fundID, contribution.Amount)
}
