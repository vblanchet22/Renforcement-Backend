# Backend Colocation

Projet backend avec PostgreSQL et TypeORM pour la gestion de colocation.

## Prérequis

- Node.js (v18 ou supérieur)
- Docker et Docker Compose
- npm ou yarn

## Installation

### 1. Installer les dépendances

```bash
npm install
```

### 2. Démarrer PostgreSQL avec Docker

```bash
docker compose up -d
```

Cette commande démarre un conteneur PostgreSQL avec les paramètres suivants:
- Host: localhost
- Port: 5432
- Database: coloc_db
- User: coloc_user
- Password: coloc_password

### 3. Vérifier que PostgreSQL est démarré

```bash
docker compose ps
```

Vous devriez voir le conteneur `coloc_postgres` en état "running".

### 4. Exécuter les migrations

```bash
npm run migration:run
```

Cette commande va créer la table `users` dans la base de données.

## Structure du projet

```
back_coloc/
├── src/
│   ├── config/
│   │   └── data-source.ts       # Configuration TypeORM
│   ├── entities/
│   │   └── User.ts              # Entité User
│   ├── migrations/
│   │   └── 1738592000000-CreateUserTable.ts  # Migration pour créer la table users
│   └── index.ts                 # Point d'entrée de l'application
├── docker-compose.yml           # Configuration Docker pour PostgreSQL
├── .env                         # Variables d'environnement
├── package.json
└── tsconfig.json
```

## Entité User

L'entité User contient les champs suivants:
- `id` (UUID) - Identifiant unique généré automatiquement
- `email` (string) - Email unique de l'utilisateur
- `nom` (string) - Nom de famille
- `prenom` (string) - Prénom
- `telephone` (string, optionnel) - Numéro de téléphone
- `created_at` (timestamp) - Date de création
- `updated_at` (timestamp) - Date de dernière modification

## Scripts disponibles

- `npm run dev` - Démarre l'application en mode développement
- `npm run build` - Compile le projet TypeScript
- `npm run migration:run` - Exécute les migrations en attente
- `npm run migration:revert` - Annule la dernière migration
- `npm run migration:generate -- -n NomDeLaMigration` - Génère une nouvelle migration

## Commandes Docker utiles

```bash
# Démarrer PostgreSQL
docker compose up -d

# Arrêter PostgreSQL
docker compose down

# Voir les logs de PostgreSQL
docker compose logs -f postgres

# Se connecter à PostgreSQL
docker compose exec postgres psql -U coloc_user -d coloc_db

# Supprimer les volumes (⚠️ supprime toutes les données)
docker compose down -v
```

## Variables d'environnement

Les variables d'environnement sont stockées dans le fichier `.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=coloc_user
DB_PASSWORD=coloc_password
DB_NAME=coloc_db
```

## Développement

Pour développer, lancez:

```bash
npm run dev
```

Cela démarrera l'application qui se connectera à la base de données PostgreSQL.
