package domain

import "time"

// CommonFund represents a common fund/pot
type CommonFund struct {
	ID            string   `json:"id" db:"id"`
	ColocationID  string   `json:"colocation_id" db:"colocation_id"`
	Name          string   `json:"name" db:"name"`
	Description   *string  `json:"description,omitempty" db:"description"`
	TargetAmount  *float64 `json:"target_amount,omitempty" db:"target_amount"`
	CurrentAmount float64  `json:"current_amount" db:"current_amount"`
	IsActive      bool     `json:"is_active" db:"is_active"`
	CreatedBy     string   `json:"created_by" db:"created_by"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`

	// Joined fields
	CreatedByNom       string               `json:"created_by_nom,omitempty"`
	CreatedByPrenom    string               `json:"created_by_prenom,omitempty"`
	ProgressPercentage float64              `json:"progress_percentage"`
	Contributors       []ContributorSummary `json:"contributors,omitempty"`
}

// ContributorSummary represents a user's total contribution to a fund
type ContributorSummary struct {
	UserID           string  `json:"user_id"`
	UserNom          string  `json:"user_nom"`
	UserPrenom       string  `json:"user_prenom"`
	TotalContributed float64 `json:"total_contributed"`
}

// FundContribution represents a single contribution to a fund
type FundContribution struct {
	ID        string    `json:"id" db:"id"`
	FundID    string    `json:"fund_id" db:"fund_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Amount    float64   `json:"amount" db:"amount"`
	Note      *string   `json:"note,omitempty" db:"note"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Joined fields
	UserNom    string `json:"user_nom,omitempty"`
	UserPrenom string `json:"user_prenom,omitempty"`
}

// Event represents an event linked to a fund
type Event struct {
	ID           string     `json:"id" db:"id"`
	ColocationID string     `json:"colocation_id" db:"colocation_id"`
	FundID       *string    `json:"fund_id,omitempty" db:"fund_id"`
	CreatedBy    string     `json:"created_by" db:"created_by"`
	Title        string     `json:"title" db:"title"`
	Description  *string    `json:"description,omitempty" db:"description"`
	Budget       *float64   `json:"budget,omitempty" db:"budget"`
	EventDate    *time.Time `json:"event_date,omitempty" db:"event_date"`
	Location     *string    `json:"location,omitempty" db:"location"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`

	// Joined fields
	CreatedByNom    string             `json:"created_by_nom,omitempty"`
	CreatedByPrenom string             `json:"created_by_prenom,omitempty"`
	Participants    []EventParticipant `json:"participants,omitempty"`
}

// EventParticipant represents a participant in an event
type EventParticipant struct {
	ID         string    `json:"id" db:"id"`
	EventID    string    `json:"event_id" db:"event_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	RSVP       string    `json:"rsvp" db:"rsvp"` // "yes", "no", "maybe"
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UserNom    string    `json:"user_nom,omitempty"`
	UserPrenom string    `json:"user_prenom,omitempty"`
}
