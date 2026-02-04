package domain

import "time"

// UserBalance represents a user's balance in a colocation
type UserBalance struct {
	UserID     string  `json:"user_id"`
	UserNom    string  `json:"user_nom"`
	UserPrenom string  `json:"user_prenom"`
	AvatarURL  *string `json:"avatar_url,omitempty"`
	TotalPaid  float64 `json:"total_paid"`  // Total amount paid by user
	TotalOwed  float64 `json:"total_owed"`  // Total amount user owes
	NetBalance float64 `json:"net_balance"` // Positive = others owe them, Negative = they owe others
}

// Debt represents a debt from one user to another
type Debt struct {
	FromUserID     string  `json:"from_user_id"`
	FromUserNom    string  `json:"from_user_nom"`
	FromUserPrenom string  `json:"from_user_prenom"`
	ToUserID       string  `json:"to_user_id"`
	ToUserNom      string  `json:"to_user_nom"`
	ToUserPrenom   string  `json:"to_user_prenom"`
	Amount         float64 `json:"amount"`
}

// SimplifiedDebt represents a simplified debt after min-cash-flow algorithm
type SimplifiedDebt struct {
	FromUserID     string  `json:"from_user_id"`
	FromUserNom    string  `json:"from_user_nom"`
	FromUserPrenom string  `json:"from_user_prenom"`
	FromAvatarURL  *string `json:"from_avatar_url,omitempty"`
	ToUserID       string  `json:"to_user_id"`
	ToUserNom      string  `json:"to_user_nom"`
	ToUserPrenom   string  `json:"to_user_prenom"`
	ToAvatarURL    *string `json:"to_avatar_url,omitempty"`
	Amount         float64 `json:"amount"`
}

// BalanceHistoryEntry represents an entry in balance history
type BalanceHistoryEntry struct {
	Date              time.Time `json:"date"`
	CumulativeBalance float64   `json:"cumulative_balance"`
	EventType         string    `json:"event_type"` // "expense" or "payment"
	EventID           string    `json:"event_id"`
	Description       string    `json:"description"`
	Amount            float64   `json:"amount"`
}

// Balance represents a stored balance between two users
type Balance struct {
	ID           string    `json:"id" db:"id"`
	ColocationID string    `json:"colocation_id" db:"colocation_id"`
	FromUserID   string    `json:"from_user_id" db:"from_user_id"`
	ToUserID     string    `json:"to_user_id" db:"to_user_id"`
	Amount       float64   `json:"amount" db:"amount"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
