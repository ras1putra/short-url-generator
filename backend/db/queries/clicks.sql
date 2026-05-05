-- name: SaveClick :one
INSERT INTO clicks (
  url_id, ip_hash, country, city, device, browser, referrer
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetStatsBySlug :many
SELECT
    c.country,
    c.device,
    c.browser,
    DATE(c.clicked_at) as click_date,
    COUNT(c.id) as click_count
FROM clicks c
JOIN urls u ON c.url_id = u.id
WHERE u.slug = $1
GROUP BY c.country, c.device, c.browser, DATE(c.clicked_at)
ORDER BY click_date DESC;

-- name: GetTotalClicksBySlug :one
SELECT COUNT(*) FROM clicks c
JOIN urls u ON c.url_id = u.id
WHERE u.slug = $1;

-- name: GetUniqueClicksBySlug :one
SELECT COUNT(DISTINCT c.ip_hash) FROM clicks c
JOIN urls u ON c.url_id = u.id
WHERE u.slug = $1 AND c.ip_hash IS NOT NULL;

-- name: GetAggregateStatsByUser :many
SELECT
    c.country,
    c.device,
    c.browser,
    DATE(c.clicked_at) as click_date,
    COUNT(c.id) as click_count
FROM clicks c
JOIN urls u ON c.url_id = u.id
WHERE u.user_id = $1
GROUP BY c.country, c.device, c.browser, DATE(c.clicked_at)
ORDER BY click_date DESC;

-- name: GetTotalClicksByUser :one
SELECT COUNT(*) FROM clicks c
JOIN urls u ON c.url_id = u.id
WHERE u.user_id = $1;

-- name: GetUniqueClicksByUser :one
SELECT COUNT(DISTINCT c.ip_hash) FROM clicks c
JOIN urls u ON c.url_id = u.id
WHERE u.user_id = $1 AND c.ip_hash IS NOT NULL;
