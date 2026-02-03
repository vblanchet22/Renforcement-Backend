-- Create expense_categories table
CREATE TABLE expense_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(50),
    color VARCHAR(7),  -- Hex color code
    colocation_id UUID REFERENCES colocations(id) ON DELETE CASCADE,  -- NULL = global category
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique constraint: name must be unique within a colocation (or globally if colocation_id is NULL)
CREATE UNIQUE INDEX idx_categories_name_colocation ON expense_categories(name, colocation_id)
    WHERE colocation_id IS NOT NULL;
CREATE UNIQUE INDEX idx_categories_name_global ON expense_categories(name)
    WHERE colocation_id IS NULL;

-- Index for listing categories by colocation
CREATE INDEX idx_categories_colocation ON expense_categories(colocation_id);
