-- Create common_funds table
CREATE TABLE common_funds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    colocation_id UUID NOT NULL REFERENCES colocations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    target_amount DECIMAL(10, 2),
    current_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create fund_contributions table
CREATE TABLE fund_contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fund_id UUID NOT NULL REFERENCES common_funds(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create events table
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    colocation_id UUID NOT NULL REFERENCES colocations(id) ON DELETE CASCADE,
    fund_id UUID REFERENCES common_funds(id) ON DELETE SET NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    budget DECIMAL(10, 2),
    event_date TIMESTAMPTZ NOT NULL,
    location TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'upcoming' CHECK (status IN ('upcoming', 'ongoing', 'completed', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create event_participants table
CREATE TABLE event_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rsvp VARCHAR(20) NOT NULL DEFAULT 'going' CHECK (rsvp IN ('going', 'maybe', 'not_going')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, user_id)
);

-- Indexes
CREATE INDEX idx_common_funds_colocation ON common_funds(colocation_id);
CREATE INDEX idx_common_funds_active ON common_funds(colocation_id, is_active) WHERE is_active = true;
CREATE INDEX idx_fund_contributions_fund ON fund_contributions(fund_id);
CREATE INDEX idx_fund_contributions_user ON fund_contributions(user_id);
CREATE INDEX idx_events_colocation ON events(colocation_id);
CREATE INDEX idx_events_fund ON events(fund_id);
CREATE INDEX idx_events_date ON events(event_date);
CREATE INDEX idx_events_upcoming ON events(colocation_id, event_date) WHERE status = 'upcoming';
CREATE INDEX idx_event_participants_event ON event_participants(event_id);
CREATE INDEX idx_event_participants_user ON event_participants(user_id);
