-- +goose Up
CREATE TABLE link_costs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id       UUID REFERENCES links(id) ON DELETE SET NULL,
    url           TEXT NOT NULL,
    input_tokens  INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    llm_cost      NUMERIC(10,6) NOT NULL DEFAULT 0,
    model_used    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Backfill existing cost data from processed links
INSERT INTO link_costs (link_id, url, input_tokens, output_tokens, llm_cost, created_at)
SELECT id, url, COALESCE(input_tokens, 0), COALESCE(output_tokens, 0), llm_cost, COALESCE(processed_at, NOW())
FROM links
WHERE llm_cost IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS link_costs;
