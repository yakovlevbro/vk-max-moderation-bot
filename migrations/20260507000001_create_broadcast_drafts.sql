-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS broadcast_drafts (
    user_id BIGINT PRIMARY KEY,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS broadcast_drafts;
-- +goose StatementEnd
