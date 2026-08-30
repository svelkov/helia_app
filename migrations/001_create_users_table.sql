-- Create users table for authentication and 2FA support
-- This migration creates the base users table with 2FA fields

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    two_fa_enabled BOOLEAN DEFAULT false,
    two_fa_secret VARCHAR(255),
    backup_codes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Add comments for documentation
COMMENT ON TABLE users IS 'Stores user account information with 2FA support';
COMMENT ON COLUMN users.id IS 'Unique user identifier';
COMMENT ON COLUMN users.username IS 'Unique username for login';
COMMENT ON COLUMN users.email IS 'Unique email address';
COMMENT ON COLUMN users.name IS 'Full name of the user';
COMMENT ON COLUMN users.password_hash IS 'Hashed password (should use bcrypt)';
COMMENT ON COLUMN users.two_fa_enabled IS 'Whether 2FA is enabled for this user';
COMMENT ON COLUMN users.two_fa_secret IS 'TOTP secret key for authenticator apps';
COMMENT ON COLUMN users.backup_codes IS 'Comma-separated backup codes for 2FA recovery';
COMMENT ON COLUMN users.created_at IS 'Account creation timestamp';
COMMENT ON COLUMN users.updated_at IS 'Last account update timestamp';
