package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vblanchet22/back_coloc/internal/auth"
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
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check membership
	isMember, err := s.colocationRepo.IsMember(ctx, input.ColocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	// Validate category
	belongs, err := s.categoryRepo.BelongsToColocation(ctx, input.CategoryID, input.ColocationID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification de la categorie: %w", err)
	}
	if !belongs {
		return nil, fmt.Errorf("categorie invalide pour cette colocation")
	}

	// Calculate splits based on split type
	splits, err := s.calculateSplits(ctx, input.ColocationID, input.Amount, input.SplitType, input.Splits)
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

	// Reload with all details
	return s.repo.GetByID(ctx, expense.ID)
}

// calculateSplits calculates expense splits based on the split type
func (s *ExpenseService) calculateSplits(ctx context.Context, colocationID string, amount float64, splitType domain.SplitType, inputSplits []domain.ExpenseSplitInput) ([]domain.ExpenseSplitInput, error) {
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
		// Equal split among all members
		splitAmount := amount / float64(len(members))
		percentage := 100.0 / float64(len(members))
		for _, m := range members {
			splits = append(splits, domain.ExpenseSplitInput{
				UserID:     m.UserID,
				Amount:     splitAmount,
				Percentage: percentage,
			})
		}

	case domain.SplitTypePercentage:
		// Custom percentage per member
		if len(inputSplits) == 0 {
			return nil, fmt.Errorf("splits requis pour le mode pourcentage")
		}

		// Validate percentages sum to 100
		var totalPercentage float64
		for _, split := range inputSplits {
			totalPercentage += split.Percentage
		}
		if totalPercentage < 99.99 || totalPercentage > 100.01 {
			return nil, fmt.Errorf("les pourcentages doivent totaliser 100%% (actuellement: %.2f%%)", totalPercentage)
		}

		// Calculate amounts from percentages
		for _, split := range inputSplits {
			splits = append(splits, domain.ExpenseSplitInput{
				UserID:     split.UserID,
				Amount:     amount * split.Percentage / 100,
				Percentage: split.Percentage,
			})
		}

	case domain.SplitTypeCustom:
		// Fixed amount per member
		if len(inputSplits) == 0 {
			return nil, fmt.Errorf("splits requis pour le mode personnalise")
		}

		// Validate amounts sum to total
		var totalAmount float64
		for _, split := range inputSplits {
			totalAmount += split.Amount
		}
		if totalAmount < amount-0.01 || totalAmount > amount+0.01 {
			return nil, fmt.Errorf("les montants doivent totaliser %.2f EUR (actuellement: %.2f EUR)", amount, totalAmount)
		}

		// Calculate percentages from amounts
		for _, split := range inputSplits {
			splits = append(splits, domain.ExpenseSplitInput{
				UserID:     split.UserID,
				Amount:     split.Amount,
				Percentage: (split.Amount / amount) * 100,
			})
		}

	default:
		return nil, fmt.Errorf("type de partage invalide")
	}

	return splits, nil
}

// GetByID retrieves an expense by ID
func (s *ExpenseService) GetByID(ctx context.Context, colocationID, expenseID string) (*domain.Expense, error) {
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

	expense, err := s.repo.GetByID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if expense == nil {
		return nil, fmt.Errorf("depense introuvable")
	}

	if expense.ColocationID != colocationID {
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
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Check membership
	isMember, err := s.colocationRepo.IsMember(ctx, input.ColocationID, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, 0, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	// Default pagination
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

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
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check membership
	isMember, err := s.colocationRepo.IsMember(ctx, input.ColocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	// Get existing expense
	expense, err := s.repo.GetByID(ctx, input.ExpenseID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if expense == nil || expense.ColocationID != input.ColocationID {
		return nil, fmt.Errorf("depense introuvable")
	}

	// Only payer can update
	if expense.PaidBy != userID {
		return nil, fmt.Errorf("seul le payeur peut modifier cette depense")
	}

	// Update fields
	if input.Title != nil {
		expense.Title = *input.Title
	}
	if input.Description != nil {
		expense.Description = input.Description
	}
	if input.Amount != nil {
		expense.Amount = *input.Amount
	}
	if input.CategoryID != nil {
		belongs, err := s.categoryRepo.BelongsToColocation(ctx, *input.CategoryID, input.ColocationID)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la verification de la categorie: %w", err)
		}
		if !belongs {
			return nil, fmt.Errorf("categorie invalide pour cette colocation")
		}
		expense.CategoryID = *input.CategoryID
	}
	if input.SplitType != nil {
		expense.SplitType = *input.SplitType
	}
	if input.ExpenseDate != nil {
		expense.ExpenseDate = *input.ExpenseDate
	}

	// Recalculate splits
	splits, err := s.calculateSplits(ctx, input.ColocationID, expense.Amount, expense.SplitType, input.Splits)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, expense, splits); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise a jour: %w", err)
	}

	return s.repo.GetByID(ctx, expense.ID)
}

// Delete deletes an expense
func (s *ExpenseService) Delete(ctx context.Context, colocationID, expenseID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Check membership
	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	// Get existing expense
	expense, err := s.repo.GetByID(ctx, expenseID)
	if err != nil {
		return fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if expense == nil || expense.ColocationID != colocationID {
		return fmt.Errorf("depense introuvable")
	}

	// Only payer can delete
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
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check membership
	isMember, err := s.colocationRepo.IsMember(ctx, input.ColocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	// Validate category
	belongs, err := s.categoryRepo.BelongsToColocation(ctx, input.CategoryID, input.ColocationID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification de la categorie: %w", err)
	}
	if !belongs {
		return nil, fmt.Errorf("categorie invalide pour cette colocation")
	}

	// Calculate splits (for recurring, we store percentages)
	splits, err := s.calculateRecurringSplits(ctx, input.ColocationID, input.Amount, input.SplitType, input.Splits)
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

// calculateRecurringSplits calculates percentage splits for recurring expenses
func (s *ExpenseService) calculateRecurringSplits(ctx context.Context, colocationID string, amount float64, splitType domain.SplitType, inputSplits []domain.ExpenseSplitInput) ([]domain.ExpenseSplitInput, error) {
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
		percentage := 100.0 / float64(len(members))
		for _, m := range members {
			splits = append(splits, domain.ExpenseSplitInput{
				UserID:     m.UserID,
				Percentage: percentage,
			})
		}

	case domain.SplitTypePercentage:
		if len(inputSplits) == 0 {
			return nil, fmt.Errorf("splits requis pour le mode pourcentage")
		}
		var total float64
		for _, split := range inputSplits {
			total += split.Percentage
		}
		if total < 99.99 || total > 100.01 {
			return nil, fmt.Errorf("les pourcentages doivent totaliser 100%%")
		}
		splits = inputSplits

	case domain.SplitTypeCustom:
		if len(inputSplits) == 0 {
			return nil, fmt.Errorf("splits requis pour le mode personnalise")
		}
		var total float64
		for _, split := range inputSplits {
			total += split.Amount
		}
		if total < amount-0.01 || total > amount+0.01 {
			return nil, fmt.Errorf("les montants doivent totaliser %.2f EUR", amount)
		}
		// Convert to percentages
		for _, split := range inputSplits {
			splits = append(splits, domain.ExpenseSplitInput{
				UserID:     split.UserID,
				Percentage: (split.Amount / amount) * 100,
			})
		}

	default:
		return nil, fmt.Errorf("type de partage invalide")
	}

	return splits, nil
}

// ListRecurring lists recurring expenses for a colocation
func (s *ExpenseService) ListRecurring(ctx context.Context, colocationID string) ([]domain.RecurringExpense, error) {
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
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check membership
	isMember, err := s.colocationRepo.IsMember(ctx, input.ColocationID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	// Get existing
	recurring, err := s.repo.GetRecurringByID(ctx, input.RecurringID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if recurring == nil || recurring.ColocationID != input.ColocationID {
		return nil, fmt.Errorf("depense recurrente introuvable")
	}

	// Only payer can update
	if recurring.PaidBy != userID {
		return nil, fmt.Errorf("seul le payeur peut modifier cette depense")
	}

	// Update fields
	if input.Title != nil {
		recurring.Title = *input.Title
	}
	if input.Description != nil {
		recurring.Description = input.Description
	}
	if input.Amount != nil {
		recurring.Amount = *input.Amount
	}
	if input.CategoryID != nil {
		belongs, err := s.categoryRepo.BelongsToColocation(ctx, *input.CategoryID, input.ColocationID)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la verification: %w", err)
		}
		if !belongs {
			return nil, fmt.Errorf("categorie invalide")
		}
		recurring.CategoryID = *input.CategoryID
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

	// Recalculate splits if needed
	var splits []domain.ExpenseSplitInput
	if len(input.Splits) > 0 || input.SplitType != nil {
		splits, err = s.calculateRecurringSplits(ctx, input.ColocationID, recurring.Amount, recurring.SplitType, input.Splits)
		if err != nil {
			return nil, err
		}
	}

	if err := s.repo.UpdateRecurring(ctx, recurring, splits); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise a jour: %w", err)
	}

	return s.repo.GetRecurringByID(ctx, recurring.ID)
}

// DeleteRecurring deletes a recurring expense
func (s *ExpenseService) DeleteRecurring(ctx context.Context, colocationID, recurringID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Check membership
	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	// Get existing
	recurring, err := s.repo.GetRecurringByID(ctx, recurringID)
	if err != nil {
		return fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if recurring == nil || recurring.ColocationID != colocationID {
		return fmt.Errorf("depense recurrente introuvable")
	}

	// Only payer can delete
	if recurring.PaidBy != userID {
		return fmt.Errorf("seul le payeur peut supprimer cette depense")
	}

	return s.repo.DeleteRecurring(ctx, recurringID)
}

// GetForecast returns expense forecast for a colocation
func (s *ExpenseService) GetForecast(ctx context.Context, colocationID string, monthsAhead int) ([]domain.MonthlyForecast, error) {
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

	if monthsAhead < 1 || monthsAhead > 12 {
		monthsAhead = 3
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
		// Create expense from recurring
		_, err := s.repo.CreateFromRecurring(ctx, &re)
		if err != nil {
			// Log error but continue processing
			continue
		}

		// Calculate next due date
		nextDue := calculateNextDueDate(re.NextDueDate, re.Recurrence)

		// Update next due date
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
