package repository

import (
	"database/sql"
	"fmt"

	"github.com/vblanchet22/back_coloc/internal/models"
)

// UserRepository gère les opérations sur les utilisateurs
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository crée une nouvelle instance de UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetAll récupère tous les utilisateurs
func (r *UserRepository) GetAll() ([]models.User, error) {
	query := `
		SELECT id, email, nom, prenom, telephone, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des utilisateurs: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Nom,
			&user.Prenom,
			&user.Telephone,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du scan d'un utilisateur: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération des résultats: %w", err)
	}

	return users, nil
}

// GetByID récupère un utilisateur par son ID
func (r *UserRepository) GetByID(id string) (*models.User, error) {
	query := `
		SELECT id, email, nom, prenom, telephone, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Nom,
		&user.Prenom,
		&user.Telephone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Pas d'erreur, mais aucun utilisateur trouvé
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur: %w", err)
	}

	return &user, nil
}

// Create crée un nouvel utilisateur
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (email, nom, prenom, telephone)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		user.Email,
		user.Nom,
		user.Prenom,
		user.Telephone,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("erreur lors de la création de l'utilisateur: %w", err)
	}

	return nil
}

// Update met à jour un utilisateur existant
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET email = $1, nom = $2, prenom = $3, telephone = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		user.Email,
		user.Nom,
		user.Prenom,
		user.Telephone,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de l'utilisateur: %w", err)
	}

	return nil
}

// Delete supprime un utilisateur
func (r *UserRepository) Delete(id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression de l'utilisateur: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification des lignes affectées: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("aucun utilisateur trouvé avec l'ID %s", id)
	}

	return nil
}
