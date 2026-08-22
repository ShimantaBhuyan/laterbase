-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE links (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url           TEXT NOT NULL UNIQUE,
    title         TEXT,
    summary       TEXT,
    content_text  TEXT,
    source_type   TEXT,
    group_id      UUID,
    status        TEXT NOT NULL DEFAULT 'new'
                  CHECK (status IN ('new','fetching','processing','ready','inaccessible','paywalled','failed')),
    error_message TEXT,
    favicon_url   TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at  TIMESTAMPTZ
);

CREATE INDEX idx_links_status ON links(status);
CREATE INDEX idx_links_group_id ON links(group_id);
CREATE INDEX idx_links_source_type ON links(source_type);
CREATE INDEX idx_links_created_at ON links(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS links;
