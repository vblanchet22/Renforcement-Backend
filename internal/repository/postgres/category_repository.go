package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/domain"
)

// CategoryRepository handles category database operations
type CategoryRepository struct {
	pool *pgxpool.Pool
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

// ListByColocation returns all categories available for a colocation (global + custom)
func (r *CategoryRepository) ListByColocation(ctx context.Context, colocationID string) ([]domain.ExpenseCategory, error) {
	query := `
		SELECT id, name, icon, color, colocation_id, created_at
		FROM expense_categories
		WHERE colocation_id IS NULL OR colocation_id = $1
		ORDER BY colocation_id NULLS FIRST, name
	`

	rows, err := r.pool.Query(ctx, query, colocationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.ExpenseCategory
	for rows.Next() {
		var c domain.ExpenseCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.Color, &c.ColocationID, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, rows.Err()
}

// GetByID returns a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id string) (*domain.ExpenseCategory, error) {
	query := `
		SELECT id, name, icon, color, colocation_id, created_at
		FROM expense_categories
		WHERE id = $1
	`

	var c domain.ExpenseCategory
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Icon, &c.Color, &c.ColocationID, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// Create creates a new custom category for a colocation
func (r *CategoryRepository) Create(ctx context.Context, category *domain.ExpenseCategory) error {
	query := `
		INSERT INTO expense_categories (name, icon, color, colocation_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	return r.pool.QueryRow(ctx, query,
		category.Name,
		category.Icon,
		category.Color,
		category.ColocationID,
	).Scan(&category.ID, &category.CreatedAt)
}

// Update updates a custom category
func (r *CategoryRepository) Update(ctx context.Context, category *domain.ExpenseCategory) error {
	query := `
		UPDATE expense_categories
		SET name = $1, icon = $2, color = $3
		WHERE id = $4 AND colocation_id IS NOT NULL
	`

	_, err := r.pool.Exec(ctx, query,
		category.Name,
		category.Icon,
		category.Color,
		category.ID,
	)
	return err
}

// Delete deletes a custom category
func (r *CategoryRepository) Delete(ctx context.Context, id string) error {
	// Only delete custom categories (colocation_id IS NOT NULL)
	query := `DELETE FROM expense_categories WHERE id = $1 AND colocation_id IS NOT NULL`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// GetStats returns category statistics for a colocation within a date range
func (r *CategoryRepository) GetStats(ctx context.Context, colocationID string, startDate, endDate *time.Time) ([]domain.CategoryStat, float64, error) {
	query := `
		SELECT
			c.id,
			c.name,
			c.icon,
			c.color,
			COALESCE(SUM(e.amount), 0) as total_amount,
			COUNT(e.id) as expense_count
		FROM expense_categories c
		LEFT JOIN expenses e ON e.category_id = c.id
			AND e.colocation_id = $1
			AND ($2::timestamp IS NULL OR e.expense_date >= $2)
			AND ($3::timestamp IS NULL OR e.expense_date <= $3)
		WHERE c.colocation_id IS NULL OR c.colocation_id = $1
		GROUP BY c.id, c.name, c.icon, c.color
		ORDER BY total_amount DESC
	`

	rows, err := r.pool.Query(ctx, query, colocationID, startDate, endDate)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var stats []domain.CategoryStat
	var totalAmount float64

	for rows.Next() {
		var s domain.CategoryStat
		if err := rows.Scan(
			&s.CategoryID,
			&s.CategoryName,
			&s.Icon,
			&s.Color,
			&s.TotalAmount,
			&s.ExpenseCount,
		); err != nil {
			return nil, 0, err
		}
		totalAmount += s.TotalAmount
		stats = append(stats, s)
	}

	// Calculate percentages
	for i := range stats {
		if totalAmount > 0 {
			stats[i].Percentage = (stats[i].TotalAmount / totalAmount) * 100
		}
	}

	return stats, totalAmount, rows.Err()
}

// BelongsToColocation checks if a category belongs to a colocation (or is global)
func (r *CategoryRepository) BelongsToColocation(ctx context.Context, categoryID, colocationID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM expense_categories
			WHERE id = $1 AND (colocation_id IS NULL OR colocation_id = $2)
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, categoryID, colocationID).Scan(&exists)
	return exists, err
}
