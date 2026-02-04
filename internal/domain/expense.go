package domain

import "time"

// SplitType defines how an expense is split among members
type SplitType string

const (
	SplitTypeEqual      SplitType = "equal"
	SplitTypePercentage SplitType = "percentage"
	SplitTypeCustom     SplitType = "custom"
)

// Recurrence defines how often a recurring expense repeats
type Recurrence string

const (
	RecurrenceDaily   Recurrence = "daily"
	RecurrenceWeekly  Recurrence = "weekly"
	RecurrenceMonthly Recurrence = "monthly"
	RecurrenceYearly  Recurrence = "yearly"
)

// Expense represents an expense in a colocation
type Expense struct {
	ID           string     `json:"id" db:"id"`
	ColocationID string     `json:"colocation_id" db:"colocation_id"`
	PaidBy       string     `json:"paid_by" db:"paid_by"`
	CategoryID   string     `json:"category_id" db:"category_id"`
	Title        string     `json:"title" db:"title"`
	Description  *string    `json:"description,omitempty" db:"description"`
	Amount       float64    `json:"amount" db:"amount"`
	SplitType    SplitType  `json:"split_type" db:"split_type"`
	ExpenseDate  time.Time  `json:"expense_date" db:"expense_date"`
	RecurringID  *string    `json:"recurring_id,omitempty" db:"recurring_id"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`

	// Joined fields
	PaidByNom     string `json:"paid_by_nom,omitempty"`
	PaidByPrenom  string `json:"paid_by_prenom,omitempty"`
	CategoryName  string `json:"category_name,omitempty"`
	Splits        []ExpenseSplit `json:"splits,omitempty"`
}

// ExpenseSplit represents how an expense is split for a specific user
type ExpenseSplit struct {
	ID        string  `json:"id" db:"id"`
	ExpenseID string  `json:"expense_id" db:"expense_id"`
	UserID    string  `json:"user_id" db:"user_id"`
	Amount    float64 `json:"amount" db:"amount"`
	Percentage float64 `json:"percentage" db:"percentage"`
	IsSettled bool    `json:"is_settled" db:"is_settled"`

	// Joined fields
	UserNom    string `json:"user_nom,omitempty"`
	UserPrenom string `json:"user_prenom,omitempty"`
}

// RecurringExpense represents a recurring expense template
type RecurringExpense struct {
	ID           string     `json:"id" db:"id"`
	ColocationID string     `json:"colocation_id" db:"colocation_id"`
	PaidBy       string     `json:"paid_by" db:"paid_by"`
	CategoryID   string     `json:"category_id" db:"category_id"`
	Title        string     `json:"title" db:"title"`
	Description  *string    `json:"description,omitempty" db:"description"`
	Amount       float64    `json:"amount" db:"amount"`
	SplitType    SplitType  `json:"split_type" db:"split_type"`
	Recurrence   Recurrence `json:"recurrence" db:"recurrence"`
	NextDueDate  time.Time  `json:"next_due_date" db:"next_due_date"`
	EndDate      *time.Time `json:"end_date,omitempty" db:"end_date"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`

	// Joined fields
	PaidByNom    string                 `json:"paid_by_nom,omitempty"`
	PaidByPrenom string                 `json:"paid_by_prenom,omitempty"`
	CategoryName string                 `json:"category_name,omitempty"`
	Splits       []RecurringExpenseSplit `json:"splits,omitempty"`
}

// RecurringExpenseSplit represents the split percentage for a recurring expense
type RecurringExpenseSplit struct {
	ID          string  `json:"id" db:"id"`
	RecurringID string  `json:"recurring_id" db:"recurring_id"`
	UserID      string  `json:"user_id" db:"user_id"`
	Percentage  float64 `json:"percentage" db:"percentage"`

	// Joined fields
	UserNom    string `json:"user_nom,omitempty"`
	UserPrenom string `json:"user_prenom,omitempty"`
}

// ExpenseSplitInput is used when creating/updating an expense split
type ExpenseSplitInput struct {
	UserID     string  `json:"user_id"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

// MonthlyForecast represents a forecast for a specific month
type MonthlyForecast struct {
	Month       string            `json:"month"` // Format: YYYY-MM
	TotalAmount float64           `json:"total_amount"`
	Categories  []CategoryForecast `json:"categories"`
}

// CategoryForecast represents a forecast for a specific category
type CategoryForecast struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount       float64 `json:"amount"`
}
