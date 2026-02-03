package domain

import "time"

// User représente un utilisateur de l'application
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Nom       string    `json:"nom" db:"nom"`
	Prenom    string    `json:"prenom" db:"prenom"`
	Telephone *string   `json:"telephone,omitempty" db:"telephone"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
