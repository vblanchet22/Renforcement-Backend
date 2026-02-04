package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/constants"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

// ExpenseService handles expense business logic
type ExpenseService struct {
	repo           *postgres.ExpenseRepository
	colocationRepo *postgres.ColocationRepository
	categoryRepo   *postgres.CategoryRepository
}

// NewExpenseService creates a new ExpenseService
func NewExpenseService(repo *postgres.ExpenseRepository, colocationRepo *postgres.ColocationRepository, categoryRepo *postgres.CategoryRepository) *ExpenseService {
	return &ExpenseService{
		repo:           repo,
		colocationRepo: colocationRepo,
		categoryRepo:   categoryRepo,
	}
}

// CreateExpenseInput contains input for creating an expense
type CreateExpenseInput struct {
	ColocationID string
	Title        string
	Description  *string
	Amount       float64
	CategoryID   string
	SplitType    domain.SplitType
	Splits       []domain.ExpenseSplitInput
	ExpenseDate  time.Time
}

// Create creates a new expense
func (s *ExpenseService) Create(ctx context.Context, input CreateExpenseInput) (*domain.Expense, error) {
	userID, err := s.ensureMembership(ctx, input.ColocationID)
	if err != nil {
		return nil, err
	}

	if err := s.validateCategory(ctx, input.CategoryID, input.ColocationID); err != nil {
		return nil, err
	}

	splits, err := s.calculateSplits(ctx, input.ColocationID, input.Amount, input.SplitType, input.Splits, false)
	if err != nil {
		return nil, err
	}

	expense := &domain.Expense{
		ColocationID: input.ColocationID,
		PaidBy:       userID,
		CategoryID:   input.CategoryID,
		Title:        input.Title,
		Description:  input.Description,
		Amount:       input.Amount,
		SplitType:    input.SplitType,
		ExpenseDate:  input.ExpenseDate,
	}

	if err := s.repo.Create(ctx, expense, splits); err != nil {
		return nil, fmt.Errorf("erreur lors de la creation de la depense: %w", err)
	}

	return s.repo.GetByID(ctx, expense.ID)
}

// ensureMembership verifies user is a member and returns the userID
func (s *ExpenseService) ensureMembership(ctx context.Context, colocationID string) (string, error) {
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

// validateCategory checks if a category belongs to the colocation
func (s *ExpenseService) validateCategory(ctx context.Context, categoryID, colocationID string) error {
	belongs, err := s.categoryRepo.BelongsToColocation(ctx, categoryID, colocationID)
	if err != nil {
		return fmt.Errorf("erreur lors de la verification de la categorie: %w", err)
	}
	if !belongs {
		return fmt.Errorf("categorie invalide pour cette colocation")
	}
	return nil
}

// calculateSplits calculates expense splits based on the split type
// If percentageOnly is true, only percentages are calculated (for recurring expenses)
func (s *ExpenseService) calculateSplits(ctx context.Context, colocationID string, amount float64, splitType domain.SplitType, inputSplits []domain.ExpenseSplitInput, percentageOnly bool) ([]domain.ExpenseSplitInput, error) {
	members, err := s.colocationRepo.ListMembers(ctx, colocationID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation des membres: %w", err)
	}

	if len(members) == 0 {
		return nil, fmt.Errorf("aucun membre dans la colocation")
	}

	var splits []domain.ExpenseSplitInput

	switch splitType {
	case domain.SplitTypeEqual:
		splits = s.calculateEqualSplits(members, amount, percentageOnly)

	case domain.SplitTypePercentage:
		splits, err = s.calculatePercentageSplits(inputSplits, amount, percentageOnly)
		if err != nil {
			return nil, err
		}

	case domain.SplitTypeCustom:
		splits, err = s.calculateCustomSplits(inputSplits, amount, percentageOnly)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("type de partage invalide")
	}

	return splits, nil
}

// calculateEqualSplits divides amount equally among all members
func (s *ExpenseService) calculateEqualSplits(members []domain.ColocationMember, amount float64, percentageOnly bool) []domain.ExpenseSplitInput {
	memberCount := float64(len(members))
	percentage := constants.PercentageBase / memberCount

	var splits []domain.ExpenseSplitInput
	for _, m := range members {
		split := domain.ExpenseSplitInput{
			UserID:     m.UserID,
			Percentage: percentage,
		}
		if !percentageOnly {
			split.Amount = amount / memberCount
		}
		splits = append(splits, split)
	}
	return splits
}

// calculatePercentageSplits validates and calculates splits from percentages
func (s *ExpenseService) calculatePercentageSplits(inputSplits []domain.ExpenseSplitInput, amount float64, percentageOnly bool) ([]domain.ExpenseSplitInput, error) {
	if len(inputSplits) == 0 {
		return nil, fmt.Errorf("splits requis pour le mode pourcentage")
	}

	if err := validatePercentageTotal(inputSplits); err != nil {
		return nil, err
	}

	if percentageOnly {
		return inputSplits, nil
	}

	var splits []domain.ExpenseSplitInput
	for _, split := range inputSplits {
		splits = append(splits, domain.ExpenseSplitInput{
			UserID:     split.UserID,
			Amount:     amount * split.Percentage / constants.PercentageBase,
			Percentage: split.Percentage,
		})
	}
	return splits, nil
}

// calculateCustomSplits validates and calculates splits from custom amounts
func (s *ExpenseService) calculateCustomSplits(inputSplits []domain.ExpenseSplitInput, amount float64, percentageOnly bool) ([]domain.ExpenseSplitInput, error) {
	if len(inputSplits) == 0 {
		return nil, fmt.Errorf("splits requis pour le mode personnalise")
	}

	if err := validateAmountTotal(inputSplits, amount); err != nil {
		return nil, err
	}

	var splits []domain.ExpenseSplitInput
	for _, split := range inputSplits {
		s := domain.ExpenseSplitInput{
			UserID:     split.UserID,
			Percentage: (split.Amount / amount) * constants.PercentageBase,
		}
		if !percentageOnly {
			s.Amount = split.Amount
		}
		splits = append(splits, s)
	}
	return splits, nil
}

// validatePercentageTotal checks that percentages sum to 100%
func validatePercentageTotal(splits []domain.ExpenseSplitInput) error {
	var total float64
	for _, split := range splits {
		total += split.Percentage
	}
	if total < constants.PercentageMinBound || total > constants.PercentageMaxBound {
		return fmt.Errorf("les pourcentages doivent totaliser 100%% (actuellement: %.2f%%)", total)
	}
	return nil
}

// validateAmountTotal checks that amounts sum to the expected total
func validateAmountTotal(splits []domain.ExpenseSplitInput, expectedTotal float64) error {
	var total float64
	for _, split := range splits {
		total += split.Amount
	}
	if total < expectedTotal-constants.AmountTolerance || total > expectedTotal+constants.AmountTolerance {
		return fmt.Errorf("les montants doivent totaliser %.2f EUR (actuellement: %.2f EUR)", expectedTotal, total)
	}
	return nil
}

// normalizePagination ensures pagination values are within valid bounds
func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > constants.MaxPageSize {
		pageSize = constants.DefaultPageSize
	}
	return page, pageSize
}

// GetByID retrieves an expense by ID
func (s *ExpenseService) GetByID(ctx context.Context, colocationID, expenseID string) (*domain.Expense, error) {
	if _, err := s.ensureMembership(ctx, colocationID); err != nil {
		return nil, err
	}

	expense, err := s.repo.GetByID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if expense == nil || expense.ColocationID != colocationID {
		return nil, fmt.Errorf("depense introuvable")
	}

	return expense, nil
}

// ListExpensesInput contains filters for listing expenses
type ListExpensesInput struct {
	ColocationID string
	CategoryID   *string
	PaidBy       *string
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	PageSize     int
}

// List lists expenses for a colocation
func (s *ExpenseService) List(ctx context.Context, input ListExpensesInput) ([]domain.Expense, int, error) {
	if _, err := s.ensureMembership(ctx, input.ColocationID); err != nil {
		return nil, 0, err
	}

	input.Page, input.PageSize = normalizePagination(input.Page, input.PageSize)

	return s.repo.ListByColocation(ctx, input.ColocationID, input.CategoryID, input.PaidBy, input.StartDate, input.EndDate, input.Page, input.PageSize)
}

// UpdateExpenseInput contains input for updating an expense
type UpdateExpenseInput struct {
	ColocationID string
	ExpenseID    string
	Title        *string
	Description  *string
	Amount       *float64
	CategoryID   *string
	SplitType    *domain.SplitType
	Splits       []domain.ExpenseSplitInput
	ExpenseDate  *time.Time
}

// Update updates an expense
func (s *ExpenseService) Update(ctx context.Context, input UpdateExpenseInput) (*domain.Expense, error) {
	userID, err := s.ensureMembership(ctx, input.ColocationID)
	if err != nil {
		return nil, err
	}

	expense, err := s.repo.GetByID(ctx, input.ExpenseID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if expense == nil || expense.ColocationID != input.ColocationID {
		return nil, fmt.Errorf("depense introuvable")
	}

	if expense.PaidBy != userID {
		return nil, fmt.Errorf("seul le payeur peut modifier cette depense")
	}

	s.applyExpenseUpdates(expense, input)

	if input.CategoryID != nil {
		if err := s.validateCategory(ctx, *input.CategoryID, input.ColocationID); err != nil {
			return nil, err
		}
		expense.CategoryID = *input.CategoryID
	}

	splits, err := s.calculateSplits(ctx, input.ColocationID, expense.Amount, expense.SplitType, input.Splits, false)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, expense, splits); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise a jour: %w", err)
	}

	return s.repo.GetByID(ctx, expense.ID)
}

// applyExpenseUpdates applies non-nil update fields to an expense
func (s *ExpenseService) applyExpenseUpdates(expense *domain.Expense, input UpdateExpenseInput) {
	if input.Title != nil {
		expense.Title = *input.Title
	}
	if input.Description != nil {
		expense.Description = input.Description
	}
	if input.Amount != nil {
		expense.Amount = *input.Amount
	}
	if input.SplitType != nil {
		expense.SplitType = *input.SplitType
	}
	if input.ExpenseDate != nil {
		expense.ExpenseDate = *input.ExpenseDate
	}
}

// Delete deletes an expense
func (s *ExpenseService) Delete(ctx context.Context, colocationID, expenseID string) error {
	userID, err := s.ensureMembership(ctx, colocationID)
	if err != nil {
		return err
	}

	expense, err := s.repo.GetByID(ctx, expenseID)
	if err != nil {
		return fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if expense == nil || expense.ColocationID != colocationID {
		return fmt.Errorf("depense introuvable")
	}

	if expense.PaidBy != userID {
		return fmt.Errorf("seul le payeur peut supprimer cette depense")
	}

	return s.repo.Delete(ctx, expenseID)
}

// CreateRecurringInput contains input for creating a recurring expense
type CreateRecurringInput struct {
	ColocationID string
	Title        string
	Description  *string
	Amount       float64
	CategoryID   string
	SplitType    domain.SplitType
	Splits       []domain.ExpenseSplitInput
	Recurrence   domain.Recurrence
	StartDate    time.Time
	EndDate      *time.Time
}

// CreateRecurring creates a recurring expense template
func (s *ExpenseService) CreateRecurring(ctx context.Context, input CreateRecurringInput) (*domain.RecurringExpense, error) {
	userID, err := s.ensureMembership(ctx, input.ColocationID)
	if err != nil {
		return nil, err
	}

	if err := s.validateCategory(ctx, input.CategoryID, input.ColocationID); err != nil {
		return nil, err
	}

	splits, err := s.calculateSplits(ctx, input.ColocationID, input.Amount, input.SplitType, input.Splits, true)
	if err != nil {
		return nil, err
	}

	recurring := &domain.RecurringExpense{
		ColocationID: input.ColocationID,
		PaidBy:       userID,
		CategoryID:   input.CategoryID,
		Title:        input.Title,
		Description:  input.Description,
		Amount:       input.Amount,
		SplitType:    input.SplitType,
		Recurrence:   input.Recurrence,
		NextDueDate:  input.StartDate,
		EndDate:      input.EndDate,
		IsActive:     true,
	}

	if err := s.repo.CreateRecurring(ctx, recurring, splits); err != nil {
		return nil, fmt.Errorf("erreur lors de la creation: %w", err)
	}

	return s.repo.GetRecurringByID(ctx, recurring.ID)
}

// ListRecurring lists recurring expenses for a colocation
func (s *ExpenseService) ListRecurring(ctx context.Context, colocationID string) ([]domain.RecurringExpense, error) {
	if _, err := s.ensureMembership(ctx, colocationID); err != nil {
		return nil, err
	}

	return s.repo.ListRecurringByColocation(ctx, colocationID)
}

// UpdateRecurringInput contains input for updating a recurring expense
type UpdateRecurringInput struct {
	ColocationID string
	RecurringID  string
	Title        *string
	Description  *string
	Amount       *float64
	CategoryID   *string
	SplitType    *domain.SplitType
	Splits       []domain.ExpenseSplitInput
	Recurrence   *domain.Recurrence
	EndDate      *time.Time
	IsActive     *bool
}

// UpdateRecurring updates a recurring expense
func (s *ExpenseService) UpdateRecurring(ctx context.Context, input UpdateRecurringInput) (*domain.RecurringExpense, error) {
	userID, err := s.ensureMembership(ctx, input.ColocationID)
	if err != nil {
		return nil, err
	}

	recurring, err := s.repo.GetRecurringByID(ctx, input.RecurringID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if recurring == nil || recurring.ColocationID != input.ColocationID {
		return nil, fmt.Errorf("depense recurrente introuvable")
	}

	if recurring.PaidBy != userID {
		return nil, fmt.Errorf("seul le payeur peut modifier cette depense")
	}

	s.applyRecurringUpdates(recurring, input)

	if input.CategoryID != nil {
		if err := s.validateCategory(ctx, *input.CategoryID, input.ColocationID); err != nil {
			return nil, err
		}
		recurring.CategoryID = *input.CategoryID
	}

	var splits []domain.ExpenseSplitInput
	if len(input.Splits) > 0 || input.SplitType != nil {
		splits, err = s.calculateSplits(ctx, input.ColocationID, recurring.Amount, recurring.SplitType, input.Splits, true)
		if err != nil {
			return nil, err
		}
	}

	if err := s.repo.UpdateRecurring(ctx, recurring, splits); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise a jour: %w", err)
	}

	return s.repo.GetRecurringByID(ctx, recurring.ID)
}

// applyRecurringUpdates applies non-nil update fields to a recurring expense
func (s *ExpenseService) applyRecurringUpdates(recurring *domain.RecurringExpense, input UpdateRecurringInput) {
	if input.Title != nil {
		recurring.Title = *input.Title
	}
	if input.Description != nil {
		recurring.Description = input.Description
	}
	if input.Amount != nil {
		recurring.Amount = *input.Amount
	}
	if input.SplitType != nil {
		recurring.SplitType = *input.SplitType
	}
	if input.Recurrence != nil {
		recurring.Recurrence = *input.Recurrence
	}
	if input.EndDate != nil {
		recurring.EndDate = input.EndDate
	}
	if input.IsActive != nil {
		recurring.IsActive = *input.IsActive
	}
}

// DeleteRecurring deletes a recurring expense
func (s *ExpenseService) DeleteRecurring(ctx context.Context, colocationID, recurringID string) error {
	userID, err := s.ensureMembership(ctx, colocationID)
	if err != nil {
		return err
	}

	recurring, err := s.repo.GetRecurringByID(ctx, recurringID)
	if err != nil {
		return fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if recurring == nil || recurring.ColocationID != colocationID {
		return fmt.Errorf("depense recurrente introuvable")
	}

	if recurring.PaidBy != userID {
		return fmt.Errorf("seul le payeur peut supprimer cette depense")
	}

	return s.repo.DeleteRecurring(ctx, recurringID)
}

// Forecast limits
const (
	maxForecastMonths = 12
)

// GetForecast returns expense forecast for a colocation
func (s *ExpenseService) GetForecast(ctx context.Context, colocationID string, monthsAhead int) ([]domain.MonthlyForecast, error) {
	if _, err := s.ensureMembership(ctx, colocationID); err != nil {
		return nil, err
	}

	if monthsAhead < 1 || monthsAhead > maxForecastMonths {
		monthsAhead = constants.DefaultForecastMonths
	}

	return s.repo.GetForecastData(ctx, colocationID, monthsAhead)
}

// ProcessDueRecurringExpenses processes all recurring expenses that are due
func (s *ExpenseService) ProcessDueRecurringExpenses(ctx context.Context) error {
	now := time.Now()

	recurrings, err := s.repo.GetActiveRecurringDue(ctx, now)
	if err != nil {
		return fmt.Errorf("erreur lors de la recuperation des recurrences: %w", err)
	}

	for _, re := range recurrings {
		if _, err := s.repo.CreateFromRecurring(ctx, &re); err != nil {
			continue
		}

		nextDue := calculateNextDueDate(re.NextDueDate, re.Recurrence)
		if err := s.repo.UpdateNextDueDate(ctx, re.ID, nextDue); err != nil {
			continue
		}
	}

	return nil
}

// calculateNextDueDate calculates the next due date based on recurrence
func calculateNextDueDate(current time.Time, recurrence domain.Recurrence) time.Time {
	switch recurrence {
	case domain.RecurrenceDaily:
		return current.AddDate(0, 0, 1)
	case domain.RecurrenceWeekly:
		return current.AddDate(0, 0, 7)
	case domain.RecurrenceMonthly:
		return current.AddDate(0, 1, 0)
	case domain.RecurrenceYearly:
		return current.AddDate(1, 0, 0)
	default:
		return current.AddDate(0, 1, 0)
	}
}
