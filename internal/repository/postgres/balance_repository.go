package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/domain"
)

// BalanceRepository handles balance database operations
type BalanceRepository struct {
	pool *pgxpool.Pool
}

// NewBalanceRepository creates a new BalanceRepository
func NewBalanceRepository(pool *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{pool: pool}
}

// GetUserBalances calculates balances for all members of a colocation
func (r *BalanceRepository) GetUserBalances(ctx context.Context, colocationID string) ([]domain.UserBalance, error) {
	query := `
		WITH member_paid AS (
			SELECT e.paid_by as user_id, COALESCE(SUM(e.amount), 0) as total_paid
			FROM expenses e
			WHERE e.colocation_id = $1
			GROUP BY e.paid_by
		),
		member_owed AS (
			SELECT es.user_id, COALESCE(SUM(es.amount), 0) as total_owed
			FROM expense_splits es
			INNER JOIN expenses e ON es.expense_id = e.id
			WHERE e.colocation_id = $1 AND es.is_settled = false
			GROUP BY es.user_id
		),
		payments_made AS (
			SELECT p.from_user_id as user_id, COALESCE(SUM(p.amount), 0) as total
			FROM payments p
			WHERE p.colocation_id = $1 AND p.status = 'confirmed'
			GROUP BY p.from_user_id
		),
		payments_received AS (
			SELECT p.to_user_id as user_id, COALESCE(SUM(p.amount), 0) as total
			FROM payments p
			WHERE p.colocation_id = $1 AND p.status = 'confirmed'
			GROUP BY p.to_user_id
		)
		SELECT
			cm.user_id,
			u.nom,
			u.prenom,
			u.avatar_url,
			COALESCE(mp.total_paid, 0) as total_paid,
			COALESCE(mo.total_owed, 0) as total_owed,
			(COALESCE(mp.total_paid, 0) - COALESCE(mo.total_owed, 0) + COALESCE(pm.total, 0) - COALESCE(pr.total, 0)) as net_balance
		FROM colocation_members cm
		INNER JOIN users u ON cm.user_id = u.id
		LEFT JOIN member_paid mp ON cm.user_id = mp.user_id
		LEFT JOIN member_owed mo ON cm.user_id = mo.user_id
		LEFT JOIN payments_made pm ON cm.user_id = pm.user_id
		LEFT JOIN payments_received pr ON cm.user_id = pr.user_id
		WHERE cm.colocation_id = $1
		ORDER BY net_balance DESC
	`

	rows, err := r.pool.Query(ctx, query, colocationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []domain.UserBalance
	for rows.Next() {
		var b domain.UserBalance
		if err := rows.Scan(
			&b.UserID, &b.UserNom, &b.UserPrenom, &b.AvatarURL,
			&b.TotalPaid, &b.TotalOwed, &b.NetBalance,
		); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}

	return balances, rows.Err()
}

// GetRawDebts returns all unsettled debts between members
func (r *BalanceRepository) GetRawDebts(ctx context.Context, colocationID string) ([]domain.Debt, error) {
	query := `
		SELECT
			es.user_id as from_user_id,
			fu.nom as from_user_nom,
			fu.prenom as from_user_prenom,
			e.paid_by as to_user_id,
			tu.nom as to_user_nom,
			tu.prenom as to_user_prenom,
			SUM(es.amount) as amount
		FROM expense_splits es
		INNER JOIN expenses e ON es.expense_id = e.id
		INNER JOIN users fu ON es.user_id = fu.id
		INNER JOIN users tu ON e.paid_by = tu.id
		WHERE e.colocation_id = $1
		  AND es.is_settled = false
		  AND es.user_id != e.paid_by
		GROUP BY es.user_id, fu.nom, fu.prenom, e.paid_by, tu.nom, tu.prenom
		HAVING SUM(es.amount) > 0.01
		ORDER BY amount DESC
	`

	rows, err := r.pool.Query(ctx, query, colocationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debts []domain.Debt
	for rows.Next() {
		var d domain.Debt
		if err := rows.Scan(
			&d.FromUserID, &d.FromUserNom, &d.FromUserPrenom,
			&d.ToUserID, &d.ToUserNom, &d.ToUserPrenom,
			&d.Amount,
		); err != nil {
			return nil, err
		}
		debts = append(debts, d)
	}

	return debts, rows.Err()
}

// MemberInfo stores member info for building simplified debts
type MemberInfo struct {
	UserID    string
	Nom       string
	Prenom    string
	AvatarURL *string
}

// GetMembersInfo returns member info for a colocation
func (r *BalanceRepository) GetMembersInfo(ctx context.Context, colocationID string) ([]MemberInfo, error) {
	query := `
		SELECT cm.user_id, u.nom, u.prenom, u.avatar_url
		FROM colocation_members cm
		INNER JOIN users u ON cm.user_id = u.id
		WHERE cm.colocation_id = $1
		ORDER BY cm.joined_at
	`

	rows, err := r.pool.Query(ctx, query, colocationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []MemberInfo
	for rows.Next() {
		var m MemberInfo
		if err := rows.Scan(&m.UserID, &m.Nom, &m.Prenom, &m.AvatarURL); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	return members, rows.Err()
}

// GetBalanceHistory returns the balance history for a user in a colocation
func (r *BalanceRepository) GetBalanceHistory(ctx context.Context, colocationID, userID string, startDate, endDate *time.Time) ([]domain.BalanceHistoryEntry, error) {
	query := `
		WITH events AS (
			-- Expenses where user paid
			SELECT
				e.expense_date as date,
				'expense' as event_type,
				e.id as event_id,
				e.title as description,
				e.amount as amount
			FROM expenses e
			WHERE e.colocation_id = $1 AND e.paid_by = $2
				AND ($3::timestamp IS NULL OR e.expense_date >= $3)
				AND ($4::timestamp IS NULL OR e.expense_date <= $4)

			UNION ALL

			-- Expense splits where user owes
			SELECT
				e.expense_date as date,
				'expense' as event_type,
				e.id as event_id,
				e.title as description,
				-es.amount as amount
			FROM expense_splits es
			INNER JOIN expenses e ON es.expense_id = e.id
			WHERE e.colocation_id = $1 AND es.user_id = $2 AND e.paid_by != $2
				AND ($3::timestamp IS NULL OR e.expense_date >= $3)
				AND ($4::timestamp IS NULL OR e.expense_date <= $4)

			UNION ALL

			-- Payments made (positive for payer)
			SELECT
				p.created_at as date,
				'payment' as event_type,
				p.id as event_id,
				COALESCE(p.note, 'Paiement') as description,
				-p.amount as amount
			FROM payments p
			WHERE p.colocation_id = $1 AND p.from_user_id = $2 AND p.status = 'confirmed'
				AND ($3::timestamp IS NULL OR p.created_at >= $3)
				AND ($4::timestamp IS NULL OR p.created_at <= $4)

			UNION ALL

			-- Payments received
			SELECT
				p.created_at as date,
				'payment' as event_type,
				p.id as event_id,
				COALESCE(p.note, 'Paiement recu') as description,
				p.amount as amount
			FROM payments p
			WHERE p.colocation_id = $1 AND p.to_user_id = $2 AND p.status = 'confirmed'
				AND ($3::timestamp IS NULL OR p.created_at >= $3)
				AND ($4::timestamp IS NULL OR p.created_at <= $4)
		)
		SELECT date, event_type, event_id, description, amount
		FROM events
		ORDER BY date ASC
	`

	rows, err := r.pool.Query(ctx, query, colocationID, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.BalanceHistoryEntry
	var cumulative float64

	for rows.Next() {
		var e domain.BalanceHistoryEntry
		if err := rows.Scan(&e.Date, &e.EventType, &e.EventID, &e.Description, &e.Amount); err != nil {
			return nil, err
		}
		cumulative += e.Amount
		e.CumulativeBalance = cumulative
		entries = append(entries, e)
	}

	return entries, rows.Err()
}
