package domain

import "time"

// NotificationType represents the type of notification
type NotificationType string

const (
	NotifExpenseCreated    NotificationType = "expense_created"
	NotifExpenseUpdated    NotificationType = "expense_updated"
	NotifExpenseDeleted    NotificationType = "expense_deleted"
	NotifPaymentReceived   NotificationType = "payment_received"
	NotifPaymentConfirmed  NotificationType = "payment_confirmed"
	NotifPaymentRejected   NotificationType = "payment_rejected"
	NotifMemberJoined      NotificationType = "member_joined"
	NotifMemberLeft        NotificationType = "member_left"
	NotifMemberRemoved     NotificationType = "member_removed"
	NotifInvitationReceived NotificationType = "invitation_received"
	NotifRoleChanged       NotificationType = "role_changed"
	NotifDecisionCreated   NotificationType = "decision_created"
	NotifDecisionClosed    NotificationType = "decision_closed"
	NotifDecisionDeadline  NotificationType = "decision_deadline"
	NotifFundCreated       NotificationType = "fund_created"
	NotifFundContribution  NotificationType = "fund_contribution"
	NotifFundGoalReached   NotificationType = "fund_goal_reached"
	NotifEventCreated      NotificationType = "event_created"
	NotifEventUpdated      NotificationType = "event_updated"
	NotifEventReminder     NotificationType = "event_reminder"
	NotifEventCancelled    NotificationType = "event_cancelled"
	NotifRecurringDue      NotificationType = "recurring_due"
)

// Notification represents a notification for a user
type Notification struct {
	ID           string            `json:"id" db:"id"`
	UserID       string            `json:"user_id" db:"user_id"`
	ColocationID *string           `json:"colocation_id,omitempty" db:"colocation_id"`
	Type         NotificationType  `json:"type" db:"type"`
	Title        string            `json:"title" db:"title"`
	Body         string            `json:"body" db:"body"`
	Data         map[string]string `json:"data,omitempty" db:"data"` // JSONB
	IsRead       bool              `json:"is_read" db:"is_read"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`

	// Joined fields
	ColocationName *string `json:"colocation_name,omitempty"`
}
