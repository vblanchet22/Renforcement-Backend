package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

// Decision validation constants
const (
	minDecisionOptions = 2
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

// ensureMembership verifies user is a member and returns the userID
func (s *DecisionService) ensureMembership(ctx context.Context, colocationID string) (string, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return "", err
	}

	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return "", fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	return userID, nil
}

// Create creates a new decision
func (s *DecisionService) Create(ctx context.Context, colocationID, title string, description *string, options []string, deadline *time.Time, allowMultiple, isAnonymous bool) (*domain.Decision, error) {
	userID, err := s.ensureMembership(ctx, colocationID)
	if err != nil {
		return nil, err
	}

	if len(options) < minDecisionOptions {
		return nil, fmt.Errorf("au moins %d options sont requises", minDecisionOptions)
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
	userID, err := s.ensureMembership(ctx, colocationID)
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

	return decision, nil
}

// List lists decisions for a colocation
func (s *DecisionService) List(ctx context.Context, colocationID string, status *string, page, pageSize int) ([]domain.Decision, int, error) {
	userID, err := s.ensureMembership(ctx, colocationID)
	if err != nil {
		return nil, 0, err
	}

	page, pageSize = normalizePagination(page, pageSize)

	return s.repo.ListByColocation(ctx, colocationID, userID, status, page, pageSize)
}

// Update updates a decision (only if no votes yet)
func (s *DecisionService) Update(ctx context.Context, colocationID, decisionID string, title *string, description *string, options []string, deadline *time.Time, allowMultiple, isAnonymous *bool) (*domain.Decision, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	decision, err := s.getDecisionForOwner(ctx, colocationID, decisionID, userID)
	if err != nil {
		return nil, err
	}

	if err := s.validateNoVotesExist(ctx, decisionID); err != nil {
		return nil, err
	}

	s.applyDecisionUpdates(decision, title, description, options, deadline, allowMultiple, isAnonymous)

	if len(options) > 0 && len(options) < minDecisionOptions {
		return nil, fmt.Errorf("au moins %d options sont requises", minDecisionOptions)
	}

	if err := s.repo.Update(ctx, decision); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise a jour: %w", err)
	}

	return s.repo.GetByID(ctx, decisionID, userID)
}

// getDecisionForOwner retrieves a decision and verifies ownership
func (s *DecisionService) getDecisionForOwner(ctx context.Context, colocationID, decisionID, userID string) (*domain.Decision, error) {
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

	return decision, nil
}

// validateNoVotesExist checks that no votes have been cast on a decision
func (s *DecisionService) validateNoVotesExist(ctx context.Context, decisionID string) error {
	hasVotes, err := s.repo.HasVotes(ctx, decisionID)
	if err != nil {
		return fmt.Errorf("erreur lors de la verification des votes: %w", err)
	}
	if hasVotes {
		return fmt.Errorf("impossible de modifier une decision avec des votes existants")
	}
	return nil
}

// applyDecisionUpdates applies non-nil update fields to a decision
func (s *DecisionService) applyDecisionUpdates(decision *domain.Decision, title *string, description *string, options []string, deadline *time.Time, allowMultiple, isAnonymous *bool) {
	if title != nil {
		decision.Title = *title
	}
	if description != nil {
		decision.Description = description
	}
	if len(options) > 0 {
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
}

// Delete deletes a decision
func (s *DecisionService) Delete(ctx context.Context, colocationID, decisionID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	if _, err := s.getDecisionForOwner(ctx, colocationID, decisionID, userID); err != nil {
		return err
	}

	return s.repo.Delete(ctx, decisionID)
}

// Vote votes on a decision
func (s *DecisionService) Vote(ctx context.Context, colocationID, decisionID string, optionIndices []int) error {
	userID, err := s.ensureMembership(ctx, colocationID)
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

	if err := s.validateVote(decision, optionIndices); err != nil {
		return err
	}

	return s.repo.Vote(ctx, decisionID, userID, optionIndices)
}

// validateVote checks that a vote is valid for the given decision
func (s *DecisionService) validateVote(decision *domain.Decision, optionIndices []int) error {
	if decision.Status != domain.DecisionStatusOpen {
		return fmt.Errorf("cette decision est fermee")
	}

	if decision.Deadline != nil && decision.Deadline.Before(time.Now()) {
		return fmt.Errorf("la deadline est passee")
	}

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

	return nil
}

// Close closes a decision
func (s *DecisionService) Close(ctx context.Context, colocationID, decisionID string) (*domain.Decision, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	decision, err := s.getDecisionForOwner(ctx, colocationID, decisionID, userID)
	if err != nil {
		return nil, err
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
	userID, err := s.ensureMembership(ctx, colocationID)
	if err != nil {
		return nil, 0, 0, nil, err
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

	winningIdx := findWinningOption(results)

	return results, totalVotes, totalVoters, winningIdx, nil
}

// findWinningOption returns the index of the option with the most votes
func findWinningOption(results []domain.OptionResult) *int {
	var winningIdx *int
	maxVotes := 0

	for _, r := range results {
		if r.VoteCount > maxVotes {
			maxVotes = r.VoteCount
			idx := r.OptionIndex
			winningIdx = &idx
		}
	}

	return winningIdx
}
