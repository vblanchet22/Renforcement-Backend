package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vblanchet22/back_coloc/internal/algorithm"
	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

// BalanceService handles balance business logic
type BalanceService struct {
	repo           *postgres.BalanceRepository
	colocationRepo *postgres.ColocationRepository
}

// NewBalanceService creates a new BalanceService
func NewBalanceService(repo *postgres.BalanceRepository, colocationRepo *postgres.ColocationRepository) *BalanceService {
	return &BalanceService{
		repo:           repo,
		colocationRepo: colocationRepo,
	}
}

// GetBalances returns all balances and raw debts for a colocation
func (s *BalanceService) GetBalances(ctx context.Context, colocationID string) ([]domain.UserBalance, []domain.Debt, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Check membership
	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	balances, err := s.repo.GetUserBalances(ctx, colocationID)
	if err != nil {
		return nil, nil, fmt.Errorf("erreur lors du calcul des soldes: %w", err)
	}

	debts, err := s.repo.GetRawDebts(ctx, colocationID)
	if err != nil {
		return nil, nil, fmt.Errorf("erreur lors du calcul des dettes: %w", err)
	}

	return balances, debts, nil
}

// GetSimplifiedDebts returns simplified debts using the min-cash-flow algorithm
func (s *BalanceService) GetSimplifiedDebts(ctx context.Context, colocationID string) ([]domain.SimplifiedDebt, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check membership
	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	// Get member info and balances
	members, err := s.repo.GetMembersInfo(ctx, colocationID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation des membres: %w", err)
	}

	if len(members) == 0 {
		return nil, nil
	}

	balances, err := s.repo.GetUserBalances(ctx, colocationID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du calcul des soldes: %w", err)
	}

	// Build user index map
	userIndexMap := make(map[string]int)
	for i, m := range members {
		userIndexMap[m.UserID] = i
	}

	// Build net balances array
	netBalances := make([]float64, len(members))
	for _, b := range balances {
		if idx, ok := userIndexMap[b.UserID]; ok {
			netBalances[idx] = b.NetBalance
		}
	}

	// Run min-cash-flow algorithm
	edges := algorithm.MinCashFlow(netBalances)

	// Convert edges to SimplifiedDebt with user info
	var simplifiedDebts []domain.SimplifiedDebt
	for _, edge := range edges {
		from := members[edge.FromIndex]
		to := members[edge.ToIndex]

		simplifiedDebts = append(simplifiedDebts, domain.SimplifiedDebt{
			FromUserID:     from.UserID,
			FromUserNom:    from.Nom,
			FromUserPrenom: from.Prenom,
			FromAvatarURL:  from.AvatarURL,
			ToUserID:       to.UserID,
			ToUserNom:      to.Nom,
			ToUserPrenom:   to.Prenom,
			ToAvatarURL:    to.AvatarURL,
			Amount:         edge.Amount,
		})
	}

	return simplifiedDebts, nil
}

// GetBalanceHistory returns balance history for the current user
func (s *BalanceService) GetBalanceHistory(ctx context.Context, colocationID string, startDate, endDate *time.Time) ([]domain.BalanceHistoryEntry, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check membership
	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	return s.repo.GetBalanceHistory(ctx, colocationID, userID, startDate, endDate)
}
