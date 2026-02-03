-- Create balances table (cached balance between users)
CREATE TABLE balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    colocation_id UUID NOT NULL REFERENCES colocations(id) ON DELETE CASCADE,
    from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL DEFAULT 0,  -- Positive = from_user owes to_user
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(colocation_id, from_user_id, to_user_id),
    CHECK (from_user_id != to_user_id)
);

-- Create payments table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    colocation_id UUID NOT NULL REFERENCES colocations(id) ON DELETE CASCADE,
    from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'rejected')),
    note TEXT,
    confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (from_user_id != to_user_id)
);

-- Indexes
CREATE INDEX idx_balances_colocation ON balances(colocation_id);
CREATE INDEX idx_balances_from_user ON balances(from_user_id);
CREATE INDEX idx_balances_to_user ON balances(to_user_id);
CREATE INDEX idx_payments_colocation ON payments(colocation_id);
CREATE INDEX idx_payments_from_user ON payments(from_user_id);
CREATE INDEX idx_payments_to_user ON payments(to_user_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_pending ON payments(to_user_id, status) WHERE status = 'pending';
