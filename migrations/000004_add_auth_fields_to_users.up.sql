-- Add authentication and profile fields to users table
ALTER TABLE users
ADD COLUMN password_hash VARCHAR(255),
ADD COLUMN avatar_url TEXT,
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- Create index for email lookups during login
CREATE INDEX idx_users_email_active ON users(email) WHERE is_active = true;
