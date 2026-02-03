package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// Connect établit une connexion à la base de données PostgreSQL
func Connect() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USERNAME", "coloc_user")
	password := getEnv("DB_PASSWORD", "coloc_password")
	dbname := getEnv("DB_NAME", "coloc_db")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'ouverture de la connexion: %w", err)
	}

	// Vérifier la connexion
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erreur lors du ping de la base de données: %w", err)
	}

	log.Println("✓ Connexion à la base de données établie")
	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
