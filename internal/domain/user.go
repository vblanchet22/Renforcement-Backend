package domain

import "time"

// User represents an application user
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash *string   `json:"-" db:"password_hash"`
	Nom          string    `json:"nom" db:"nom"`
	Prenom       string    `json:"prenom" db:"prenom"`
	Telephone    *string   `json:"telephone,omitempty" db:"telephone"`
	AvatarURL    *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RefreshToken represents a stored refresh token
type RefreshToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
