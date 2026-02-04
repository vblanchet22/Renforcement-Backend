# Backend Colocation - Go + PostgreSQL

Projet backend en Go avec PostgreSQL pour la gestion de colocation.

## Prérequis

- Go 1.21 ou supérieur
- Docker
- golang-migrate (installé via Homebrew)

## Installation rapide

### 1. Installer les dépendances Go

```bash
make deps
```

### 2. Démarrer PostgreSQL

PostgreSQL tourne déjà sur votre machine. Pour vérifier :

```bash
docker ps | grep postgres
```

Si vous devez le redémarrer :

```bash
make docker-up
```

### 3. Exécuter les migrations

```bash
make migrate-up
```

Cette commande crée la table `users` dans la base de données.

### 4. Lancer l'application

```bash
make run
```

## Structure du projet

```
back_coloc/
├── cmd/
│   └── server/
│       └── main.go              # Point d'entrée de l'application
├── internal/
│   ├── models/
│   │   └── user.go              # Modèle User
│   └── database/
│       └── database.go          # Connexion à la base de données
├── migrations/
│   ├── 000001_create_users_table.up.sql    # Migration création table
│   └── 000001_create_users_table.down.sql  # Migration rollback
├── docker-compose.yml           # Configuration Docker PostgreSQL
├── .env                         # Variables d'environnement
├── Makefile                     # Commandes utiles
└── go.mod                       # Dépendances Go
```

## Modèle User

```go
type User struct {
    ID        string    // UUID généré automatiquement
    Email     string    // Email unique
    Nom       string    // Nom de famille
    Prenom    string    // Prénom
    Telephone *string   // Numéro de téléphone (optionnel)
    CreatedAt time.Time // Date de création
    UpdatedAt time.Time // Date de modification
}
```

## Commandes disponibles (Makefile)

```bash
make help          # Affiche toutes les commandes disponibles
make docker-up     # Démarre PostgreSQL avec Docker
make docker-down   # Arrête PostgreSQL
make migrate-up    # Applique les migrations
make migrate-down  # Annule la dernière migration
make run           # Lance l'application
make build         # Compile l'application
make test          # Lance les tests
make deps          # Installe les dépendances
```

## Données de démonstration

Vous pouvez remplir rapidement la dernière colocation créée (ou une colocation précise) avec des membres, catégories et dépenses de test :

```bash
# utilise la dernière colocation créée
go run cmd/seed/main.go

# cible une colocation précise (ID visible dans la base/API)
go run cmd/seed/main.go --colocation-id=<ID_DE_LA_COLOCATION>
```

Le script lit la configuration `DB_*`, ajoute des membres démo si besoin puis crée des dépenses `[DEMO]` cohérentes. Il nettoie d'abord les anciennes dépenses marquées `[DEMO]` pour éviter les doublons.

## Migrations

### Créer une nouvelle migration

```bash
make migrate-create NAME=nom_de_la_migration
```

### Appliquer les migrations

```bash
make migrate-up
```

### Annuler une migration

```bash
make migrate-down
```

## Variables d'environnement

Le fichier `.env` contient :

```env
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=coloc_user
DB_PASSWORD=coloc_password
DB_NAME=coloc_db
```

## Connexion à la base de données

Pour vous connecter manuellement à PostgreSQL :

```bash
docker compose exec postgres psql -U coloc_user -d coloc_db
```

Ou avec psql local :

```bash
psql -h localhost -p 5432 -U coloc_user -d coloc_db
```

## Docker (sans Docker Desktop)

Si Docker Desktop ne fonctionne pas, Docker CLI fonctionne quand même :

```bash
# Vérifier que Docker tourne
docker ps

# Démarrer PostgreSQL
docker compose up -d

# Voir les logs
docker compose logs -f postgres

# Arrêter
docker compose down
```

## Développement

1. Modifier le code dans `cmd/server/main.go` ou créer de nouveaux handlers
2. Lancer avec `make run`
3. Pour ajouter une nouvelle table, créer une migration avec `make migrate-create`
