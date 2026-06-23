DROP INDEX IF EXISTS users_account_unique;

ALTER TABLE users
    DROP COLUMN IF EXISTS password_hash,
    DROP COLUMN IF EXISTS account,
    ALTER COLUMN phone SET NOT NULL;
