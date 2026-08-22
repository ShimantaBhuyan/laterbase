-- +goose Up
CREATE TABLE groups (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL UNIQUE,
    color      TEXT NOT NULL DEFAULT '#5e6ad2',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed default groups
INSERT INTO groups (name, color, sort_order) VALUES
    ('articles',       '#5e6ad2', 1),
    ('github',         '#10b981', 2),
    ('docs',           '#f59e0b', 3),
    ('videos',         '#ef4444', 4),
    ('tools',          '#8b5cf6', 5),
    ('uncategorized',  '#6b7280', 6);

-- Add FK from links to groups once groups table exists
ALTER TABLE links ADD CONSTRAINT fk_links_group
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE links DROP CONSTRAINT IF EXISTS fk_links_group;
DROP TABLE IF EXISTS groups;
