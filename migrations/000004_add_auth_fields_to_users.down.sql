-- Remove authentication and profile fields from users table
DROP INDEX IF EXISTS idx_users_email_active;

ALTER TABLE users
DROP COLUMN IF EXISTS password_hash,
DROP COLUMN IF EXISTS avatar_url,
DROP COLUMN IF EXISTS is_active;
