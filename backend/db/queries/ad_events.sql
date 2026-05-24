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
    COUNT(*) FILTER (WHERE event_type = 'COMPLETION') as completions
FROM ad_events 
WHERE ad_id = $1;
