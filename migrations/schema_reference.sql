-- Table schema after migration
-- This file shows the users table structure and includes test data examples

-- ============================================================================
-- Table Structure (created by 001_create_users_table.sql)
-- ============================================================================

/*
CREATE TABLE users (
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
*/

-- ============================================================================
-- Verify Table Structure
-- ============================================================================

-- View the table structure
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns 
WHERE table_name = 'users'
ORDER BY ordinal_position;

-- View table statistics
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE tablename = 'users';

-- ============================================================================
-- Test Data (Optional - for development only)
-- ============================================================================

-- Example 1: User without 2FA (basic login)
INSERT INTO users (username, email, name, password_hash, two_fa_enabled)
VALUES (
    'testuser',
    'testuser@example.com',
    'Test User',
    '$2a$10$N9qo8uLOickgx2ZMRZoMye', -- bcrypt hash (example)
    false
) ON CONFLICT (username) DO NOTHING;

-- Example 2: User with 2FA enabled
INSERT INTO users (username, email, name, password_hash, two_fa_enabled, two_fa_secret, backup_codes)
VALUES (
    'admin',
    'admin@example.com',
    'Admin User',
    '$2a$10$N9qo8uLOickgx2ZMRZoMye', -- bcrypt hash (example)
    true,
    'JBSWY3DPEBLW64TMMQ======', -- TOTP secret (base32 encoded)
    'ABCD1234,EFGH5678,IJKL9012,MNOP3456,QRST7890,UVWX1234,YZAB5678,CDEF9012' -- backup codes
) ON CONFLICT (username) DO NOTHING;

-- ============================================================================
-- Useful Queries
-- ============================================================================

-- Get all users
SELECT id, username, email, name, two_fa_enabled, created_at FROM users;

-- Get a specific user by username
SELECT * FROM users WHERE username = 'testuser';

-- Get all users with 2FA enabled
SELECT username, email, two_fa_enabled FROM users WHERE two_fa_enabled = true;

-- Count total users
SELECT COUNT(*) as total_users FROM users;

-- Find users created in the last 24 hours
SELECT username, email, created_at FROM users 
WHERE created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;

-- Find users who haven't updated their password in 90 days
SELECT username, email, updated_at FROM users
WHERE updated_at < NOW() - INTERVAL '90 days'
ORDER BY updated_at;

-- Check for duplicate emails or usernames
SELECT username, COUNT(*) FROM users GROUP BY username HAVING COUNT(*) > 1;
SELECT email, COUNT(*) FROM users GROUP BY email HAVING COUNT(*) > 1;

-- ============================================================================
-- Data Maintenance
-- ============================================================================

-- Update a user's password
UPDATE users 
SET password_hash = 'new_bcrypt_hash_here' 
WHERE username = 'testuser';

-- Enable 2FA for a user
UPDATE users 
SET two_fa_enabled = true, 
    two_fa_secret = 'JBSWY3DPEBLW64TMMQ======',
    backup_codes = 'CODE1,CODE2,CODE3,CODE4,CODE5,CODE6,CODE7,CODE8'
WHERE username = 'testuser';

-- Disable 2FA for a user
UPDATE users 
SET two_fa_enabled = false, 
    two_fa_secret = NULL,
    backup_codes = NULL
WHERE username = 'testuser';

-- Delete a user
DELETE FROM users WHERE username = 'testuser';

-- ============================================================================
-- Important Notes for Password Hashing
-- ============================================================================

/*
IMPORTANT: Use bcrypt for password hashing in your Go application!

Example Go code:
    import "golang.org/x/crypto/bcrypt"
    
    // Hash password during registration
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    
    // Verify password during login
    err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))

Never store plain text passwords!
*/

-- ============================================================================
-- TOTP Secret Format
-- ============================================================================

/*
The two_fa_secret should be a base32-encoded string that can be used with:
- Google Authenticator
- Authy
- Microsoft Authenticator
- Any RFC 6238 TOTP-compatible app

Example TOTP secrets:
- JBSWY3DPEBLW64TMMQ======
- GFZXG43JGVZWC4LVGY======
- HGKFKQQC4L4GFXLQ======

The pquerna/otp library in Go handles generation and verification.
*/

-- ============================================================================
-- Backup Codes Format
-- ============================================================================

/*
Backup codes should be stored as comma-separated values.
Each code is typically:
- 8 characters long
- Alphanumeric (A-Z, 0-9)
- Uppercase preferred for readability

Example backup codes:
ABCD1234,EFGH5678,IJKL9012,MNOP3456,QRST7890,UVWX1234,YZAB5678,CDEF9012

A user typically gets 8 backup codes during 2FA setup.
Each code can only be used once for recovery.
*/
