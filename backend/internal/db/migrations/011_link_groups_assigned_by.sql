-- +goose Up
ALTER TABLE link_groups ADD COLUMN assigned_by TEXT NOT NULL DEFAULT 'auto';

-- Existing rows predate provenance tracking; treat them as auto-assigned so a
-- bulk reprocess gives every link a clean, fully re-categorized slate.
ALTER TABLE link_groups DROP CONSTRAINT IF EXISTS link_groups_assigned_by_check;
ALTER TABLE link_groups ADD CONSTRAINT link_groups_assigned_by_check CHECK (assigned_by IN ('auto', 'manual'));

-- +goose Down
ALTER TABLE link_groups DROP COLUMN IF EXISTS assigned_by;
