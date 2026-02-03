# Vérification de la Gestion des Fuseaux Horaires (Timezone)

## Configuration

### Base de données (PostgreSQL)

**Type de colonne utilisé**: `TIMESTAMPTZ` (TIMESTAMP WITH TIME ZONE)

Les colonnes `created_at` et `updated_at` de la table `users` utilisent le type `TIMESTAMPTZ` qui :
- **Stocke toujours les dates en UTC** en interne
- Convertit automatiquement lors de l'insertion/lecture selon le timezone de la session PostgreSQL
- Garantit une cohérence des dates quelle que soit la timezone du serveur

### Application Go

**Package utilisé**: `time` (standard library)

**Fuseau horaire**: `Europe/Paris` (UTC+1 en hiver, UTC+2 en été)

**Fichier**: `internal/utils/time_utils.go`

```go
// Charge le fuseau horaire Europe/Paris au démarrage
var ParisTZ *time.Location

func init() {
    loc, err := time.LoadLocation("Europe/Paris")
    if err != nil {
        loc = time.FixedZone("CET", 1*60*60) // Fallback UTC+1
    }
    ParisTZ = loc
}
```

## Flux de Conversion

### 1. Insertion en Base de Données
```
Application (time.Now())
    ↓ [UTC par défaut dans Go]
PostgreSQL (TIMESTAMPTZ)
    ↓ [Stockage en UTC]
Base de données (UTC)
```

### 2. Lecture depuis la Base de Données
```
Base de données (UTC)
    ↓ [time.Time en UTC]
Application Go (time.Time)
    ↓ [Conversion via FormatFrenchDateTime]
Fuseau Europe/Paris (UTC+1 ou UTC+2)
    ↓ [Format "02/01/2006 15:04:05"]
API Response JSON (format français)
```

## Format de Date dans l'API

**Format retourné**: `"21/12/2025 14:30:45"` (jour/mois/année heure:minute:seconde)

**Fuseau horaire**: Europe/Paris
- **Hiver** (novembre à mars): UTC+1
- **Été** (avril à octobre): UTC+2

### Exemple de Conversion

Si un utilisateur est créé à **14:30:45 UTC** le **21 décembre 2025** :

- **Stockage BDD**: `2025-12-21 14:30:45+00` (UTC)
- **API Response** (hiver): `"21/12/2025 15:30:45"` (UTC+1)

Si un utilisateur est créé à **14:30:45 UTC** le **21 juin 2025** :

- **Stockage BDD**: `2025-06-21 14:30:45+00` (UTC)
- **API Response** (été): `"21/06/2025 16:30:45"` (UTC+2)

## Tests de Vérification

### Test 1: Vérifier le type de colonne dans PostgreSQL

```bash
# Se connecter à PostgreSQL
docker-compose exec postgres psql -U coloc_user -d coloc_db

# Vérifier le type des colonnes
\d users
```

**Résultat attendu**:
```
created_at | timestamp with time zone | not null | now()
updated_at | timestamp with time zone | not null | now()
```

### Test 2: Insérer et vérifier le stockage UTC

```sql
-- Insérer un utilisateur
INSERT INTO users (email, nom, prenom)
VALUES ('test@example.com', 'Test', 'User');

-- Vérifier l'heure stockée (doit être en UTC)
SELECT
    created_at,
    created_at AT TIME ZONE 'UTC' as utc_time,
    created_at AT TIME ZONE 'Europe/Paris' as paris_time
FROM users
WHERE email = 'test@example.com';
```

### Test 3: Tester l'API avec curl

```bash
# Créer un utilisateur
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "nom": "Dupont",
    "prenom": "Jean"
  }'

# Récupérer tous les utilisateurs
curl http://localhost:8080/api/users
```

**Réponse attendue**:
```json
[
  {
    "id": "uuid-here",
    "email": "test@example.com",
    "nom": "Dupont",
    "prenom": "Jean",
    "created_at": "03/02/2026 15:30:45",  // Heure de Paris (UTC+1 en hiver)
    "updated_at": "03/02/2026 15:30:45"
  }
]
```

### Test 4: Vérifier la différence UTC vs Paris

```bash
# Dans un terminal, vérifier l'heure actuelle
date -u  # Heure UTC

# Créer un utilisateur via l'API
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "timezone-test@example.com", "nom": "Test", "prenom": "TZ"}'

# Vérifier que l'heure retournée est bien UTC+1 (hiver) ou UTC+2 (été)
curl http://localhost:8080/api/users | jq '.[] | select(.email == "timezone-test@example.com")'
```

## Points Importants

✅ **Les dates sont TOUJOURS stockées en UTC** dans la base de données
✅ **Les dates sont TOUJOURS retournées en heure française** (Europe/Paris) dans l'API
✅ **Le changement heure d'été/hiver est géré automatiquement** par Go (time.Location)
✅ **Format français utilisé**: `"21/12/2025 14:30:45"` (et non ISO 8601)

## Migrations Appliquées

1. `000001_create_users_table` - Création de la table avec TIMESTAMP
2. `000002_add_index_created_at` - Ajout de l'index sur created_at
3. `000003_convert_timestamps_to_timestamptz` - Conversion vers TIMESTAMPTZ (UTC)

## Commandes Utiles

```bash
# Appliquer toutes les migrations
make migrate-up

# Démarrer le serveur
make run

# Vérifier les logs PostgreSQL
docker-compose logs -f postgres
```
