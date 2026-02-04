package domain

import "time"

// DecisionStatus represents the status of a decision
type DecisionStatus string

const (
	DecisionStatusOpen   DecisionStatus = "open"
	DecisionStatusClosed DecisionStatus = "closed"
)

// Decision represents a collective decision/poll
type Decision struct {
	ID            string         `json:"id" db:"id"`
	ColocationID  string         `json:"colocation_id" db:"colocation_id"`
	CreatedBy     string         `json:"created_by" db:"created_by"`
	Title         string         `json:"title" db:"title"`
	Description   *string        `json:"description,omitempty" db:"description"`
	Options       []string       `json:"options" db:"options"` // JSONB
	Status        DecisionStatus `json:"status" db:"status"`
	Deadline      *time.Time     `json:"deadline,omitempty" db:"deadline"`
	AllowMultiple bool           `json:"allow_multiple" db:"allow_multiple"`
	IsAnonymous   bool           `json:"is_anonymous" db:"is_anonymous"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`

	// Joined fields
	CreatedByNom    string `json:"created_by_nom,omitempty"`
	CreatedByPrenom string `json:"created_by_prenom,omitempty"`
	VoteCount       int    `json:"vote_count"`
	HasVoted        bool   `json:"has_voted"`
	UserVotes       []int  `json:"user_votes,omitempty"`
}

// DecisionVote represents a vote on a decision
type DecisionVote struct {
	ID          string    `json:"id" db:"id"`
	DecisionID  string    `json:"decision_id" db:"decision_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	OptionIndex int       `json:"option_index" db:"option_index"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// OptionResult represents the result for one option
type OptionResult struct {
	OptionIndex int      `json:"option_index"`
	OptionText  string   `json:"option_text"`
	VoteCount   int      `json:"vote_count"`
	Percentage  float64  `json:"percentage"`
	Voters      []Voter  `json:"voters,omitempty"`
}

// Voter represents a voter
type Voter struct {
	UserID     string `json:"user_id"`
	UserNom    string `json:"user_nom"`
	UserPrenom string `json:"user_prenom"`
}
