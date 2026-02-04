-- Rollback: Revenir de ULID vers UUID et restaurer created_at

-- 1. Ajouter une colonne temporaire pour les UUIDs
ALTER TABLE users ADD COLUMN id_uuid UUID DEFAULT uuid_generate_v4();

-- 2. Supprimer la clé primaire actuelle
ALTER TABLE users DROP CONSTRAINT users_pkey;

-- 3. Supprimer les contraintes liées à l'ULID
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_id_format_check;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_id_ulid_unique;

-- 4. Supprimer l'ancienne colonne id (ULID)
ALTER TABLE users DROP COLUMN id;

-- 5. Renommer id_uuid en id
ALTER TABLE users RENAME COLUMN id_uuid TO id;

-- 6. Définir id comme nouvelle clé primaire
ALTER TABLE users ADD PRIMARY KEY (id);

-- 7. Restaurer la colonne created_at avec TIMESTAMPTZ
ALTER TABLE users ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- 8. Tenter d'extraire le timestamp depuis l'ULID (approximation)
-- Note: Les données exactes sont perdues, on utilise NOW() par défaut
UPDATE users SET created_at = NOW();

-- 9. Réactiver l'extension UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 10. Supprimer la fonction generate_ulid
DROP FUNCTION IF EXISTS generate_ulid();

-- Restaurer les commentaires
COMMENT ON COLUMN users.id IS NULL;
COMMENT ON TABLE users IS 'Table des utilisateurs';
