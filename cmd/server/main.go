package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/vblanchet22/back_coloc/internal/database"
	"github.com/vblanchet22/back_coloc/internal/handlers"
	"github.com/vblanchet22/back_coloc/internal/repository"
)

func main() {
	// Charger les variables d'environnement
	if err := godotenv.Load(); err != nil {
		log.Println("Aucun fichier .env trouvé, utilisation des valeurs par défaut")
	}

	// Connexion à la base de données
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Erreur de connexion à la base de données: %v", err)
	}
	defer db.Close()

	// Initialiser le repository et les handlers
	userRepo := repository.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	// Configuration des routes
	mux := http.NewServeMux()

	// Routes API pour les utilisateurs
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetAllUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		// Vérifier qu'il y a bien un ID après /api/users/
		path := strings.TrimPrefix(r.URL.Path, "/api/users/")
		if path == "" {
			http.Error(w, "ID manquant", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			userHandler.GetUserByID(w, r)
		case http.MethodPut:
			userHandler.UpdateUser(w, r)
		case http.MethodDelete:
			userHandler.DeleteUser(w, r)
		default:
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	})

	// Route de santé
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Port du serveur
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Serveur démarré sur le port %s", port)
	log.Printf("📡 API disponible sur http://localhost:%s/api/users", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
	}
}
