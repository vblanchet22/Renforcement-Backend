-- Migration: Convertir UUID en ULID pour users.id et toutes les FK associees
-- ULID = Universally Unique Lexicographically Sortable Identifier (26 caracteres base32)
-- Le timestamp est encode dans les 10 premiers caracteres => plus besoin de created_at

-- Fonction pour generer un ULID en PostgreSQL
CREATE OR REPLACE FUNCTION generate_ulid() RETURNS TEXT AS $$
DECLARE
    timestamp_ms BIGINT;
    encoding TEXT := '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
    ulid TEXT := '';
    i INT;
BEGIN
    timestamp_ms := (EXTRACT(EPOCH FROM NOW()) * 1000)::BIGINT;

    -- Encoder le timestamp (10 caracteres)
    FOR i IN 1..10 LOOP
        ulid := SUBSTR(encoding, (timestamp_ms % 32) + 1, 1) || ulid;
        timestamp_ms := timestamp_ms / 32;
    END LOOP;

    -- Generer 16 caracteres aleatoires (80 bits)
    FOR i IN 1..16 LOOP
        ulid := ulid || SUBSTR(encoding, (FLOOR(RANDOM() * 32) + 1)::INT, 1);
    END LOOP;

    RETURN ulid;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- 1. Supprimer toutes les FK qui referencent users(id)
-- ============================================================

-- colocations
ALTER TABLE colocations DROP CONSTRAINT IF EXISTS colocations_created_by_fkey;

-- colocation_members
ALTER TABLE colocation_members DROP CONSTRAINT IF EXISTS colocation_members_user_id_fkey;

-- colocation_invitations
ALTER TABLE colocation_invitations DROP CONSTRAINT IF EXISTS colocation_invitations_invited_by_fkey;

-- expenses
ALTER TABLE expenses DROP CONSTRAINT IF EXISTS expenses_paid_by_fkey;

-- expense_splits
ALTER TABLE expense_splits DROP CONSTRAINT IF EXISTS expense_splits_user_id_fkey;

-- recurring_expenses
ALTER TABLE recurring_expenses DROP CONSTRAINT IF EXISTS recurring_expenses_paid_by_fkey;

-- recurring_expense_splits
ALTER TABLE recurring_expense_splits DROP CONSTRAINT IF EXISTS recurring_expense_splits_user_id_fkey;

-- balances
ALTER TABLE balances DROP CONSTRAINT IF EXISTS balances_from_user_id_fkey;
ALTER TABLE balances DROP CONSTRAINT IF EXISTS balances_to_user_id_fkey;

-- payments
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_from_user_id_fkey;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_to_user_id_fkey;

-- decisions
ALTER TABLE decisions DROP CONSTRAINT IF EXISTS decisions_created_by_fkey;

-- decision_votes
ALTER TABLE decision_votes DROP CONSTRAINT IF EXISTS decision_votes_user_id_fkey;

-- common_funds
ALTER TABLE common_funds DROP CONSTRAINT IF EXISTS common_funds_created_by_fkey;

-- fund_contributions
ALTER TABLE fund_contributions DROP CONSTRAINT IF EXISTS fund_contributions_user_id_fkey;

-- events
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_created_by_fkey;

-- event_participants
ALTER TABLE event_participants DROP CONSTRAINT IF EXISTS event_participants_user_id_fkey;

-- notifications
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_user_id_fkey;

-- ============================================================
-- 2. Convertir users.id de UUID vers TEXT (ULID)
-- ============================================================

-- Ajouter colonne temporaire
ALTER TABLE users ADD COLUMN id_ulid TEXT;

-- Generer des ULIDs pour les lignes existantes
UPDATE users SET id_ulid = generate_ulid();

-- Contraintes sur la nouvelle colonne
ALTER TABLE users ALTER COLUMN id_ulid SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_id_ulid_unique UNIQUE (id_ulid);

-- Creer un mapping temporaire (ancien UUID -> nouveau ULID)
CREATE TEMP TABLE user_id_mapping AS
SELECT id::TEXT AS old_id, id_ulid AS new_id FROM users;

-- ============================================================
-- 3. Convertir toutes les colonnes FK de UUID vers TEXT
-- ============================================================

-- colocations.created_by
ALTER TABLE colocations ALTER COLUMN created_by TYPE TEXT USING created_by::TEXT;
UPDATE colocations SET created_by = m.new_id FROM user_id_mapping m WHERE colocations.created_by = m.old_id;

-- colocation_members.user_id
ALTER TABLE colocation_members ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;
UPDATE colocation_members SET user_id = m.new_id FROM user_id_mapping m WHERE colocation_members.user_id = m.old_id;

-- colocation_invitations.invited_by
ALTER TABLE colocation_invitations ALTER COLUMN invited_by TYPE TEXT USING invited_by::TEXT;
UPDATE colocation_invitations SET invited_by = m.new_id FROM user_id_mapping m WHERE colocation_invitations.invited_by = m.old_id;

-- expenses.paid_by
ALTER TABLE expenses ALTER COLUMN paid_by TYPE TEXT USING paid_by::TEXT;
UPDATE expenses SET paid_by = m.new_id FROM user_id_mapping m WHERE expenses.paid_by = m.old_id;

-- expense_splits.user_id
ALTER TABLE expense_splits ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;
UPDATE expense_splits SET user_id = m.new_id FROM user_id_mapping m WHERE expense_splits.user_id = m.old_id;

-- recurring_expenses.paid_by
ALTER TABLE recurring_expenses ALTER COLUMN paid_by TYPE TEXT USING paid_by::TEXT;
UPDATE recurring_expenses SET paid_by = m.new_id FROM user_id_mapping m WHERE recurring_expenses.paid_by = m.old_id;

-- recurring_expense_splits.user_id
ALTER TABLE recurring_expense_splits ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;
UPDATE recurring_expense_splits SET user_id = m.new_id FROM user_id_mapping m WHERE recurring_expense_splits.user_id = m.old_id;

-- balances.from_user_id & to_user_id
ALTER TABLE balances ALTER COLUMN from_user_id TYPE TEXT USING from_user_id::TEXT;
ALTER TABLE balances ALTER COLUMN to_user_id TYPE TEXT USING to_user_id::TEXT;
UPDATE balances SET from_user_id = m.new_id FROM user_id_mapping m WHERE balances.from_user_id = m.old_id;
UPDATE balances SET to_user_id = m.new_id FROM user_id_mapping m WHERE balances.to_user_id = m.old_id;

-- payments.from_user_id & to_user_id
ALTER TABLE payments ALTER COLUMN from_user_id TYPE TEXT USING from_user_id::TEXT;
ALTER TABLE payments ALTER COLUMN to_user_id TYPE TEXT USING to_user_id::TEXT;
UPDATE payments SET from_user_id = m.new_id FROM user_id_mapping m WHERE payments.from_user_id = m.old_id;
UPDATE payments SET to_user_id = m.new_id FROM user_id_mapping m WHERE payments.to_user_id = m.old_id;

-- decisions.created_by
ALTER TABLE decisions ALTER COLUMN created_by TYPE TEXT USING created_by::TEXT;
UPDATE decisions SET created_by = m.new_id FROM user_id_mapping m WHERE decisions.created_by = m.old_id;

-- decision_votes.user_id
ALTER TABLE decision_votes ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;
UPDATE decision_votes SET user_id = m.new_id FROM user_id_mapping m WHERE decision_votes.user_id = m.old_id;

-- common_funds.created_by
ALTER TABLE common_funds ALTER COLUMN created_by TYPE TEXT USING created_by::TEXT;
UPDATE common_funds SET created_by = m.new_id FROM user_id_mapping m WHERE common_funds.created_by = m.old_id;

-- fund_contributions.user_id
ALTER TABLE fund_contributions ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;
UPDATE fund_contributions SET user_id = m.new_id FROM user_id_mapping m WHERE fund_contributions.user_id = m.old_id;

-- events.created_by
ALTER TABLE events ALTER COLUMN created_by TYPE TEXT USING created_by::TEXT;
UPDATE events SET created_by = m.new_id FROM user_id_mapping m WHERE events.created_by = m.old_id;

-- event_participants.user_id
ALTER TABLE event_participants ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;
UPDATE event_participants SET user_id = m.new_id FROM user_id_mapping m WHERE event_participants.user_id = m.old_id;

-- notifications.user_id
ALTER TABLE notifications ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;
UPDATE notifications SET user_id = m.new_id FROM user_id_mapping m WHERE notifications.user_id = m.old_id;

-- ============================================================
-- 4. Remplacer users.id par la colonne ULID
-- ============================================================

ALTER TABLE users DROP CONSTRAINT users_pkey;
ALTER TABLE users DROP COLUMN id;
ALTER TABLE users RENAME COLUMN id_ulid TO id;
ALTER TABLE users ADD PRIMARY KEY (id);
ALTER TABLE users ALTER COLUMN id SET DEFAULT generate_ulid();

-- Contrainte de format ULID (26 caracteres Crockford base32)
ALTER TABLE users ADD CONSTRAINT users_id_format_check
    CHECK (id ~ '^[0-9A-HJKMNP-TV-Z]{26}$');

-- ============================================================
-- 5. Supprimer created_at (le timestamp est dans le ULID)
-- ============================================================

DROP INDEX IF EXISTS idx_users_created_at;
ALTER TABLE users DROP COLUMN IF EXISTS created_at;

-- ============================================================
-- 6. Recreer toutes les FK vers users(id)
-- ============================================================

ALTER TABLE colocations ADD CONSTRAINT colocations_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE colocation_members ADD CONSTRAINT colocation_members_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE colocation_invitations ADD CONSTRAINT colocation_invitations_invited_by_fkey
    FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE expenses ADD CONSTRAINT expenses_paid_by_fkey
    FOREIGN KEY (paid_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE expense_splits ADD CONSTRAINT expense_splits_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE recurring_expenses ADD CONSTRAINT recurring_expenses_paid_by_fkey
    FOREIGN KEY (paid_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE recurring_expense_splits ADD CONSTRAINT recurring_expense_splits_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE balances ADD CONSTRAINT balances_from_user_id_fkey
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE balances ADD CONSTRAINT balances_to_user_id_fkey
    FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE payments ADD CONSTRAINT payments_from_user_id_fkey
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE payments ADD CONSTRAINT payments_to_user_id_fkey
    FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE decisions ADD CONSTRAINT decisions_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE decision_votes ADD CONSTRAINT decision_votes_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE common_funds ADD CONSTRAINT common_funds_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE fund_contributions ADD CONSTRAINT fund_contributions_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE events ADD CONSTRAINT events_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE event_participants ADD CONSTRAINT event_participants_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE notifications ADD CONSTRAINT notifications_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ============================================================
-- Commentaires
-- ============================================================
COMMENT ON COLUMN users.id IS 'ULID: Identifiant unique triable chronologiquement (26 chars base32, contient timestamp)';
COMMENT ON TABLE users IS 'Table des utilisateurs - IDs en format ULID pour tri chronologique natif';
