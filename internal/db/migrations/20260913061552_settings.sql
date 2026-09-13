-- +goose Up
CREATE TABLE
    setting (
        key TEXT PRIMARY KEY,
        value TEXT NOT NULL,
        modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

-- +goose Down
DROP TABLE IF EXISTS setting;