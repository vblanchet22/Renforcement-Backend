package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

// NotificationService handles notification business logic and streaming
type NotificationService struct {
	repo *postgres.NotificationRepository

	// Hub for real-time streaming
	mu          sync.RWMutex
	subscribers map[string][]chan *domain.Notification // userID -> channels
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(repo *postgres.NotificationRepository) *NotificationService {
	return &NotificationService{
		repo:        repo,
		subscribers: make(map[string][]chan *domain.Notification),
	}
}

// List lists notifications for the current user
func (s *NotificationService) List(ctx context.Context, colocationID *string, unreadOnly bool, page, pageSize int) ([]domain.Notification, int, int, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, 0, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.repo.ListByUser(ctx, userID, colocationID, unreadOnly, page, pageSize)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notifID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.MarkAsRead(ctx, notifID, userID)
}

// MarkAllAsRead marks all notifications as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context, colocationID *string) (int, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return 0, err
	}

	return s.repo.MarkAllAsRead(ctx, userID, colocationID)
}

// Delete deletes a notification
func (s *NotificationService) Delete(ctx context.Context, notifID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, notifID, userID)
}

// GetUnreadCount returns the unread notification count
func (s *NotificationService) GetUnreadCount(ctx context.Context, colocationID *string) (int, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return 0, err
	}

	return s.repo.GetUnreadCount(ctx, userID, colocationID)
}

// Subscribe adds a subscriber for real-time notifications
func (s *NotificationService) Subscribe(userID string) chan *domain.Notification {
	ch := make(chan *domain.Notification, 100)

	s.mu.Lock()
	s.subscribers[userID] = append(s.subscribers[userID], ch)
	s.mu.Unlock()

	return ch
}

// Unsubscribe removes a subscriber
func (s *NotificationService) Unsubscribe(userID string, ch chan *domain.Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()

	channels := s.subscribers[userID]
	for i, c := range channels {
		if c == ch {
			s.subscribers[userID] = append(channels[:i], channels[i+1:]...)
			close(ch)
			break
		}
	}

	if len(s.subscribers[userID]) == 0 {
		delete(s.subscribers, userID)
	}
}

// Notify sends a notification to a specific user (and persists it)
func (s *NotificationService) Notify(ctx context.Context, notif *domain.Notification) error {
	// Persist
	if err := s.repo.Create(ctx, notif); err != nil {
		return fmt.Errorf("erreur lors de la creation de la notification: %w", err)
	}

	// Broadcast to subscribers
	s.mu.RLock()
	channels := s.subscribers[notif.UserID]
	s.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- notif:
		default:
			// Channel full, skip
		}
	}

	return nil
}

// NotifyColocationMembers sends a notification to all members of a colocation
func (s *NotificationService) NotifyColocationMembers(ctx context.Context, colocationID, excludeUserID string, notifType domain.NotificationType, title, body string, data map[string]string) error {
	return s.repo.CreateForColocationMembers(ctx, colocationID, excludeUserID, notifType, title, body, data)
}
