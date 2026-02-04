package domain

import "time"

// User represents an application user
// Note: CreatedAt is encoded in the ULID (first 10 chars = timestamp)
type User struct {
	ID           string    `json:"id" db:"id"` // ULID (26 chars) - contains creation timestamp
	Email        string    `json:"email" db:"email"`
	PasswordHash *string   `json:"-" db:"password_hash"`
	Nom          string    `json:"nom" db:"nom"`
	Prenom       string    `json:"prenom" db:"prenom"`
	Telephone    *string   `json:"telephone,omitempty" db:"telephone"`
	AvatarURL    *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	IsActive     bool      `json:"is_active" db:"is_active"`
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
