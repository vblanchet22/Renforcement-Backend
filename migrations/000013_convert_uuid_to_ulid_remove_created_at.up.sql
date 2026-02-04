-- Migration: Remplacer UUID par ULID et supprimer created_at
-- ULID = Universally Unique Lexicographically Sortable Identifier (26 caractères base32)
-- Avantages: tri chronologique naturel, contient le timestamp, plus compact visuellement

-- Fonction pour générer un ULID en PostgreSQL (sans extension externe)
CREATE OR REPLACE FUNCTION generate_ulid() RETURNS TEXT AS $$
DECLARE
    -- Timestamp: 48 bits (millisecondes depuis epoch Unix)
    timestamp_ms BIGINT;
    -- Randomness: 80 bits
    randomness TEXT;
    -- Caractères Crockford Base32 (pas de I, L, O, U pour éviter confusion)
    encoding TEXT := '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
    ulid TEXT := '';
    i INT;
BEGIN
    -- Obtenir le timestamp en millisecondes
    timestamp_ms := (EXTRACT(EPOCH FROM NOW()) * 1000)::BIGINT;

    -- Encoder le timestamp (10 caractères)
    FOR i IN 1..10 LOOP
        ulid := SUBSTR(encoding, (timestamp_ms % 32) + 1, 1) || ulid;
        timestamp_ms := timestamp_ms / 32;
    END LOOP;

    -- Générer 16 caractères aléatoires (80 bits)
    FOR i IN 1..16 LOOP
        randomness := SUBSTR(encoding, (FLOOR(RANDOM() * 32) + 1)::INT, 1);
        ulid := ulid || randomness;
    END LOOP;

    RETURN ulid;
END;
$$ LANGUAGE plpgsql;

-- 1. Ajouter une colonne temporaire pour les ULIDs
ALTER TABLE users ADD COLUMN id_ulid TEXT;

-- 2. Générer des ULIDs pour toutes les lignes existantes
-- Note: On génère avec un léger délai pour garantir l'unicité temporelle
UPDATE users SET id_ulid = generate_ulid() || LPAD(FLOOR(RANDOM() * 1000)::TEXT, 3, '0');

-- 3. Définir id_ulid comme NOT NULL et UNIQUE
ALTER TABLE users ALTER COLUMN id_ulid SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_id_ulid_unique UNIQUE (id_ulid);

-- 4. Supprimer l'ancienne clé primaire et la colonne UUID
ALTER TABLE users DROP CONSTRAINT users_pkey;
ALTER TABLE users DROP COLUMN id;

-- 5. Renommer id_ulid en id
ALTER TABLE users RENAME COLUMN id_ulid TO id;

-- 6. Définir id comme nouvelle clé primaire
ALTER TABLE users ADD PRIMARY KEY (id);

-- 7. Ajouter une contrainte pour valider le format ULID (26 caractères alphanumériques)
ALTER TABLE users ADD CONSTRAINT users_id_format_check
    CHECK (id ~ '^[0-9A-HJKMNP-TV-Z]{26}$');

-- 8. Supprimer la colonne created_at (l'info est dans le ULID)
ALTER TABLE users DROP COLUMN created_at;

-- 9. Mettre à jour les valeurs par défaut
ALTER TABLE users ALTER COLUMN id SET DEFAULT generate_ulid();

-- Afficher la nouvelle structure
COMMENT ON COLUMN users.id IS 'ULID: Identifiant unique triable chronologiquement (26 chars base32, contient timestamp)';
COMMENT ON TABLE users IS 'Table des utilisateurs - IDs en format ULID pour tri chronologique natif';
