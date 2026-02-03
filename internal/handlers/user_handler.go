package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/vblanchet22/back_coloc/internal/models"
	"github.com/vblanchet22/back_coloc/internal/repository"
	"github.com/vblanchet22/back_coloc/internal/utils"
)

// UserHandler gère les requêtes HTTP liées aux utilisateurs
type UserHandler struct {
	repo *repository.UserRepository
}

// NewUserHandler crée un nouveau gestionnaire d'utilisateurs
func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// UserResponse représente un utilisateur avec les dates formatées en français
type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Nom       string  `json:"nom"`
	Prenom    string  `json:"prenom"`
	Telephone *string `json:"telephone,omitempty"`
	CreatedAt string  `json:"created_at"` // Format français: "21/12/2025 14:30:45"
	UpdatedAt string  `json:"updated_at"` // Format français: "21/12/2025 14:30:45"
}

// toUserResponse convertit un modèle User en UserResponse avec dates françaises
func toUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Nom:       user.Nom,
		Prenom:    user.Prenom,
		Telephone: user.Telephone,
		CreatedAt: utils.FormatFrenchDateTime(user.CreatedAt),
		UpdatedAt: utils.FormatFrenchDateTime(user.UpdatedAt),
	}
}

// GetAllUsers récupère tous les utilisateurs
// GET /api/users
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	users, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des utilisateurs", http.StatusInternalServerError)
		return
	}

	// Convertir en UserResponse avec dates françaises
	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = toUserResponse(&user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserByID récupère un utilisateur par son ID
// GET /api/users/{id}
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Extraire l'ID depuis le path
	id := r.URL.Path[len("/api/users/"):]
	if id == "" {
		http.Error(w, "ID manquant", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de l'utilisateur", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
		return
	}

	response := toUserResponse(user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateUserRequest représente les données pour créer un utilisateur
type CreateUserRequest struct {
	Email     string  `json:"email"`
	Nom       string  `json:"nom"`
	Prenom    string  `json:"prenom"`
	Telephone *string `json:"telephone,omitempty"`
}

// CreateUser crée un nouvel utilisateur
// POST /api/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Validation basique
	if req.Email == "" || req.Nom == "" || req.Prenom == "" {
		http.Error(w, "Email, nom et prénom sont obligatoires", http.StatusBadRequest)
		return
	}

	user := &models.User{
		Email:     req.Email,
		Nom:       req.Nom,
		Prenom:    req.Prenom,
		Telephone: req.Telephone,
	}

	if err := h.repo.Create(user); err != nil {
		http.Error(w, "Erreur lors de la création de l'utilisateur", http.StatusInternalServerError)
		return
	}

	response := toUserResponse(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateUserRequest représente les données pour mettre à jour un utilisateur
type UpdateUserRequest struct {
	Email     string  `json:"email"`
	Nom       string  `json:"nom"`
	Prenom    string  `json:"prenom"`
	Telephone *string `json:"telephone,omitempty"`
}

// UpdateUser met à jour un utilisateur
// PUT /api/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Extraire l'ID depuis le path
	id := r.URL.Path[len("/api/users/"):]
	if id == "" {
		http.Error(w, "ID manquant", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Validation basique
	if req.Email == "" || req.Nom == "" || req.Prenom == "" {
		http.Error(w, "Email, nom et prénom sont obligatoires", http.StatusBadRequest)
		return
	}

	// Vérifier que l'utilisateur existe
	existingUser, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de l'utilisateur", http.StatusInternalServerError)
		return
	}
	if existingUser == nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
		return
	}

	user := &models.User{
		ID:        id,
		Email:     req.Email,
		Nom:       req.Nom,
		Prenom:    req.Prenom,
		Telephone: req.Telephone,
	}

	if err := h.repo.Update(user); err != nil {
		http.Error(w, "Erreur lors de la mise à jour de l'utilisateur", http.StatusInternalServerError)
		return
	}

	// Récupérer l'utilisateur mis à jour pour avoir toutes les infos
	updatedUser, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de l'utilisateur mis à jour", http.StatusInternalServerError)
		return
	}

	response := toUserResponse(updatedUser)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteUser supprime un utilisateur
// DELETE /api/users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Extraire l'ID depuis le path
	id := r.URL.Path[len("/api/users/"):]
	if id == "" {
		http.Error(w, "ID manquant", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		http.Error(w, "Erreur lors de la suppression de l'utilisateur", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
