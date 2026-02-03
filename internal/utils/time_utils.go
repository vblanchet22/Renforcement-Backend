package utils

import (
	"time"
)

// ParisTZ représente le fuseau horaire de Paris (Europe/Paris)
var ParisTZ *time.Location

func init() {
	// Charger le fuseau horaire Europe/Paris
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		// Fallback en cas d'erreur
		loc = time.FixedZone("CET", 1*60*60) // UTC+1
	}
	ParisTZ = loc
}

// FormatFrenchDateTime formate une date au format français "21/12/2025 14:30:45"
// en convertissant depuis UTC vers le fuseau horaire français (UTC+1 ou UTC+2)
func FormatFrenchDateTime(t time.Time) string {
	// Convertir depuis UTC vers Europe/Paris
	parisTime := t.In(ParisTZ)
	// Format: jour/mois/année heure:minute:seconde
	return parisTime.Format("02/01/2006 15:04:05")
}

// FormatFrenchDate formate une date au format français "21/12/2025"
func FormatFrenchDate(t time.Time) string {
	parisTime := t.In(ParisTZ)
	return parisTime.Format("02/01/2006")
}
