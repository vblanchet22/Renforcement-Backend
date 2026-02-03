-- Convertir les colonnes TIMESTAMP en TIMESTAMPTZ pour une gestion correcte des fuseaux horaires
-- TIMESTAMPTZ stocke toujours en UTC et convertit automatiquement selon le timezone de la session

-- Modifier created_at
ALTER TABLE users ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';

-- Modifier updated_at
ALTER TABLE users ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- Mettre à jour les valeurs par défaut pour utiliser TIMESTAMPTZ
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT NOW();
ALTER TABLE users ALTER COLUMN updated_at SET DEFAULT NOW();
