
-- +goose Up
CREATE TABLE feeds (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    name VARCHAR(255) NOT NULL UNIQUE,
    url VARCHAR(2048) NOT NULL UNIQUE,
    user_id UUID REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_feeds_user_id ON feeds (user_id);

-- +goose Down
DROP INDEX idx_feeds_user_id;
DROP TABLE feeds;