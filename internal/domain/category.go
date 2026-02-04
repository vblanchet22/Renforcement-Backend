package domain

import "time"

// ExpenseCategory represents an expense category
type ExpenseCategory struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Icon         *string   `json:"icon,omitempty" db:"icon"`
	Color        *string   `json:"color,omitempty" db:"color"`
	ColocationID *string   `json:"colocation_id,omitempty" db:"colocation_id"` // NULL = global category
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// IsGlobal returns true if this is a global category (not colocation-specific)
func (c *ExpenseCategory) IsGlobal() bool {
	return c.ColocationID == nil
}

// CategoryStat represents statistics for a category
type CategoryStat struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Icon         *string `json:"icon,omitempty"`
	Color        *string `json:"color,omitempty"`
	TotalAmount  float64 `json:"total_amount"`
	ExpenseCount int     `json:"expense_count"`
	Percentage   float64 `json:"percentage"`
}
