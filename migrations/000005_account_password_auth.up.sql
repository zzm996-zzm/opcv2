ALTER TABLE users
    ALTER COLUMN phone DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS account VARCHAR(32),
    ADD COLUMN IF NOT EXISTS password_hash TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS users_account_unique
    ON users (account)
    WHERE account IS NOT NULL AND account <> '';
