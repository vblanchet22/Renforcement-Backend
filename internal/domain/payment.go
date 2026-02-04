package domain

import "time"

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusConfirmed PaymentStatus = "confirmed"
	PaymentStatusRejected  PaymentStatus = "rejected"
)

// Payment represents a reimbursement between two users
type Payment struct {
	ID           string        `json:"id" db:"id"`
	ColocationID string        `json:"colocation_id" db:"colocation_id"`
	FromUserID   string        `json:"from_user_id" db:"from_user_id"`
	ToUserID     string        `json:"to_user_id" db:"to_user_id"`
	Amount       float64       `json:"amount" db:"amount"`
	Status       PaymentStatus `json:"status" db:"status"`
	Note         *string       `json:"note,omitempty" db:"note"`
	ConfirmedAt  *time.Time    `json:"confirmed_at,omitempty" db:"confirmed_at"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`

	// Joined fields
	FromUserNom    string  `json:"from_user_nom,omitempty"`
	FromUserPrenom string  `json:"from_user_prenom,omitempty"`
	FromAvatarURL  *string `json:"from_avatar_url,omitempty"`
	ToUserNom      string  `json:"to_user_nom,omitempty"`
	ToUserPrenom   string  `json:"to_user_prenom,omitempty"`
	ToAvatarURL    *string `json:"to_avatar_url,omitempty"`
}
