-- Remove default global expense categories
DELETE FROM expense_categories WHERE colocation_id IS NULL;
