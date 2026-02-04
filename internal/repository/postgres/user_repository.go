package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/domain"
)

// UserRepository manages user database operations
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// GetAll retrieves all users
func (r *UserRepository) GetAll(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, email, nom, prenom, telephone, updated_at
		FROM users
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation des utilisateurs: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Nom,
			&user.Prenom,
			&user.Telephone,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du scan d'un utilisateur: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'iteration des resultats: %w", err)
	}

	return users, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, nom, prenom, telephone, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Nom,
		&user.Prenom,
		&user.Telephone,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation de l'utilisateur: %w", err)
	}

	return &user, nil
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, nom, prenom, telephone)
		VALUES ($1, $2, $3, $4)
		RETURNING id, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		user.Email,
		user.Nom,
		user.Prenom,
		user.Telephone,
	).Scan(&user.ID, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("erreur lors de la creation de l'utilisateur: %w", err)
	}

	return nil
}

// Update modifies an existing user
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, nom = $2, prenom = $3, telephone = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		user.Email,
		user.Nom,
		user.Prenom,
		user.Telephone,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("erreur lors de la mise a jour de l'utilisateur: %w", err)
	}

	return nil
}

// Delete removes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression de l'utilisateur: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("aucun utilisateur trouve avec l'ID %s", id)
	}

	return nil
}
