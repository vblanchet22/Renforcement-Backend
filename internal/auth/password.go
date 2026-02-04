package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the bcrypt cost factor
	DefaultCost = 12
)

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", fmt.Errorf("le mot de passe doit contenir au moins 8 caracteres")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("erreur lors du hashage du mot de passe: %w", err)
	}

	return string(hash), nil
}

// CheckPassword compares a password with a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
