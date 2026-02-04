# TP: Migration UUID vers ULID

## 🎯 Objectifs

1. Remplacer les UUID par des ULIDs (Universally Unique Lexicographically Sortable Identifier)
2. Supprimer la colonne `created_at` (timestamp inclus dans l'ULID)
3. Mesurer les différences de performance entre UUID et ULID
4. Comprendre les avantages et inconvénients de chaque approche

## 📚 Qu'est-ce qu'un ULID ?

### Caractéristiques

- **Format**: 26 caractères en base32 Crockford (vs 36 pour UUID avec tirets)
- **Structure**:
  - 10 caractères: timestamp (48 bits = millisecondes depuis epoch)
  - 16 caractères: random (80 bits)
- **Tri**: Lexicographiquement triable par ordre chronologique
- **Unicité**: Garantie de la même manière que les UUID

### Exemple

```
01ARZ3NDEKTSV4RRFFQ69G5FAV
|------------|-----------|
  Timestamp     Random
  (10 chars)   (16 chars)
```

### Avantages vs UUID

| Critère | UUID v4 | ULID |
|---------|---------|------|
| Taille (texte) | 36 chars | 26 chars |
| Tri chronologique | ❌ Non | ✅ Oui |
| Index B-tree | Fragmentation | Séquentiel |
| Lisibilité | Moyenne | Meilleure |
| Contient timestamp | ❌ Non | ✅ Oui (48-bit ms) |

## 🚀 Étapes du TP

### 1. Créer la branche

```bash
git checkout -b tp-ulid
```

✅ **Fait !** Vous êtes maintenant sur la branche `tp-ulid`

### 2. Appliquer la migration ULID

La migration `000013` va :
- Créer la fonction `generate_ulid()` en PL/pgSQL
- Convertir la colonne `id` de UUID vers TEXT(26)
- Supprimer la colonne `created_at` (info dans l'ULID)
- Ajouter une contrainte de format ULID

```bash
docker exec -i coloc_postgres psql -U coloc_user -d coloc_db < migrations/000013_convert_uuid_to_ulid_remove_created_at.up.sql
```

### 3. Vérifier la structure

```bash
docker exec coloc_postgres psql -U coloc_user -d coloc_db -c "\d users"
```

Vous devriez voir :
- `id` de type `TEXT` avec contrainte CHECK pour le format ULID
- Pas de colonne `created_at`
- Clé primaire sur `id`

### 4. Tester le code Go

Le code a été adapté :
- `internal/domain/user.go` : Suppression de `CreatedAt`
- `internal/repository/postgres/user_repository.go` : Suppression des références à `created_at`
- `internal/utils/ulid.go` : Fonctions utilitaires pour générer et parser les ULIDs

```bash
# Compiler
go build ./cmd/server

# Lancer (assurez-vous que PostgreSQL local est arrêté)
make run
```

### 5. Exécuter les benchmarks

Le script va comparer les performances INSERT et SELECT entre UUID et ULID :

```bash
./benchmarks/run_benchmarks.sh
```

## 📊 Résultats Attendus

### Performance INSERT

**Hypothèse**: ULID devrait être légèrement plus rapide car :
- Génération de fonction PL/pgSQL vs extension C (uuid-ossp)
- Moins de fragmentation dans l'index B-tree (ordre séquentiel)

### Performance SELECT avec tri chronologique

**Hypothèse**: ULID devrait être **significativement plus rapide** car :
- UUID nécessite un tri sur `created_at` (index séparé)
- ULID peut trier directement sur la clé primaire `id`
- Évite une jointure d'index

### Taille sur disque

**Hypothèse**: ULID devrait être légèrement plus compact car :
- UUID stocké : 16 bytes (binaire) ou 36 chars (texte)
- ULID stocké : 26 chars (texte)
- Mais ULID a un index en moins (`created_at`)

## 🔬 Analyse des Résultats

Après avoir exécuté les benchmarks, analysez :

1. **Throughput (TPS)**: Transactions par seconde
   - Plus c'est élevé, mieux c'est

2. **Latence moyenne**: Temps de réponse moyen
   - Plus c'est bas, mieux c'est

3. **Taille sur disque**: Espace utilisé pour 10k lignes + index
   - Important pour le scaling

4. **Cas d'usage**:
   - UUIDs : Quand on a besoin de génération distribuée sans coordination
   - ULIDs : Quand on veut du tri chronologique naturel et moins de colonnes

## 📝 Questions de Réflexion

1. **Pourquoi les ULIDs améliorent-ils les performances de tri chronologique ?**
   - Les ULIDs contiennent le timestamp dans les premiers caractères
   - Tri lexicographique = tri chronologique
   - Pas besoin d'un index séparé sur `created_at`

2. **Quel est l'impact sur la fragmentation de l'index B-tree ?**
   - UUIDs : Random, fragmentation élevée
   - ULIDs : Séquentiels, fragmentation faible, meilleure localité des données

3. **Peut-on encore extraire le timestamp d'un ULID ?**
   - Oui ! Utiliser la fonction `ULIDToTime()` en Go
   - Ou décoder manuellement les 10 premiers caractères en base32

4. **Quand préférer UUID vs ULID ?**
   - UUID : Génération distribuée, compatibilité avec systèmes existants
   - ULID : Tri chronologique, moins de colonnes, meilleure performance index

## 🔄 Rollback

Si vous voulez revenir à UUID :

```bash
docker exec -i coloc_postgres psql -U coloc_user -d coloc_db < migrations/000013_convert_uuid_to_ulid_remove_created_at.down.sql
```

⚠️ **Attention**: Les timestamps exacts seront perdus (remplacés par NOW())

## 📚 Ressources

- [Spécification ULID](https://github.com/ulid/spec)
- [Package Go oklog/ulid](https://github.com/oklog/ulid)
- [Why ULID is better than UUID](https://blog.frankel.ch/ulid-vs-uuid/)
- [PostgreSQL B-tree Index Fragmentation](https://www.postgresql.org/docs/current/btree-implementation.html)

## ✅ Checklist

- [x] Branche `tp-ulid` créée
- [x] Migration ULID créée
- [x] Code Go adapté
- [x] Scripts de benchmark créés
- [ ] Migration appliquée à la base
- [ ] Benchmarks exécutés
- [ ] Résultats analysés et documentés
- [ ] Commit des changements
- [ ] (Optionnel) Push de la branche pour review

## 🎓 Conclusion

Ce TP vous a permis de :
- Comprendre les différences entre UUID et ULID
- Créer des migrations PostgreSQL complexes
- Mesurer les performances avec pgbench
- Adapter du code existant pour un nouveau format d'ID

Les ULIDs sont particulièrement utiles dans les applications modernes où le tri chronologique est fréquent et où on veut minimiser le nombre de colonnes et d'index.
