package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/domain"
)

// NotificationRepository handles notification database operations
type NotificationRepository struct {
	pool *pgxpool.Pool
}

// NewNotificationRepository creates a new NotificationRepository
func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	dataJSON, err := json.Marshal(notif.Data)
	if err != nil {
		dataJSON = []byte("{}")
	}

	query := `
		INSERT INTO notifications (user_id, colocation_id, type, title, body, data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, is_read, created_at
	`

	return r.pool.QueryRow(ctx, query,
		notif.UserID,
		notif.ColocationID,
		notif.Type,
		notif.Title,
		notif.Body,
		dataJSON,
	).Scan(&notif.ID, &notif.IsRead, &notif.CreatedAt)
}

// ListByUser lists notifications for a user
func (r *NotificationRepository) ListByUser(ctx context.Context, userID string, colocationID *string, unreadOnly bool, page, pageSize int) ([]domain.Notification, int, int, error) {
	baseQuery := `
		FROM notifications n
		LEFT JOIN colocations c ON n.colocation_id = c.id
		WHERE n.user_id = $1
	`

	args := []interface{}{userID}
	argIndex := 2

	if colocationID != nil {
		baseQuery += fmt.Sprintf(" AND n.colocation_id = $%d", argIndex)
		args = append(args, *colocationID)
		argIndex++
	}

	if unreadOnly {
		baseQuery += " AND n.is_read = false"
	}

	// Count total
	var totalCount int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) "+baseQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, 0, err
	}

	// Count unread
	var unreadCount int
	unreadQuery := `
		FROM notifications n
		WHERE n.user_id = $1 AND n.is_read = false
	`
	unreadArgs := []interface{}{userID}
	if colocationID != nil {
		unreadQuery += " AND n.colocation_id = $2"
		unreadArgs = append(unreadArgs, *colocationID)
	}
	err = r.pool.QueryRow(ctx, "SELECT COUNT(*) "+unreadQuery, unreadArgs...).Scan(&unreadCount)
	if err != nil {
		return nil, 0, 0, err
	}

	// Select
	selectQuery := fmt.Sprintf(`
		SELECT n.id, n.user_id, n.colocation_id, n.type, n.title, n.body, n.data, n.is_read, n.created_at,
		       c.name
	`+baseQuery+" ORDER BY n.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		var dataJSON []byte

		if err := rows.Scan(
			&n.ID, &n.UserID, &n.ColocationID, &n.Type, &n.Title, &n.Body, &dataJSON, &n.IsRead, &n.CreatedAt,
			&n.ColocationName,
		); err != nil {
			return nil, 0, 0, err
		}

		if dataJSON != nil {
			json.Unmarshal(dataJSON, &n.Data)
		}

		notifications = append(notifications, n)
	}

	return notifications, totalCount, unreadCount, rows.Err()
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id, userID string) error {
	query := `UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`
	result, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification introuvable")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID string, colocationID *string) (int, error) {
	query := `UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false`
	args := []interface{}{userID}

	if colocationID != nil {
		query += " AND colocation_id = $2"
		args = append(args, *colocationID)
	}

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM notifications WHERE id = $1 AND user_id = $2`
	result, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification introuvable")
	}

	return nil
}

// GetUnreadCount returns the unread notification count
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID string, colocationID *string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	args := []interface{}{userID}

	if colocationID != nil {
		query += " AND colocation_id = $2"
		args = append(args, *colocationID)
	}

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

// CreateForColocationMembers creates notifications for all members of a colocation except the sender
func (r *NotificationRepository) CreateForColocationMembers(ctx context.Context, colocationID, excludeUserID string, notifType domain.NotificationType, title, body string, data map[string]string) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		dataJSON = []byte("{}")
	}

	query := `
		INSERT INTO notifications (user_id, colocation_id, type, title, body, data)
		SELECT cm.user_id, $1, $2, $3, $4, $5
		FROM colocation_members cm
		WHERE cm.colocation_id = $1 AND cm.user_id != $6
	`

	_, err = r.pool.Exec(ctx, query, colocationID, notifType, title, body, dataJSON, excludeUserID)
	return err
}
