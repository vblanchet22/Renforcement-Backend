-- Script de benchmark: Test avec UUID
-- Créer une table de test avec UUID

DROP TABLE IF EXISTS users_uuid_test CASCADE;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users_uuid_test (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    nom VARCHAR(100) NOT NULL,
    prenom VARCHAR(100) NOT NULL,
    telephone VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Créer les mêmes index que la table réelle
CREATE INDEX idx_users_uuid_email ON users_uuid_test(email);
CREATE INDEX idx_users_uuid_created_at ON users_uuid_test(created_at DESC);

-- Insérer des données de test (10k lignes)
INSERT INTO users_uuid_test (email, nom, prenom, telephone)
SELECT
    'user' || i || '@test.com',
    'Nom' || i,
    'Prenom' || i,
    '060000' || LPAD(i::TEXT, 4, '0')
FROM generate_series(1, 10000) AS i;

-- Analyser la table pour mettre à jour les statistiques
ANALYZE users_uuid_test;
