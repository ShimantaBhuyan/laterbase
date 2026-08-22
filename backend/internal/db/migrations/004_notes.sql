-- +goose Up
ALTER TABLE links ADD COLUMN notes TEXT;

-- +goose Down
ALTER TABLE links DROP COLUMN IF EXISTS notes;
