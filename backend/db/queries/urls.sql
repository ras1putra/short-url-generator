-- name: CreateURL :one
INSERT INTO urls (
  user_id, slug, original, custom, expires_at
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetURLBySlug :one
SELECT * FROM urls
WHERE slug = $1 LIMIT 1;

-- name: ListURLsByUser :many
SELECT * FROM urls
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListURLsByUserPaginated :many
SELECT * FROM urls
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountURLsByUser :one
SELECT COUNT(*) FROM urls WHERE user_id = $1;

-- name: UpdateURL :one
UPDATE urls
SET slug = $2, expires_at = $3, updated_at = NOW()
WHERE id = $1 AND user_id = $4
RETURNING *;

-- name: DeleteURL :exec
DELETE FROM urls
WHERE id = $1 AND user_id = $2;

-- name: DeleteExpiredURLs :exec
DELETE FROM urls
WHERE expires_at IS NOT NULL AND expires_at < NOW();