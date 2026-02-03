package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/config"
)

// Connect establishes a connection pool to the PostgreSQL database
func Connect(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la creation du pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("erreur lors du ping de la base de donnees: %w", err)
	}

	log.Println("Connexion a la base de donnees etablie")
	return pool, nil
}
