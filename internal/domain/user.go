package domain

import "time"

// User représente un utilisateur de l'application
type User struct {
	ID        string    `json:"id" db:"id"`          // ULID (26 chars) - contient le timestamp de création
	Email     string    `json:"email" db:"email"`
	Nom       string    `json:"nom" db:"nom"`
	Prenom    string    `json:"prenom" db:"prenom"`
	Telephone *string   `json:"telephone,omitempty" db:"telephone"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	// Note: CreatedAt supprimé - le timestamp est inclus dans l'ULID
}
