-- +goose Up
ALTER TABLE links ADD COLUMN input_tokens  INTEGER;
ALTER TABLE links ADD COLUMN output_tokens INTEGER;
ALTER TABLE links ADD COLUMN llm_cost     NUMERIC(10,6);

-- +goose Down
ALTER TABLE links DROP COLUMN IF EXISTS llm_cost;
ALTER TABLE links DROP COLUMN IF EXISTS output_tokens;
ALTER TABLE links DROP COLUMN IF EXISTS input_tokens;
