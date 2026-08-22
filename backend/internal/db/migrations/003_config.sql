-- +goose Up
CREATE TABLE config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT INTO config (key, value) VALUES
    ('fetcher_mode', 'simple');

-- +goose Down
DROP TABLE IF EXISTS config;
