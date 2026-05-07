-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS broadcast_selections (
    user_id BIGINT NOT NULL,
    chat_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, chat_id)
);
CREATE INDEX IF NOT EXISTS idx_broadcast_selections_user_id ON broadcast_selections(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS broadcast_selections;
-- +goose StatementEnd
