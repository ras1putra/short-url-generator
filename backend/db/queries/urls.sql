-- name: CreateURL :one
INSERT INTO urls (
  user_id, slug, original, custom, expires_at, is_monetized, allowed_categories
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
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
SET slug = $2, expires_at = $3, is_monetized = $5, allowed_categories = $6, updated_at = NOW()
WHERE id = $1 AND user_id = $4
RETURNING *;

-- name: DeleteURL :exec
DELETE FROM urls
WHERE id = $1 AND user_id = $2;

-- name: DeleteExpiredURLs :exec
DELETE FROM urls
WHERE expires_at IS NOT NULL AND expires_at < NOW();

-- name: ListURLsByUserFiltered :many
SELECT * FROM urls
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('q')::text IS NULL OR sqlc.narg('q') = '' OR slug ILIKE '%' || sqlc.narg('q') || '%' OR original ILIKE '%' || sqlc.narg('q') || '%')
  AND (sqlc.narg('is_monetized')::bool IS NULL OR is_monetized = sqlc.narg('is_monetized'))
ORDER BY
  CASE WHEN sqlc.arg('sort_by')::text = 'slug' AND sqlc.arg('sort_dir')::text = 'ASC' THEN slug END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'slug' AND sqlc.arg('sort_dir')::text = 'DESC' THEN slug END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'created_at' AND sqlc.arg('sort_dir')::text = 'ASC' THEN created_at::text END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'created_at' AND sqlc.arg('sort_dir')::text = 'DESC' THEN created_at::text END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'expires_at' AND sqlc.arg('sort_dir')::text = 'ASC' THEN expires_at::text END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'expires_at' AND sqlc.arg('sort_dir')::text = 'DESC' THEN expires_at::text END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'updated_at' AND sqlc.arg('sort_dir')::text = 'ASC' THEN updated_at::text END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'updated_at' AND sqlc.arg('sort_dir')::text = 'DESC' THEN updated_at::text END DESC NULLS LAST,
  created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountURLsByUserFiltered :one
SELECT COUNT(*) FROM urls
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('q')::text IS NULL OR sqlc.narg('q') = '' OR slug ILIKE '%' || sqlc.narg('q') || '%' OR original ILIKE '%' || sqlc.narg('q') || '%')
  AND (sqlc.narg('is_monetized')::bool IS NULL OR is_monetized = sqlc.narg('is_monetized'));