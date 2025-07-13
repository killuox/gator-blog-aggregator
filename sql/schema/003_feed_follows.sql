
-- +goose Up
CREATE TABLE feed_follows (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    feed_id UUID NOT NULL REFERENCES feeds (id) ON DELETE CASCADE,
    UNIQUE(user_id, feed_id)
);

CREATE INDEX idx_feed_follows_user_id ON feed_follows (user_id);
CREATE INDEX idx_feed_follows_feed_id ON feed_follows (feed_id);

-- +goose Down
DROP INDEX idx_feed_follows_feed_id
DROP INDEX idx_feed_follows_user_id

DROP TABLE feed_follows;