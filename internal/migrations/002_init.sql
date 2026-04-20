-- +goose Up
ALTER TABLE users ADD COLUMN tg_chat_id BIGINT UNIQUE;

CREATE TABLE IF NOT EXISTS tg_links
(
    token      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS tg_links;
ALTER TABLE users DROP COLUMN IF EXISTS tg_chat_id;