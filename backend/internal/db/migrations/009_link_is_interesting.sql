-- +goose Up
ALTER TABLE links ADD COLUMN is_interesting BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE links DROP COLUMN IF EXISTS is_interesting;
