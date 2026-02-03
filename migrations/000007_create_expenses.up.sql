-- Create expenses table
CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    colocation_id UUID NOT NULL REFERENCES colocations(id) ON DELETE CASCADE,
    paid_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES expense_categories(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    split_type VARCHAR(20) NOT NULL CHECK (split_type IN ('equal', 'percentage', 'custom')),
    expense_date DATE NOT NULL DEFAULT CURRENT_DATE,
    recurring_id UUID,  -- Reference to recurring_expenses if generated from template
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create expense_splits table
CREATE TABLE expense_splits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount >= 0),
    percentage DECIMAL(5, 2) CHECK (percentage >= 0 AND percentage <= 100),
    is_settled BOOLEAN NOT NULL DEFAULT false,
    UNIQUE(expense_id, user_id)
);

-- Create recurring_expenses table
CREATE TABLE recurring_expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    colocation_id UUID NOT NULL REFERENCES colocations(id) ON DELETE CASCADE,
    paid_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES expense_categories(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    split_type VARCHAR(20) NOT NULL CHECK (split_type IN ('equal', 'percentage', 'custom')),
    recurrence VARCHAR(20) NOT NULL CHECK (recurrence IN ('daily', 'weekly', 'monthly', 'yearly')),
    next_due_date DATE NOT NULL,
    end_date DATE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create recurring_expense_splits table (percentage allocation template)
CREATE TABLE recurring_expense_splits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recurring_id UUID NOT NULL REFERENCES recurring_expenses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    percentage DECIMAL(5, 2) NOT NULL CHECK (percentage >= 0 AND percentage <= 100),
    UNIQUE(recurring_id, user_id)
);

-- Add foreign key for recurring_id in expenses
ALTER TABLE expenses
ADD CONSTRAINT fk_expenses_recurring
FOREIGN KEY (recurring_id) REFERENCES recurring_expenses(id) ON DELETE SET NULL;

-- Indexes
CREATE INDEX idx_expenses_colocation ON expenses(colocation_id);
CREATE INDEX idx_expenses_paid_by ON expenses(paid_by);
CREATE INDEX idx_expenses_category ON expenses(category_id);
CREATE INDEX idx_expenses_date ON expenses(expense_date DESC);
CREATE INDEX idx_expenses_colocation_date ON expenses(colocation_id, expense_date DESC);
CREATE INDEX idx_expense_splits_expense ON expense_splits(expense_id);
CREATE INDEX idx_expense_splits_user ON expense_splits(user_id);
CREATE INDEX idx_expense_splits_unsettled ON expense_splits(user_id) WHERE is_settled = false;
CREATE INDEX idx_recurring_expenses_colocation ON recurring_expenses(colocation_id);
CREATE INDEX idx_recurring_expenses_active ON recurring_expenses(is_active, next_due_date) WHERE is_active = true;
