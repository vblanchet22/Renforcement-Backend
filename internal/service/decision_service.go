package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

// DecisionService handles decision business logic
type DecisionService struct {
	repo           *postgres.DecisionRepository
	colocationRepo *postgres.ColocationRepository
}

// NewDecisionService creates a new DecisionService
func NewDecisionService(repo *postgres.DecisionRepository, colocationRepo *postgres.ColocationRepository) *DecisionService {
	return &DecisionService{
		repo:           repo,
		colocationRepo: colocationRepo,
	}
}

// Create creates a new decision
func (s *DecisionService) Create(ctx context.Context, colocationID, title string, description *string, options []string, deadline *time.Time, allowMultiple, isAnonymous bool) (*domain.Decision, error) {
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

	if len(options) < 2 {
		return nil, fmt.Errorf("au moins 2 options sont requises")
	}

	decision := &domain.Decision{
		ColocationID:  colocationID,
		CreatedBy:     userID,
		Title:         title,
		Description:   description,
		Options:       options,
		Deadline:      deadline,
		AllowMultiple: allowMultiple,
		IsAnonymous:   isAnonymous,
	}

	if err := s.repo.Create(ctx, decision); err != nil {
		return nil, fmt.Errorf("erreur lors de la creation: %w", err)
	}

	return s.repo.GetByID(ctx, decision.ID, userID)
}

// GetByID retrieves a decision by ID
func (s *DecisionService) GetByID(ctx context.Context, colocationID, decisionID string) (*domain.Decision, error) {
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

	decision, err := s.repo.GetByID(ctx, decisionID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if decision == nil || decision.ColocationID != colocationID {
		return nil, fmt.Errorf("decision introuvable")
	}

	return decision, nil
}

// List lists decisions for a colocation
func (s *DecisionService) List(ctx context.Context, colocationID string, status *string, page, pageSize int) ([]domain.Decision, int, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, 0, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.repo.ListByColocation(ctx, colocationID, userID, status, page, pageSize)
}

// Update updates a decision (only if no votes yet)
func (s *DecisionService) Update(ctx context.Context, colocationID, decisionID string, title *string, description *string, options []string, deadline *time.Time, allowMultiple, isAnonymous *bool) (*domain.Decision, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	decision, err := s.repo.GetByID(ctx, decisionID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if decision == nil || decision.ColocationID != colocationID {
		return nil, fmt.Errorf("decision introuvable")
	}

	if decision.CreatedBy != userID {
		return nil, fmt.Errorf("seul le createur peut modifier cette decision")
	}

	// Check if votes exist
	hasVotes, err := s.repo.HasVotes(ctx, decisionID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification des votes: %w", err)
	}
	if hasVotes {
		return nil, fmt.Errorf("impossible de modifier une decision avec des votes existants")
	}

	if title != nil {
		decision.Title = *title
	}
	if description != nil {
		decision.Description = description
	}
	if len(options) > 0 {
		if len(options) < 2 {
			return nil, fmt.Errorf("au moins 2 options sont requises")
		}
		decision.Options = options
	}
	if deadline != nil {
		decision.Deadline = deadline
	}
	if allowMultiple != nil {
		decision.AllowMultiple = *allowMultiple
	}
	if isAnonymous != nil {
		decision.IsAnonymous = *isAnonymous
	}

	if err := s.repo.Update(ctx, decision); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise a jour: %w", err)
	}

	return s.repo.GetByID(ctx, decisionID, userID)
}

// Delete deletes a decision
func (s *DecisionService) Delete(ctx context.Context, colocationID, decisionID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	decision, err := s.repo.GetByID(ctx, decisionID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if decision == nil || decision.ColocationID != colocationID {
		return fmt.Errorf("decision introuvable")
	}

	if decision.CreatedBy != userID {
		return fmt.Errorf("seul le createur peut supprimer cette decision")
	}

	return s.repo.Delete(ctx, decisionID)
}

// Vote votes on a decision
func (s *DecisionService) Vote(ctx context.Context, colocationID, decisionID string, optionIndices []int) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	decision, err := s.repo.GetByID(ctx, decisionID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if decision == nil || decision.ColocationID != colocationID {
		return fmt.Errorf("decision introuvable")
	}

	if decision.Status != domain.DecisionStatusOpen {
		return fmt.Errorf("cette decision est fermee")
	}

	// Check deadline
	if decision.Deadline != nil && decision.Deadline.Before(time.Now()) {
		return fmt.Errorf("la deadline est passee")
	}

	// Validate option indices
	if len(optionIndices) == 0 {
		return fmt.Errorf("au moins un choix est requis")
	}

	if !decision.AllowMultiple && len(optionIndices) > 1 {
		return fmt.Errorf("un seul choix est autorise pour cette decision")
	}

	for _, idx := range optionIndices {
		if idx < 0 || idx >= len(decision.Options) {
			return fmt.Errorf("index d'option invalide: %d", idx)
		}
	}

	return s.repo.Vote(ctx, decisionID, userID, optionIndices)
}

// Close closes a decision
func (s *DecisionService) Close(ctx context.Context, colocationID, decisionID string) (*domain.Decision, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	decision, err := s.repo.GetByID(ctx, decisionID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if decision == nil || decision.ColocationID != colocationID {
		return nil, fmt.Errorf("decision introuvable")
	}

	if decision.CreatedBy != userID {
		return nil, fmt.Errorf("seul le createur peut fermer cette decision")
	}

	if decision.Status != domain.DecisionStatusOpen {
		return nil, fmt.Errorf("cette decision est deja fermee")
	}

	if err := s.repo.UpdateStatus(ctx, decisionID, domain.DecisionStatusClosed); err != nil {
		return nil, fmt.Errorf("erreur lors de la fermeture: %w", err)
	}

	return s.repo.GetByID(ctx, decisionID, userID)
}

// GetResults returns the results of a decision
func (s *DecisionService) GetResults(ctx context.Context, colocationID, decisionID string) ([]domain.OptionResult, int, int, *int, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, 0, 0, nil, err
	}

	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, 0, 0, nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, 0, 0, nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	decision, err := s.repo.GetByID(ctx, decisionID, userID)
	if err != nil {
		return nil, 0, 0, nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if decision == nil || decision.ColocationID != colocationID {
		return nil, 0, 0, nil, fmt.Errorf("decision introuvable")
	}

	results, totalVotes, totalVoters, err := s.repo.GetResults(ctx, decisionID)
	if err != nil {
		return nil, 0, 0, nil, fmt.Errorf("erreur lors du calcul des resultats: %w", err)
	}

	// Find winning option
	var winningIdx *int
	maxVotes := 0
	for _, r := range results {
		if r.VoteCount > maxVotes {
			maxVotes = r.VoteCount
			idx := r.OptionIndex
			winningIdx = &idx
		}
	}

	return results, totalVotes, totalVoters, winningIdx, nil
}
