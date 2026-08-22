-- +goose Up
-- Ensure the interesting group exists (will be created if not already present).
INSERT INTO groups (name, color, sort_order) VALUES ('interesting', '#f59e0b', 99)
ON CONFLICT (name) DO NOTHING;

-- Migrate links that had the old is_interesting flag into the interesting group.
INSERT INTO link_groups (link_id, group_id)
SELECT l.id, g.id
FROM links l
CROSS JOIN groups g
WHERE l.is_interesting = true
  AND g.name = 'interesting'
ON CONFLICT DO NOTHING;

-- Remove the old boolean column.
ALTER TABLE links DROP COLUMN IF EXISTS is_interesting;

-- +goose Down
ALTER TABLE links ADD COLUMN is_interesting BOOLEAN NOT NULL DEFAULT false;

UPDATE links l
SET is_interesting = true
FROM link_groups lg
JOIN groups g ON g.id = lg.group_id
WHERE lg.link_id = l.id
  AND g.name = 'interesting';
