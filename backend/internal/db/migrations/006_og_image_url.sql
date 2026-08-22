-- +goose Up
ALTER TABLE links ADD COLUMN og_image_url TEXT;

-- +goose Down
ALTER TABLE links DROP COLUMN IF EXISTS og_image_url;
