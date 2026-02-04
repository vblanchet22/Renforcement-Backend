-- Script de benchmark: Test avec ULID
-- Créer une table de test avec ULID

DROP TABLE IF EXISTS users_ulid_test CASCADE;

-- Fonction pour générer un ULID
CREATE OR REPLACE FUNCTION generate_ulid() RETURNS TEXT AS $$
DECLARE
    timestamp_ms BIGINT;
    randomness TEXT;
    encoding TEXT := '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
    ulid TEXT := '';
    i INT;
BEGIN
    timestamp_ms := (EXTRACT(EPOCH FROM NOW()) * 1000)::BIGINT;

    FOR i IN 1..10 LOOP
        ulid := SUBSTR(encoding, (timestamp_ms % 32) + 1, 1) || ulid;
        timestamp_ms := timestamp_ms / 32;
    END LOOP;

    FOR i IN 1..16 LOOP
        randomness := SUBSTR(encoding, (FLOOR(RANDOM() * 32) + 1)::INT, 1);
        ulid := ulid || randomness;
    END LOOP;

    RETURN ulid;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users_ulid_test (
    id TEXT PRIMARY KEY DEFAULT generate_ulid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    nom VARCHAR(100) NOT NULL,
    prenom VARCHAR(100) NOT NULL,
    telephone VARCHAR(255),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_ulid_id_format_check CHECK (id ~ '^[0-9A-HJKMNP-TV-Z]{26}$')
);

-- Créer les mêmes index
CREATE INDEX idx_users_ulid_email ON users_ulid_test(email);
CREATE INDEX idx_users_ulid_id ON users_ulid_test(id DESC);

-- Insérer des données de test (10k lignes)
INSERT INTO users_ulid_test (email, nom, prenom, telephone)
SELECT
    'user' || i || '@test.com',
    'Nom' || i,
    'Prenom' || i,
    '060000' || LPAD(i::TEXT, 4, '0')
FROM generate_series(1, 10000) AS i;

-- Analyser la table
ANALYZE users_ulid_test;
