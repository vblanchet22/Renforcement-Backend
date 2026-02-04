package handler

import (
	"encoding/json"
	"net/http"

	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
	"github.com/vblanchet22/back_coloc/internal/utils"
)

// UserHandler manages HTTP requests for users
type UserHandler struct {
	repo *postgres.UserRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(repo *postgres.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// UserResponse represents a user with French-formatted dates
type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Nom       string  `json:"nom"`
	Prenom    string  `json:"prenom"`
	Telephone *string `json:"telephone,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func toUserResponse(user *domain.User) UserResponse {
	// Extraire le timestamp depuis l'ULID (démontre la fonctionnalité des ULIDs)
	createdAt := ""
	if createdTime, err := utils.ULIDToTime(user.ID); err == nil {
		createdAt = utils.FormatFrenchDateTime(createdTime)
	}

	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Nom:       user.Nom,
		Prenom:    user.Prenom,
		Telephone: user.Telephone,
		CreatedAt: createdAt, // Extrait depuis l'ULID
		UpdatedAt: utils.FormatFrenchDateTime(user.UpdatedAt),
	}
}

// GetAllUsers retrieves all users
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Methode non autorisee", http.StatusMethodNotAllowed)
		return
	}

	users, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Erreur lors de la recuperation des utilisateurs", http.StatusInternalServerError)
		return
	}

	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = toUserResponse(&user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserByID retrieves a user by ID
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Methode non autorisee", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/api/users/"):]
	if id == "" {
		http.Error(w, "ID manquant", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Erreur lors de la recuperation de l'utilisateur", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "Utilisateur non trouve", http.StatusNotFound)
		return
	}

	response := toUserResponse(user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateUserRequest represents the data to create a user
type CreateUserRequest struct {
	Email     string  `json:"email"`
	Nom       string  `json:"nom"`
	Prenom    string  `json:"prenom"`
	Telephone *string `json:"telephone,omitempty"`
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methode non autorisee", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Donnees invalides", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Nom == "" || req.Prenom == "" {
		http.Error(w, "Email, nom et prenom sont obligatoires", http.StatusBadRequest)
		return
	}

	user := &domain.User{
		Email:     req.Email,
		Nom:       req.Nom,
		Prenom:    req.Prenom,
		Telephone: req.Telephone,
	}

	if err := h.repo.Create(r.Context(), user); err != nil {
		http.Error(w, "Erreur lors de la creation de l'utilisateur", http.StatusInternalServerError)
		return
	}

	response := toUserResponse(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateUserRequest represents the data to update a user
type UpdateUserRequest struct {
	Email     string  `json:"email"`
	Nom       string  `json:"nom"`
	Prenom    string  `json:"prenom"`
	Telephone *string `json:"telephone,omitempty"`
}

// UpdateUser updates a user
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Methode non autorisee", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/api/users/"):]
	if id == "" {
		http.Error(w, "ID manquant", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Donnees invalides", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Nom == "" || req.Prenom == "" {
		http.Error(w, "Email, nom et prenom sont obligatoires", http.StatusBadRequest)
		return
	}

	existingUser, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Erreur lors de la recuperation de l'utilisateur", http.StatusInternalServerError)
		return
	}
	if existingUser == nil {
		http.Error(w, "Utilisateur non trouve", http.StatusNotFound)
		return
	}

	user := &domain.User{
		ID:        id,
		Email:     req.Email,
		Nom:       req.Nom,
		Prenom:    req.Prenom,
		Telephone: req.Telephone,
	}

	if err := h.repo.Update(r.Context(), user); err != nil {
		http.Error(w, "Erreur lors de la mise a jour de l'utilisateur", http.StatusInternalServerError)
		return
	}

	updatedUser, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Erreur lors de la recuperation de l'utilisateur mis a jour", http.StatusInternalServerError)
		return
	}

	response := toUserResponse(updatedUser)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Methode non autorisee", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/api/users/"):]
	if id == "" {
		http.Error(w, "ID manquant", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		http.Error(w, "Erreur lors de la suppression de l'utilisateur", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
