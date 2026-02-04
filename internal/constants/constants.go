// Package constants defines named constants used throughout the application.
// This avoids magic numbers and makes the codebase more maintainable.
package constants

import "time"

// Pagination defaults and limits
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// JWT token expiry defaults
const (
	DefaultAccessTokenExpiry  = 24 * time.Hour  // 1 day
	DefaultRefreshTokenExpiry = 168 * time.Hour // 7 days
)

// Percentage calculation constants
const (
	PercentageBase     = 100.0
	PercentageTolerance = 0.01
	PercentageMinBound  = 99.99
	PercentageMaxBound  = 100.01
)

// Amount validation constants
const (
	AmountTolerance     = 0.01 // 1 cent tolerance for floating point comparisons
	MinPositiveAmount   = 0.01 // Minimum amount for debt/payment
)

// Forecast defaults
const (
	DefaultForecastMonths = 3
)

// Channel buffer sizes
const (
	NotificationChannelBuffer = 100
)
