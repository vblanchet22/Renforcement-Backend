.PHONY: help migrate-up migrate-down migrate-create run docker-up docker-down

# Variables
DB_URL=postgresql://coloc_user:coloc_password@localhost:5432/coloc_db?sslmode=disable

help: ## Affiche cette aide
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

docker-up: ## Démarre PostgreSQL avec Docker
	docker compose up -d
	@echo "Attente de PostgreSQL..."
	@sleep 3
	docker compose ps

docker-down: ## Arrête PostgreSQL
	docker compose down

migrate-up: ## Exécute les migrations (applique les changements)
	migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Annule la dernière migration
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-create: ## Crée une nouvelle migration (usage: make migrate-create NAME=nom_migration)
	migrate create -ext sql -dir migrations -seq $(NAME)

run: ## Lance l'application
	go run cmd/server/main.go

build: ## Compile l'application
	go build -o bin/server cmd/server/main.go

test: ## Lance les tests
	go test -v ./...

deps: ## Installe les dépendances Go
	go mod download
	go mod tidy
