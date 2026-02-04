package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/domain"
)

// PaymentRepository handles payment database operations
type PaymentRepository struct {
	pool *pgxpool.Pool
}

// NewPaymentRepository creates a new PaymentRepository
func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

// Create creates a new payment
func (r *PaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (colocation_id, from_user_id, to_user_id, amount, note)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, status, created_at
	`

	return r.pool.QueryRow(ctx, query,
		payment.ColocationID,
		payment.FromUserID,
		payment.ToUserID,
		payment.Amount,
		payment.Note,
	).Scan(&payment.ID, &payment.Status, &payment.CreatedAt)
}

// GetByID retrieves a payment by ID with user details
func (r *PaymentRepository) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	query := `
		SELECT p.id, p.colocation_id, p.from_user_id, p.to_user_id, p.amount, p.status, p.note, p.confirmed_at, p.created_at,
		       fu.nom, fu.prenom, fu.avatar_url,
		       tu.nom, tu.prenom, tu.avatar_url
		FROM payments p
		INNER JOIN users fu ON p.from_user_id = fu.id
		INNER JOIN users tu ON p.to_user_id = tu.id
		WHERE p.id = $1
	`

	var p domain.Payment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.ColocationID, &p.FromUserID, &p.ToUserID, &p.Amount, &p.Status, &p.Note, &p.ConfirmedAt, &p.CreatedAt,
		&p.FromUserNom, &p.FromUserPrenom, &p.FromAvatarURL,
		&p.ToUserNom, &p.ToUserPrenom, &p.ToAvatarURL,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation du paiement: %w", err)
	}

	return &p, nil
}

// ListByColocation lists payments for a colocation with filters
func (r *PaymentRepository) ListByColocation(ctx context.Context, colocationID string, status, fromUserID, toUserID *string, page, pageSize int) ([]domain.Payment, int, error) {
	baseQuery := `
		FROM payments p
		INNER JOIN users fu ON p.from_user_id = fu.id
		INNER JOIN users tu ON p.to_user_id = tu.id
		WHERE p.colocation_id = $1
	`

	args := []interface{}{colocationID}
	argIndex := 2

	if status != nil {
		baseQuery += fmt.Sprintf(" AND p.status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	if fromUserID != nil {
		baseQuery += fmt.Sprintf(" AND p.from_user_id = $%d", argIndex)
		args = append(args, *fromUserID)
		argIndex++
	}

	if toUserID != nil {
		baseQuery += fmt.Sprintf(" AND p.to_user_id = $%d", argIndex)
		args = append(args, *toUserID)
		argIndex++
	}

	// Count
	var totalCount int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) "+baseQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Select
	selectQuery := `
		SELECT p.id, p.colocation_id, p.from_user_id, p.to_user_id, p.amount, p.status, p.note, p.confirmed_at, p.created_at,
		       fu.nom, fu.prenom, fu.avatar_url,
		       tu.nom, tu.prenom, tu.avatar_url
	` + baseQuery + fmt.Sprintf(" ORDER BY p.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(
			&p.ID, &p.ColocationID, &p.FromUserID, &p.ToUserID, &p.Amount, &p.Status, &p.Note, &p.ConfirmedAt, &p.CreatedAt,
			&p.FromUserNom, &p.FromUserPrenom, &p.FromAvatarURL,
			&p.ToUserNom, &p.ToUserPrenom, &p.ToAvatarURL,
		); err != nil {
			return nil, 0, err
		}
		payments = append(payments, p)
	}

	return payments, totalCount, rows.Err()
}

// UpdateStatus updates the status of a payment
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id, status string) error {
	var query string
	if status == "confirmed" {
		query = `UPDATE payments SET status = $1, confirmed_at = NOW() WHERE id = $2`
	} else {
		query = `UPDATE payments SET status = $1 WHERE id = $2`
	}

	result, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise a jour du statut: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("paiement introuvable")
	}

	return nil
}

// Delete deletes a payment
func (r *PaymentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM payments WHERE id = $1 AND status = 'pending'`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("paiement introuvable ou deja traite")
	}

	return nil
}

// SettleExpenseSplits marks expense splits as settled when a payment is confirmed
func (r *PaymentRepository) SettleExpenseSplits(ctx context.Context, colocationID, fromUserID, toUserID string, amount float64) error {
	// Get unsettled splits where fromUser owes toUser
	query := `
		UPDATE expense_splits es
		SET is_settled = true
		FROM expenses e
		WHERE es.expense_id = e.id
		  AND e.colocation_id = $1
		  AND es.user_id = $2
		  AND e.paid_by = $3
		  AND es.is_settled = false
		  AND es.id IN (
			SELECT es2.id
			FROM expense_splits es2
			INNER JOIN expenses e2 ON es2.expense_id = e2.id
			WHERE e2.colocation_id = $1 AND es2.user_id = $2 AND e2.paid_by = $3 AND es2.is_settled = false
			ORDER BY e2.expense_date ASC
			LIMIT (
				SELECT COUNT(*) FROM expense_splits es3
				INNER JOIN expenses e3 ON es3.expense_id = e3.id
				WHERE e3.colocation_id = $1 AND es3.user_id = $2 AND e3.paid_by = $3 AND es3.is_settled = false
				AND (SELECT COALESCE(SUM(es4.amount), 0) FROM expense_splits es4
					INNER JOIN expenses e4 ON es4.expense_id = e4.id
					WHERE e4.colocation_id = $1 AND es4.user_id = $2 AND e4.paid_by = $3 AND es4.is_settled = false
					AND e4.expense_date <= e3.expense_date) <= $4
			)
		  )
	`

	_, err := r.pool.Exec(ctx, query, colocationID, fromUserID, toUserID, amount)
	return err
}

// SaveBalance upserts a balance record
func (r *PaymentRepository) SaveBalance(ctx context.Context, colocationID, fromUserID, toUserID string, amount float64) error {
	query := `
		INSERT INTO balances (colocation_id, from_user_id, to_user_id, amount)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (colocation_id, from_user_id, to_user_id)
		DO UPDATE SET amount = $4, updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query, colocationID, fromUserID, toUserID, amount)
	return err
}

// PaymentExistsAndPending checks if a payment exists and is pending
func (r *PaymentRepository) PaymentExistsAndPending(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM payments WHERE id = $1 AND status = 'pending')`
	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	return exists, err
}

// dummy usage to prevent import errors
var _ = time.Now
