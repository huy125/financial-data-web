-- Remove trigger first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Remove function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Remove index
DROP INDEX IF EXISTS idx_users_email;

-- Remove constraints
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_email_unique,
    DROP CONSTRAINT IF EXISTS users_email_check,
    DROP CONSTRAINT IF EXISTS users_email_format;

-- Remove columns
ALTER TABLE users
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS firstname,
    DROP COLUMN IF EXISTS lastname,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

-- Add username and hash columns back
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS username VARCHAR(255),
    ADD COLUMN IF NOT EXISTS hash VARCHAR(255);
