package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

// AuthService handles authentication business logic
type AuthService struct {
	repo       *postgres.AuthRepository
	jwtManager *auth.JWTManager
}

// NewAuthService creates a new AuthService
func NewAuthService(repo *postgres.AuthRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

// AuthResult contains authentication tokens and user info
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         *domain.User
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, email, password, nom, prenom string, telephone *string) (*AuthResult, error) {
	// Check if email already exists
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification de l'email: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("cet email est deja utilise")
	}

	// Hash password
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Email:        email,
		PasswordHash: &hash,
		Nom:          nom,
		Prenom:       prenom,
		Telephone:    telephone,
		IsActive:     true,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	return s.generateTokens(ctx, user)
}

// Login authenticates a user with email and password
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la connexion: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("email ou mot de passe incorrect")
	}

	// Check password
	if user.PasswordHash == nil || !auth.CheckPassword(password, *user.PasswordHash) {
		return nil, fmt.Errorf("email ou mot de passe incorrect")
	}

	// Generate tokens
	return s.generateTokens(ctx, user)
}

// RefreshToken generates new tokens using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Get refresh token from database
	rt, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la verification du token: %w", err)
	}
	if rt == nil {
		return nil, fmt.Errorf("token invalide ou expire")
	}

	// Get user
	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation de l'utilisateur: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("utilisateur introuvable")
	}

	// Delete old refresh token
	if err := s.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	// Generate new tokens
	return s.generateTokens(ctx, user)
}

// Logout invalidates a refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.DeleteRefreshToken(ctx, refreshToken)
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *AuthService) generateTokens(ctx context.Context, user *domain.User) (*AuthResult, error) {
	// Generate access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la generation du token d'acces: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la generation du refresh token: %w", err)
	}

	// Save refresh token
	expiresAt := time.Now().Add(s.jwtManager.RefreshTokenExpiry())
	if err := s.repo.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtManager.AccessTokenExpiry(),
		User:         user,
	}, nil
}
