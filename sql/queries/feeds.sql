-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT feeds.name, feeds.url, users.name AS user_name FROM feeds
INNER JOIN users ON users.id = feeds.user_id;

-- name: GetFeedByUrl :one
SELECT * FROM feeds
WHERE feeds.url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $1, updated_at = $1
WHERE feeds.id = $2;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY
    last_fetched_at ASC NULLS FIRST,
    last_fetched_at ASC
LIMIT 1;
