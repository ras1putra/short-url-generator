-- name: CreateAdEvent :one
INSERT INTO ad_events (
    ad_id, link_id, event_type, is_valid, quality_score, rejection_reason, ip_address, user_agent, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetAdEventStats :one
SELECT 
    COUNT(*) FILTER (WHERE event_type = 'IMPRESSION') as impressions,
    COUNT(*) FILTER (WHERE event_type = 'CLICK') as clicks,
    COUNT(*) FILTER (WHERE event_type = 'COMPLETION') as completions,
    COUNT(*) FILTER (WHERE event_type = 'COMPLETION' AND is_valid = true) as valid_completions,
    COUNT(*) FILTER (WHERE event_type = 'COMPLETION' AND is_valid = false) as invalid_completions,
    COUNT(*) FILTER (WHERE event_type = 'SKIP') as skips,
    COALESCE(AVG(quality_score) FILTER (WHERE event_type = 'COMPLETION' AND is_valid = true), 0.0)::float as avg_quality_score
FROM ad_events 
WHERE ad_id = $1;

-- name: GetLinkAdEvents :many
SELECT 
    ae.created_at as time,
    ae.event_type,
    ae.is_valid,
    ae.quality_score,
    ae.rejection_reason,
    a.title as ad_title,
    a.ad_type
FROM ad_events ae
JOIN ads a ON ae.ad_id = a.id
JOIN urls u ON ae.link_id = u.id
WHERE u.slug = $1
ORDER BY ae.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountLinkAdEvents :one
SELECT COUNT(*) 
FROM ad_events ae
JOIN urls u ON ae.link_id = u.id
WHERE u.slug = $1;

-- name: GetLinkAdEventsFiltered :many
SELECT 
    ae.created_at as time,
    ae.event_type,
    ae.is_valid,
    ae.quality_score,
    ae.rejection_reason,
    a.title as ad_title,
    a.ad_type
FROM ad_events ae
JOIN ads a ON ae.ad_id = a.id
JOIN urls u ON ae.link_id = u.id
WHERE u.slug = sqlc.arg('slug')
ORDER BY
  CASE WHEN sqlc.arg('sort_by')::text = 'time' AND sqlc.arg('sort_dir')::text = 'asc' THEN ae.created_at::text END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'time' AND sqlc.arg('sort_dir')::text = 'desc' THEN ae.created_at::text END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'event_type' AND sqlc.arg('sort_dir')::text = 'asc' THEN ae.event_type END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'event_type' AND sqlc.arg('sort_dir')::text = 'desc' THEN ae.event_type END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'ad_title' AND sqlc.arg('sort_dir')::text = 'asc' THEN a.title END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'ad_title' AND sqlc.arg('sort_dir')::text = 'desc' THEN a.title END DESC NULLS LAST,
  ae.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountLinkAdEventsFiltered :one
SELECT COUNT(*) 
FROM ad_events ae
JOIN ads a ON ae.ad_id = a.id
JOIN urls u ON ae.link_id = u.id
WHERE u.slug = sqlc.arg('slug');
