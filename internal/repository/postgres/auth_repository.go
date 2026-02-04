package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/domain"
)

// AuthRepository manages authentication database operations
type AuthRepository struct {
	pool *pgxpool.Pool
}

// NewAuthRepository creates a new AuthRepository instance
func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

// CreateUser creates a new user with password
func (r *AuthRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, nom, prenom, telephone, avatar_url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
		user.Nom,
		user.Prenom,
		user.Telephone,
		user.AvatarURL,
		user.IsActive,
	).Scan(&user.ID, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("erreur lors de la creation de l'utilisateur: %w", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, nom, prenom, telephone, avatar_url, is_active, updated_at
		FROM users
		WHERE email = $1 AND is_active = true
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Nom,
		&user.Prenom,
		&user.Telephone,
		&user.AvatarURL,
		&user.IsActive,
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

// GetUserByID retrieves a user by ID
func (r *AuthRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, nom, prenom, telephone, avatar_url, is_active, updated_at
		FROM users
		WHERE id = $1 AND is_active = true
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Nom,
		&user.Prenom,
		&user.Telephone,
		&user.AvatarURL,
		&user.IsActive,
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

// EmailExists checks if an email is already registered
func (r *AuthRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("erreur lors de la verification de l'email: %w", err)
	}

	return exists, nil
}

// SaveRefreshToken stores a new refresh token
func (r *AuthRepository) SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("erreur lors de la sauvegarde du refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves a refresh token
func (r *AuthRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1 AND expires_at > NOW()
	`

	var rt domain.RefreshToken
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation du refresh token: %w", err)
	}

	return &rt, nil
}

// DeleteRefreshToken removes a refresh token
func (r *AuthRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`

	_, err := r.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du refresh token: %w", err)
	}

	return nil
}

// DeleteUserRefreshTokens removes all refresh tokens for a user
func (r *AuthRepository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression des refresh tokens: %w", err)
	}

	return nil
}

// CleanExpiredTokens removes expired refresh tokens
func (r *AuthRepository) CleanExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("erreur lors du nettoyage des tokens expires: %w", err)
	}

	return nil
}

// UpdateUser updates a user's profile
func (r *AuthRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET nom = $1, prenom = $2, telephone = $3, avatar_url = $4, password_hash = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		user.Nom,
		user.Prenom,
		user.Telephone,
		user.AvatarURL,
		user.PasswordHash,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("erreur lors de la mise a jour de l'utilisateur: %w", err)
	}

	return nil
}

// DeactivateUser sets a user's is_active to false (soft delete)
func (r *AuthRepository) DeactivateUser(ctx context.Context, userID string) error {
	query := `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la desactivation de l'utilisateur: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("utilisateur introuvable")
	}

	return nil
}
