-- +goose Up
CREATE TABLE link_groups (
    link_id   UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    group_id  UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    PRIMARY KEY (link_id, group_id)
);

CREATE INDEX idx_link_groups_link_id ON link_groups(link_id);
CREATE INDEX idx_link_groups_group_id ON link_groups(group_id);

-- Migrate existing single-group assignments
INSERT INTO link_groups (link_id, group_id)
SELECT id, group_id FROM links WHERE group_id IS NOT NULL;

-- Drop the old column and FK
ALTER TABLE links DROP CONSTRAINT IF EXISTS fk_links_group;
ALTER TABLE links DROP COLUMN IF EXISTS group_id;

-- +goose Down
ALTER TABLE links ADD COLUMN group_id UUID REFERENCES groups(id) ON DELETE SET NULL;
UPDATE links SET group_id = lg.group_id FROM link_groups lg WHERE lg.link_id = links.id;
DROP TABLE IF EXISTS link_groups;
