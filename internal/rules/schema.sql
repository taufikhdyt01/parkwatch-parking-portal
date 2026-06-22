-- rules service schema. Bootstrapped on startup (idempotent).
CREATE TABLE IF NOT EXISTS rule_versions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version    INTEGER NOT NULL UNIQUE,
    is_active  BOOLEAN NOT NULL DEFAULT false,
    config     JSONB NOT NULL,
    created_by TEXT NOT NULL DEFAULT 'system',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Enforce at most one active version at any time.
CREATE UNIQUE INDEX IF NOT EXISTS rule_versions_one_active
    ON rule_versions (is_active) WHERE is_active;
