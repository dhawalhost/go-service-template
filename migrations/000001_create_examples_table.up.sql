CREATE TABLE IF NOT EXISTS examples (
    id          VARCHAR(36)  PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    tenant_id   VARCHAR(36)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_examples_tenant_id ON examples(tenant_id);
CREATE INDEX IF NOT EXISTS idx_examples_tenant_name ON examples(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_examples_deleted_at ON examples(deleted_at);
