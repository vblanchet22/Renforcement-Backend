package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/domain"
)

// ExpenseRepository handles expense database operations
type ExpenseRepository struct {
	pool *pgxpool.Pool
}

// NewExpenseRepository creates a new ExpenseRepository
func NewExpenseRepository(pool *pgxpool.Pool) *ExpenseRepository {
	return &ExpenseRepository{pool: pool}
}

// Create creates a new expense with its splits
func (r *ExpenseRepository) Create(ctx context.Context, expense *domain.Expense, splits []domain.ExpenseSplitInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erreur lors du demarrage de la transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert expense
	query := `
		INSERT INTO expenses (colocation_id, paid_by, category_id, title, description, amount, split_type, expense_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	err = tx.QueryRow(ctx, query,
		expense.ColocationID,
		expense.PaidBy,
		expense.CategoryID,
		expense.Title,
		expense.Description,
		expense.Amount,
		expense.SplitType,
		expense.ExpenseDate,
	).Scan(&expense.ID, &expense.CreatedAt)

	if err != nil {
		return fmt.Errorf("erreur lors de la creation de la depense: %w", err)
	}

	// Insert splits
	for _, split := range splits {
		splitQuery := `
			INSERT INTO expense_splits (expense_id, user_id, amount, percentage)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.Exec(ctx, splitQuery, expense.ID, split.UserID, split.Amount, split.Percentage)
		if err != nil {
			return fmt.Errorf("erreur lors de la creation du split: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetByID retrieves an expense by ID with all its details
func (r *ExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	query := `
		SELECT e.id, e.colocation_id, e.paid_by, e.category_id, e.title, e.description,
		       e.amount, e.split_type, e.expense_date, e.recurring_id, e.created_at,
		       u.nom, u.prenom, c.name
		FROM expenses e
		INNER JOIN users u ON e.paid_by = u.id
		INNER JOIN expense_categories c ON e.category_id = c.id
		WHERE e.id = $1
	`

	var expense domain.Expense
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&expense.ID,
		&expense.ColocationID,
		&expense.PaidBy,
		&expense.CategoryID,
		&expense.Title,
		&expense.Description,
		&expense.Amount,
		&expense.SplitType,
		&expense.ExpenseDate,
		&expense.RecurringID,
		&expense.CreatedAt,
		&expense.PaidByNom,
		&expense.PaidByPrenom,
		&expense.CategoryName,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation de la depense: %w", err)
	}

	// Get splits
	splits, err := r.GetSplits(ctx, expense.ID)
	if err != nil {
		return nil, err
	}
	expense.Splits = splits

	return &expense, nil
}

// GetSplits retrieves all splits for an expense
func (r *ExpenseRepository) GetSplits(ctx context.Context, expenseID string) ([]domain.ExpenseSplit, error) {
	query := `
		SELECT es.id, es.expense_id, es.user_id, es.amount, es.percentage, es.is_settled,
		       u.nom, u.prenom
		FROM expense_splits es
		INNER JOIN users u ON es.user_id = u.id
		WHERE es.expense_id = $1
	`

	rows, err := r.pool.Query(ctx, query, expenseID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation des splits: %w", err)
	}
	defer rows.Close()

	var splits []domain.ExpenseSplit
	for rows.Next() {
		var s domain.ExpenseSplit
		if err := rows.Scan(
			&s.ID, &s.ExpenseID, &s.UserID, &s.Amount, &s.Percentage, &s.IsSettled,
			&s.UserNom, &s.UserPrenom,
		); err != nil {
			return nil, fmt.Errorf("erreur lors du scan du split: %w", err)
		}
		splits = append(splits, s)
	}

	return splits, rows.Err()
}

// ListByColocation lists expenses for a colocation with filters
func (r *ExpenseRepository) ListByColocation(ctx context.Context, colocationID string, categoryID, paidBy *string, startDate, endDate *time.Time, page, pageSize int) ([]domain.Expense, int, error) {
	// Base query
	baseQuery := `
		FROM expenses e
		INNER JOIN users u ON e.paid_by = u.id
		INNER JOIN expense_categories c ON e.category_id = c.id
		WHERE e.colocation_id = $1
	`

	args := []interface{}{colocationID}
	argIndex := 2

	if categoryID != nil {
		baseQuery += fmt.Sprintf(" AND e.category_id = $%d", argIndex)
		args = append(args, *categoryID)
		argIndex++
	}

	if paidBy != nil {
		baseQuery += fmt.Sprintf(" AND e.paid_by = $%d", argIndex)
		args = append(args, *paidBy)
		argIndex++
	}

	if startDate != nil {
		baseQuery += fmt.Sprintf(" AND e.expense_date >= $%d", argIndex)
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil {
		baseQuery += fmt.Sprintf(" AND e.expense_date <= $%d", argIndex)
		args = append(args, *endDate)
		argIndex++
	}

	// Count total
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors du comptage des depenses: %w", err)
	}

	// Get expenses
	selectQuery := `
		SELECT e.id, e.colocation_id, e.paid_by, e.category_id, e.title, e.description,
		       e.amount, e.split_type, e.expense_date, e.recurring_id, e.created_at,
		       u.nom, u.prenom, c.name
	` + baseQuery + fmt.Sprintf(" ORDER BY e.expense_date DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors de la recuperation des depenses: %w", err)
	}
	defer rows.Close()

	var expenses []domain.Expense
	for rows.Next() {
		var e domain.Expense
		if err := rows.Scan(
			&e.ID, &e.ColocationID, &e.PaidBy, &e.CategoryID, &e.Title, &e.Description,
			&e.Amount, &e.SplitType, &e.ExpenseDate, &e.RecurringID, &e.CreatedAt,
			&e.PaidByNom, &e.PaidByPrenom, &e.CategoryName,
		); err != nil {
			return nil, 0, fmt.Errorf("erreur lors du scan de la depense: %w", err)
		}

		// Get splits for each expense
		splits, err := r.GetSplits(ctx, e.ID)
		if err != nil {
			return nil, 0, err
		}
		e.Splits = splits

		expenses = append(expenses, e)
	}

	return expenses, totalCount, rows.Err()
}

// Update updates an expense and its splits
func (r *ExpenseRepository) Update(ctx context.Context, expense *domain.Expense, splits []domain.ExpenseSplitInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erreur lors du demarrage de la transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update expense
	query := `
		UPDATE expenses
		SET title = $1, description = $2, amount = $3, category_id = $4, split_type = $5, expense_date = $6
		WHERE id = $7
	`

	_, err = tx.Exec(ctx, query,
		expense.Title,
		expense.Description,
		expense.Amount,
		expense.CategoryID,
		expense.SplitType,
		expense.ExpenseDate,
		expense.ID,
	)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise a jour de la depense: %w", err)
	}

	// Delete old splits and insert new ones
	_, err = tx.Exec(ctx, "DELETE FROM expense_splits WHERE expense_id = $1", expense.ID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression des splits: %w", err)
	}

	for _, split := range splits {
		splitQuery := `
			INSERT INTO expense_splits (expense_id, user_id, amount, percentage)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.Exec(ctx, splitQuery, expense.ID, split.UserID, split.Amount, split.Percentage)
		if err != nil {
			return fmt.Errorf("erreur lors de la creation du split: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// Delete deletes an expense and its splits
func (r *ExpenseRepository) Delete(ctx context.Context, id string) error {
	// Splits are deleted by CASCADE
	query := `DELETE FROM expenses WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression de la depense: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("depense introuvable")
	}

	return nil
}

// BelongsToColocation checks if an expense belongs to a colocation
func (r *ExpenseRepository) BelongsToColocation(ctx context.Context, expenseID, colocationID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM expenses WHERE id = $1 AND colocation_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, expenseID, colocationID).Scan(&exists)
	return exists, err
}

// CreateRecurring creates a recurring expense template
func (r *ExpenseRepository) CreateRecurring(ctx context.Context, recurring *domain.RecurringExpense, splits []domain.ExpenseSplitInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erreur lors du demarrage de la transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO recurring_expenses (colocation_id, paid_by, category_id, title, description, amount, split_type, recurrence, next_due_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, is_active, created_at
	`

	err = tx.QueryRow(ctx, query,
		recurring.ColocationID,
		recurring.PaidBy,
		recurring.CategoryID,
		recurring.Title,
		recurring.Description,
		recurring.Amount,
		recurring.SplitType,
		recurring.Recurrence,
		recurring.NextDueDate,
		recurring.EndDate,
	).Scan(&recurring.ID, &recurring.IsActive, &recurring.CreatedAt)

	if err != nil {
		return fmt.Errorf("erreur lors de la creation de la depense recurrente: %w", err)
	}

	// Insert splits (percentages only for recurring)
	for _, split := range splits {
		splitQuery := `
			INSERT INTO recurring_expense_splits (recurring_id, user_id, percentage)
			VALUES ($1, $2, $3)
		`
		_, err = tx.Exec(ctx, splitQuery, recurring.ID, split.UserID, split.Percentage)
		if err != nil {
			return fmt.Errorf("erreur lors de la creation du split recurrent: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetRecurringByID retrieves a recurring expense by ID
func (r *ExpenseRepository) GetRecurringByID(ctx context.Context, id string) (*domain.RecurringExpense, error) {
	query := `
		SELECT re.id, re.colocation_id, re.paid_by, re.category_id, re.title, re.description,
		       re.amount, re.split_type, re.recurrence, re.next_due_date, re.end_date, re.is_active, re.created_at,
		       u.nom, u.prenom, c.name
		FROM recurring_expenses re
		INNER JOIN users u ON re.paid_by = u.id
		INNER JOIN expense_categories c ON re.category_id = c.id
		WHERE re.id = $1
	`

	var re domain.RecurringExpense
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&re.ID, &re.ColocationID, &re.PaidBy, &re.CategoryID, &re.Title, &re.Description,
		&re.Amount, &re.SplitType, &re.Recurrence, &re.NextDueDate, &re.EndDate, &re.IsActive, &re.CreatedAt,
		&re.PaidByNom, &re.PaidByPrenom, &re.CategoryName,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation de la depense recurrente: %w", err)
	}

	// Get splits
	splits, err := r.GetRecurringSplits(ctx, re.ID)
	if err != nil {
		return nil, err
	}
	re.Splits = splits

	return &re, nil
}

// GetRecurringSplits retrieves all splits for a recurring expense
func (r *ExpenseRepository) GetRecurringSplits(ctx context.Context, recurringID string) ([]domain.RecurringExpenseSplit, error) {
	query := `
		SELECT res.id, res.recurring_id, res.user_id, res.percentage, u.nom, u.prenom
		FROM recurring_expense_splits res
		INNER JOIN users u ON res.user_id = u.id
		WHERE res.recurring_id = $1
	`

	rows, err := r.pool.Query(ctx, query, recurringID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation des splits: %w", err)
	}
	defer rows.Close()

	var splits []domain.RecurringExpenseSplit
	for rows.Next() {
		var s domain.RecurringExpenseSplit
		if err := rows.Scan(&s.ID, &s.RecurringID, &s.UserID, &s.Percentage, &s.UserNom, &s.UserPrenom); err != nil {
			return nil, fmt.Errorf("erreur lors du scan du split: %w", err)
		}
		splits = append(splits, s)
	}

	return splits, rows.Err()
}

// ListRecurringByColocation lists recurring expenses for a colocation
func (r *ExpenseRepository) ListRecurringByColocation(ctx context.Context, colocationID string) ([]domain.RecurringExpense, error) {
	query := `
		SELECT re.id, re.colocation_id, re.paid_by, re.category_id, re.title, re.description,
		       re.amount, re.split_type, re.recurrence, re.next_due_date, re.end_date, re.is_active, re.created_at,
		       u.nom, u.prenom, c.name
		FROM recurring_expenses re
		INNER JOIN users u ON re.paid_by = u.id
		INNER JOIN expense_categories c ON re.category_id = c.id
		WHERE re.colocation_id = $1
		ORDER BY re.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, colocationID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation des depenses recurrentes: %w", err)
	}
	defer rows.Close()

	var recurrings []domain.RecurringExpense
	for rows.Next() {
		var re domain.RecurringExpense
		if err := rows.Scan(
			&re.ID, &re.ColocationID, &re.PaidBy, &re.CategoryID, &re.Title, &re.Description,
			&re.Amount, &re.SplitType, &re.Recurrence, &re.NextDueDate, &re.EndDate, &re.IsActive, &re.CreatedAt,
			&re.PaidByNom, &re.PaidByPrenom, &re.CategoryName,
		); err != nil {
			return nil, fmt.Errorf("erreur lors du scan: %w", err)
		}

		// Get splits
		splits, err := r.GetRecurringSplits(ctx, re.ID)
		if err != nil {
			return nil, err
		}
		re.Splits = splits

		recurrings = append(recurrings, re)
	}

	return recurrings, rows.Err()
}

// UpdateRecurring updates a recurring expense
func (r *ExpenseRepository) UpdateRecurring(ctx context.Context, recurring *domain.RecurringExpense, splits []domain.ExpenseSplitInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erreur lors du demarrage de la transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE recurring_expenses
		SET title = $1, description = $2, amount = $3, category_id = $4, split_type = $5, recurrence = $6, end_date = $7, is_active = $8
		WHERE id = $9
	`

	_, err = tx.Exec(ctx, query,
		recurring.Title,
		recurring.Description,
		recurring.Amount,
		recurring.CategoryID,
		recurring.SplitType,
		recurring.Recurrence,
		recurring.EndDate,
		recurring.IsActive,
		recurring.ID,
	)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise a jour: %w", err)
	}

	// Update splits if provided
	if len(splits) > 0 {
		_, err = tx.Exec(ctx, "DELETE FROM recurring_expense_splits WHERE recurring_id = $1", recurring.ID)
		if err != nil {
			return fmt.Errorf("erreur lors de la suppression des splits: %w", err)
		}

		for _, split := range splits {
			splitQuery := `
				INSERT INTO recurring_expense_splits (recurring_id, user_id, percentage)
				VALUES ($1, $2, $3)
			`
			_, err = tx.Exec(ctx, splitQuery, recurring.ID, split.UserID, split.Percentage)
			if err != nil {
				return fmt.Errorf("erreur lors de la creation du split: %w", err)
			}
		}
	}

	return tx.Commit(ctx)
}

// DeleteRecurring deletes a recurring expense
func (r *ExpenseRepository) DeleteRecurring(ctx context.Context, id string) error {
	query := `DELETE FROM recurring_expenses WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("depense recurrente introuvable")
	}

	return nil
}

// RecurringBelongsToColocation checks if a recurring expense belongs to a colocation
func (r *ExpenseRepository) RecurringBelongsToColocation(ctx context.Context, recurringID, colocationID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM recurring_expenses WHERE id = $1 AND colocation_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, recurringID, colocationID).Scan(&exists)
	return exists, err
}

// GetActiveRecurringDue gets all active recurring expenses due by a certain date
func (r *ExpenseRepository) GetActiveRecurringDue(ctx context.Context, dueDate time.Time) ([]domain.RecurringExpense, error) {
	query := `
		SELECT re.id, re.colocation_id, re.paid_by, re.category_id, re.title, re.description,
		       re.amount, re.split_type, re.recurrence, re.next_due_date, re.end_date, re.is_active, re.created_at,
		       u.nom, u.prenom, c.name
		FROM recurring_expenses re
		INNER JOIN users u ON re.paid_by = u.id
		INNER JOIN expense_categories c ON re.category_id = c.id
		WHERE re.is_active = true AND re.next_due_date <= $1
		  AND (re.end_date IS NULL OR re.end_date >= $1)
	`

	rows, err := r.pool.Query(ctx, query, dueDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recurrings []domain.RecurringExpense
	for rows.Next() {
		var re domain.RecurringExpense
		if err := rows.Scan(
			&re.ID, &re.ColocationID, &re.PaidBy, &re.CategoryID, &re.Title, &re.Description,
			&re.Amount, &re.SplitType, &re.Recurrence, &re.NextDueDate, &re.EndDate, &re.IsActive, &re.CreatedAt,
			&re.PaidByNom, &re.PaidByPrenom, &re.CategoryName,
		); err != nil {
			return nil, err
		}

		splits, err := r.GetRecurringSplits(ctx, re.ID)
		if err != nil {
			return nil, err
		}
		re.Splits = splits

		recurrings = append(recurrings, re)
	}

	return recurrings, rows.Err()
}

// UpdateNextDueDate updates the next due date for a recurring expense
func (r *ExpenseRepository) UpdateNextDueDate(ctx context.Context, id string, nextDueDate time.Time) error {
	query := `UPDATE recurring_expenses SET next_due_date = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, nextDueDate, id)
	return err
}

// CreateFromRecurring creates an expense from a recurring template
func (r *ExpenseRepository) CreateFromRecurring(ctx context.Context, recurring *domain.RecurringExpense) (*domain.Expense, error) {
	expense := &domain.Expense{
		ColocationID: recurring.ColocationID,
		PaidBy:       recurring.PaidBy,
		CategoryID:   recurring.CategoryID,
		Title:        recurring.Title,
		Description:  recurring.Description,
		Amount:       recurring.Amount,
		SplitType:    recurring.SplitType,
		ExpenseDate:  recurring.NextDueDate,
		RecurringID:  &recurring.ID,
	}

	// Convert recurring splits to expense splits
	var splits []domain.ExpenseSplitInput
	for _, rs := range recurring.Splits {
		splits = append(splits, domain.ExpenseSplitInput{
			UserID:     rs.UserID,
			Percentage: rs.Percentage,
			Amount:     recurring.Amount * rs.Percentage / 100,
		})
	}

	if err := r.Create(ctx, expense, splits); err != nil {
		return nil, err
	}

	return expense, nil
}

// GetForecastData returns data needed for expense forecast
func (r *ExpenseRepository) GetForecastData(ctx context.Context, colocationID string, months int) ([]domain.MonthlyForecast, error) {
	// Get recurring expenses for forecast
	recurrings, err := r.ListRecurringByColocation(ctx, colocationID)
	if err != nil {
		return nil, err
	}

	// Get historical monthly averages by category
	query := `
		SELECT c.id, c.name, AVG(e.amount) as avg_amount
		FROM expenses e
		INNER JOIN expense_categories c ON e.category_id = c.id
		WHERE e.colocation_id = $1
		  AND e.expense_date >= NOW() - INTERVAL '3 months'
		  AND e.recurring_id IS NULL
		GROUP BY c.id, c.name
	`

	rows, err := r.pool.Query(ctx, query, colocationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	historicalAvg := make(map[string]float64)
	categoryNames := make(map[string]string)
	for rows.Next() {
		var catID, catName string
		var avg float64
		if err := rows.Scan(&catID, &catName, &avg); err != nil {
			return nil, err
		}
		historicalAvg[catID] = avg
		categoryNames[catID] = catName
	}

	// Build forecasts
	var forecasts []domain.MonthlyForecast
	now := time.Now()

	for i := 0; i < months; i++ {
		month := now.AddDate(0, i+1, 0)
		monthStr := month.Format("2006-01")

		categoryTotals := make(map[string]float64)

		// Add recurring expenses
		for _, re := range recurrings {
			if !re.IsActive {
				continue
			}
			if re.EndDate != nil && re.EndDate.Before(month) {
				continue
			}

			// Count occurrences in this month
			occurrences := countOccurrencesInMonth(re.Recurrence, month)
			total := re.Amount * float64(occurrences)

			categoryTotals[re.CategoryID] += total
			categoryNames[re.CategoryID] = re.CategoryName
		}

		// Add historical averages for non-recurring
		for catID, avg := range historicalAvg {
			categoryTotals[catID] += avg
		}

		// Build category forecasts
		var categories []domain.CategoryForecast
		var totalAmount float64
		for catID, amount := range categoryTotals {
			categories = append(categories, domain.CategoryForecast{
				CategoryID:   catID,
				CategoryName: categoryNames[catID],
				Amount:       amount,
			})
			totalAmount += amount
		}

		forecasts = append(forecasts, domain.MonthlyForecast{
			Month:       monthStr,
			TotalAmount: totalAmount,
			Categories:  categories,
		})
	}

	return forecasts, nil
}

// countOccurrencesInMonth counts how many times a recurring expense occurs in a given month
func countOccurrencesInMonth(recurrence domain.Recurrence, month time.Time) int {
	switch recurrence {
	case domain.RecurrenceDaily:
		return daysInMonth(month)
	case domain.RecurrenceWeekly:
		return daysInMonth(month) / 7
	case domain.RecurrenceMonthly:
		return 1
	case domain.RecurrenceYearly:
		return 0 // Simplified: assume yearly expenses don't occur every month
	default:
		return 0
	}
}

func daysInMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
}
