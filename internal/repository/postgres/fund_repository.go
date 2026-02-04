package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/domain"
)

// FundRepository handles fund database operations
type FundRepository struct {
	pool *pgxpool.Pool
}

// NewFundRepository creates a new FundRepository
func NewFundRepository(pool *pgxpool.Pool) *FundRepository {
	return &FundRepository{pool: pool}
}

// Create creates a new fund
func (r *FundRepository) Create(ctx context.Context, fund *domain.CommonFund) error {
	query := `
		INSERT INTO common_funds (colocation_id, name, description, target_amount, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, current_amount, is_active, created_at
	`

	return r.pool.QueryRow(ctx, query,
		fund.ColocationID,
		fund.Name,
		fund.Description,
		fund.TargetAmount,
		fund.CreatedBy,
	).Scan(&fund.ID, &fund.CurrentAmount, &fund.IsActive, &fund.CreatedAt)
}

// GetByID retrieves a fund by ID with contributor summary
func (r *FundRepository) GetByID(ctx context.Context, id string) (*domain.CommonFund, error) {
	query := `
		SELECT f.id, f.colocation_id, f.name, f.description, f.target_amount, f.current_amount,
		       f.is_active, f.created_by, f.created_at,
		       u.nom, u.prenom
		FROM common_funds f
		INNER JOIN users u ON f.created_by = u.id
		WHERE f.id = $1
	`

	var f domain.CommonFund
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&f.ID, &f.ColocationID, &f.Name, &f.Description, &f.TargetAmount, &f.CurrentAmount,
		&f.IsActive, &f.CreatedBy, &f.CreatedAt,
		&f.CreatedByNom, &f.CreatedByPrenom,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}

	// Calculate progress
	if f.TargetAmount != nil && *f.TargetAmount > 0 {
		f.ProgressPercentage = (f.CurrentAmount / *f.TargetAmount) * 100
	}

	// Get contributors
	contributors, err := r.GetContributors(ctx, id)
	if err != nil {
		return nil, err
	}
	f.Contributors = contributors

	return &f, nil
}

// GetContributors returns contributor summaries for a fund
func (r *FundRepository) GetContributors(ctx context.Context, fundID string) ([]domain.ContributorSummary, error) {
	query := `
		SELECT fc.user_id, u.nom, u.prenom, SUM(fc.amount) as total
		FROM fund_contributions fc
		INNER JOIN users u ON fc.user_id = u.id
		WHERE fc.fund_id = $1
		GROUP BY fc.user_id, u.nom, u.prenom
		ORDER BY total DESC
	`

	rows, err := r.pool.Query(ctx, query, fundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contributors []domain.ContributorSummary
	for rows.Next() {
		var c domain.ContributorSummary
		if err := rows.Scan(&c.UserID, &c.UserNom, &c.UserPrenom, &c.TotalContributed); err != nil {
			return nil, err
		}
		contributors = append(contributors, c)
	}

	return contributors, rows.Err()
}

// ListByColocation lists funds for a colocation
func (r *FundRepository) ListByColocation(ctx context.Context, colocationID string, isActive *bool) ([]domain.CommonFund, error) {
	query := `
		SELECT f.id, f.colocation_id, f.name, f.description, f.target_amount, f.current_amount,
		       f.is_active, f.created_by, f.created_at,
		       u.nom, u.prenom
		FROM common_funds f
		INNER JOIN users u ON f.created_by = u.id
		WHERE f.colocation_id = $1
	`

	args := []interface{}{colocationID}
	if isActive != nil {
		query += " AND f.is_active = $2"
		args = append(args, *isActive)
	}

	query += " ORDER BY f.created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var funds []domain.CommonFund
	for rows.Next() {
		var f domain.CommonFund
		if err := rows.Scan(
			&f.ID, &f.ColocationID, &f.Name, &f.Description, &f.TargetAmount, &f.CurrentAmount,
			&f.IsActive, &f.CreatedBy, &f.CreatedAt,
			&f.CreatedByNom, &f.CreatedByPrenom,
		); err != nil {
			return nil, err
		}

		if f.TargetAmount != nil && *f.TargetAmount > 0 {
			f.ProgressPercentage = (f.CurrentAmount / *f.TargetAmount) * 100
		}

		contributors, err := r.GetContributors(ctx, f.ID)
		if err != nil {
			return nil, err
		}
		f.Contributors = contributors

		funds = append(funds, f)
	}

	return funds, rows.Err()
}

// Update updates a fund
func (r *FundRepository) Update(ctx context.Context, fund *domain.CommonFund) error {
	query := `
		UPDATE common_funds
		SET name = $1, description = $2, target_amount = $3, is_active = $4
		WHERE id = $5
	`

	_, err := r.pool.Exec(ctx, query,
		fund.Name, fund.Description, fund.TargetAmount, fund.IsActive, fund.ID,
	)
	return err
}

// Delete deletes a fund
func (r *FundRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM common_funds WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("fonds introuvable")
	}

	return nil
}

// AddContribution adds a contribution to a fund
func (r *FundRepository) AddContribution(ctx context.Context, contribution *domain.FundContribution) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO fund_contributions (fund_id, user_id, amount, note)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err = tx.QueryRow(ctx, query,
		contribution.FundID, contribution.UserID, contribution.Amount, contribution.Note,
	).Scan(&contribution.ID, &contribution.CreatedAt)
	if err != nil {
		return err
	}

	// Update fund current_amount
	_, err = tx.Exec(ctx,
		"UPDATE common_funds SET current_amount = current_amount + $1 WHERE id = $2",
		contribution.Amount, contribution.FundID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ListContributions lists contributions for a fund
func (r *FundRepository) ListContributions(ctx context.Context, fundID string) ([]domain.FundContribution, error) {
	query := `
		SELECT fc.id, fc.fund_id, fc.user_id, fc.amount, fc.note, fc.created_at,
		       u.nom, u.prenom
		FROM fund_contributions fc
		INNER JOIN users u ON fc.user_id = u.id
		WHERE fc.fund_id = $1
		ORDER BY fc.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, fundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contributions []domain.FundContribution
	for rows.Next() {
		var c domain.FundContribution
		if err := rows.Scan(
			&c.ID, &c.FundID, &c.UserID, &c.Amount, &c.Note, &c.CreatedAt,
			&c.UserNom, &c.UserPrenom,
		); err != nil {
			return nil, err
		}
		contributions = append(contributions, c)
	}

	return contributions, rows.Err()
}

// GetContribution retrieves a contribution by ID
func (r *FundRepository) GetContribution(ctx context.Context, id string) (*domain.FundContribution, error) {
	query := `
		SELECT fc.id, fc.fund_id, fc.user_id, fc.amount, fc.note, fc.created_at,
		       u.nom, u.prenom
		FROM fund_contributions fc
		INNER JOIN users u ON fc.user_id = u.id
		WHERE fc.id = $1
	`

	var c domain.FundContribution
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.FundID, &c.UserID, &c.Amount, &c.Note, &c.CreatedAt,
		&c.UserNom, &c.UserPrenom,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// DeleteContribution deletes a contribution and updates the fund amount
func (r *FundRepository) DeleteContribution(ctx context.Context, id, fundID string, amount float64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, "DELETE FROM fund_contributions WHERE id = $1", id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("contribution introuvable")
	}

	// Update fund current_amount
	_, err = tx.Exec(ctx,
		"UPDATE common_funds SET current_amount = current_amount - $1 WHERE id = $2",
		amount, fundID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
