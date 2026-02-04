-- Rollback: Revenir de ULID vers UUID pour users.id et toutes les FK

-- ============================================================
-- 1. Supprimer toutes les FK qui referencent users(id)
-- ============================================================

ALTER TABLE colocations DROP CONSTRAINT IF EXISTS colocations_created_by_fkey;
ALTER TABLE colocation_members DROP CONSTRAINT IF EXISTS colocation_members_user_id_fkey;
ALTER TABLE colocation_invitations DROP CONSTRAINT IF EXISTS colocation_invitations_invited_by_fkey;
ALTER TABLE expenses DROP CONSTRAINT IF EXISTS expenses_paid_by_fkey;
ALTER TABLE expense_splits DROP CONSTRAINT IF EXISTS expense_splits_user_id_fkey;
ALTER TABLE recurring_expenses DROP CONSTRAINT IF EXISTS recurring_expenses_paid_by_fkey;
ALTER TABLE recurring_expense_splits DROP CONSTRAINT IF EXISTS recurring_expense_splits_user_id_fkey;
ALTER TABLE balances DROP CONSTRAINT IF EXISTS balances_from_user_id_fkey;
ALTER TABLE balances DROP CONSTRAINT IF EXISTS balances_to_user_id_fkey;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_from_user_id_fkey;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_to_user_id_fkey;
ALTER TABLE decisions DROP CONSTRAINT IF EXISTS decisions_created_by_fkey;
ALTER TABLE decision_votes DROP CONSTRAINT IF EXISTS decision_votes_user_id_fkey;
ALTER TABLE common_funds DROP CONSTRAINT IF EXISTS common_funds_created_by_fkey;
ALTER TABLE fund_contributions DROP CONSTRAINT IF EXISTS fund_contributions_user_id_fkey;
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_created_by_fkey;
ALTER TABLE event_participants DROP CONSTRAINT IF EXISTS event_participants_user_id_fkey;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_user_id_fkey;

-- ============================================================
-- 2. Restaurer users.id en UUID
-- ============================================================

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_id_format_check;
ALTER TABLE users DROP CONSTRAINT users_pkey;

ALTER TABLE users ADD COLUMN id_uuid UUID DEFAULT uuid_generate_v4();
ALTER TABLE users DROP COLUMN id;
ALTER TABLE users RENAME COLUMN id_uuid TO id;
ALTER TABLE users ADD PRIMARY KEY (id);

-- Restaurer created_at
ALTER TABLE users ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- ============================================================
-- 3. Convertir toutes les FK de TEXT vers UUID
-- ============================================================

-- Note: les donnees de mapping sont perdues, les FK deviennent invalides
-- Il faut vider les tables dependantes ou les reconstruire

ALTER TABLE colocations ALTER COLUMN created_by TYPE UUID USING NULL;
ALTER TABLE colocation_members ALTER COLUMN user_id TYPE UUID USING NULL;
ALTER TABLE colocation_invitations ALTER COLUMN invited_by TYPE UUID USING NULL;
ALTER TABLE expenses ALTER COLUMN paid_by TYPE UUID USING NULL;
ALTER TABLE expense_splits ALTER COLUMN user_id TYPE UUID USING NULL;
ALTER TABLE recurring_expenses ALTER COLUMN paid_by TYPE UUID USING NULL;
ALTER TABLE recurring_expense_splits ALTER COLUMN user_id TYPE UUID USING NULL;
ALTER TABLE balances ALTER COLUMN from_user_id TYPE UUID USING NULL;
ALTER TABLE balances ALTER COLUMN to_user_id TYPE UUID USING NULL;
ALTER TABLE payments ALTER COLUMN from_user_id TYPE UUID USING NULL;
ALTER TABLE payments ALTER COLUMN to_user_id TYPE UUID USING NULL;
ALTER TABLE decisions ALTER COLUMN created_by TYPE UUID USING NULL;
ALTER TABLE decision_votes ALTER COLUMN user_id TYPE UUID USING NULL;
ALTER TABLE common_funds ALTER COLUMN created_by TYPE UUID USING NULL;
ALTER TABLE fund_contributions ALTER COLUMN user_id TYPE UUID USING NULL;
ALTER TABLE events ALTER COLUMN created_by TYPE UUID USING NULL;
ALTER TABLE event_participants ALTER COLUMN user_id TYPE UUID USING NULL;
ALTER TABLE notifications ALTER COLUMN user_id TYPE UUID USING NULL;

-- ============================================================
-- 4. Recreer les FK vers users(id) UUID
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

-- Supprimer la fonction ULID
DROP FUNCTION IF EXISTS generate_ulid();

-- Restaurer extension UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

COMMENT ON COLUMN users.id IS NULL;
COMMENT ON TABLE users IS 'Table des utilisateurs';
