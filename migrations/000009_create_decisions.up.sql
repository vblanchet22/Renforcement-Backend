-- Create decisions table
CREATE TABLE decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    colocation_id UUID NOT NULL REFERENCES colocations(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    options JSONB NOT NULL,  -- Array of option strings
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed')),
    deadline TIMESTAMPTZ,
    allow_multiple BOOLEAN NOT NULL DEFAULT false,
    is_anonymous BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create decision_votes table
CREATE TABLE decision_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id UUID NOT NULL REFERENCES decisions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    option_index INTEGER NOT NULL CHECK (option_index >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(decision_id, user_id, option_index)  -- User can vote once per option
);

-- Indexes
CREATE INDEX idx_decisions_colocation ON decisions(colocation_id);
CREATE INDEX idx_decisions_status ON decisions(status);
CREATE INDEX idx_decisions_deadline ON decisions(deadline) WHERE status = 'open' AND deadline IS NOT NULL;
CREATE INDEX idx_decision_votes_decision ON decision_votes(decision_id);
CREATE INDEX idx_decision_votes_user ON decision_votes(user_id);
