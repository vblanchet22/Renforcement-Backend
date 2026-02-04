package utils

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// GenerateULID génère un nouvel ULID (Universally Unique Lexicographically Sortable Identifier)
// Format: 26 caractères en base32 Crockford
// Structure: 10 chars timestamp (48 bits ms) + 16 chars random (80 bits)
func GenerateULID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.Reader, 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return id.String()
}

// ParseULID parse une chaîne ULID et retourne un objet ulid.ULID
func ParseULID(s string) (ulid.ULID, error) {
	return ulid.Parse(s)
}

// ULIDToTime extrait le timestamp d'un ULID
func ULIDToTime(s string) (time.Time, error) {
	id, err := ulid.Parse(s)
	if err != nil {
		return time.Time{}, err
	}
	ms := id.Time()
	return time.UnixMilli(int64(ms)), nil
}
