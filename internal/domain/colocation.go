package domain

import "time"

// Colocation represents a shared housing
type Colocation struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	Address     *string   `json:"address,omitempty" db:"address"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	InviteCode  string    `json:"invite_code" db:"invite_code"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ColocationMember represents a member of a colocation
type ColocationMember struct {
	ID           string    `json:"id" db:"id"`
	ColocationID string    `json:"colocation_id" db:"colocation_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Role         string    `json:"role" db:"role"` // "admin" or "member"
	JoinedAt     time.Time `json:"joined_at" db:"joined_at"`
	// User details (joined)
	Email     string  `json:"email" db:"email"`
	Nom       string  `json:"nom" db:"nom"`
	Prenom    string  `json:"prenom" db:"prenom"`
	AvatarURL *string `json:"avatar_url,omitempty" db:"avatar_url"`
}

// ColocationInvitation represents an invitation to join a colocation
type ColocationInvitation struct {
	ID           string    `json:"id" db:"id"`
	ColocationID string    `json:"colocation_id" db:"colocation_id"`
	InvitedBy    string    `json:"invited_by" db:"invited_by"`
	InvitedEmail string    `json:"invited_email" db:"invited_email"`
	Status       string    `json:"status" db:"status"` // "pending", "accepted", "rejected", "expired"
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

const (
	RoleAdmin  = "admin"
	RoleMember = "member"

	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusRejected = "rejected"
	InvitationStatusExpired  = "expired"
)
