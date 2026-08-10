ALTER TABLE sandbox_exports
    ADD COLUMN IF NOT EXISTS payload_bytes BYTEA;
