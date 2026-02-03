package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
	"github.com/vblanchet22/back_coloc/internal/config"
	handler "github.com/vblanchet22/back_coloc/internal/grpc"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aucun fichier .env trouve, utilisation des valeurs par defaut")
	}

	cfg := config.Load()

	pool, err := postgres.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Erreur de connexion a la base de donnees: %v", err)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepository(pool)
	userHandler := handler.NewUserHandler(userRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetAllUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "Methode non autorisee", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "Methode non autorisee", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := cfg.Server.HTTPPort

	log.Printf("Serveur demarre sur le port %s", port)
	log.Printf("API disponible sur http://localhost:%s/api/users", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Erreur lors du demarrage du serveur: %v", err)
	}
}
