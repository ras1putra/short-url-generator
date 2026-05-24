-- name: CreateAd :one
INSERT INTO ads (
    advertiser_id, title, description, image_url, target_url, category, total_budget, remaining_budget, cpm, ad_type
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetAdByID :one
SELECT * FROM ads WHERE id = $1 LIMIT 1;

-- name: ListAdsByAdvertiser :many
SELECT * FROM ads WHERE advertiser_id = $1 ORDER BY created_at DESC;

-- name: GetActiveAds :many
SELECT * FROM ads 
WHERE status = 'active' 
AND remaining_budget > 0 
ORDER BY cpm DESC;

-- name: GetActiveAdsByCategory :many
SELECT * FROM ads 
WHERE status = 'active' 
AND remaining_budget > 0 
AND category = ANY($1::text[])
ORDER BY cpm DESC;

-- name: UpdateAdStatus :exec
UPDATE ads SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: DeductAdBudget :exec
UPDATE ads 
SET remaining_budget = remaining_budget - $2, updated_at = NOW() 
WHERE id = $1 AND remaining_budget >= $2;

-- name: GetCPMByAdType :one
SELECT cpm FROM ad_cpm_rates WHERE ad_type = $1 LIMIT 1;

-- name: ListAdTypes :many
SELECT t.ad_type, c.cpm, t.label, t.aspect_ratio, t.recommended_resolution
FROM ad_types t
LEFT JOIN ad_cpm_rates c ON t.ad_type = c.ad_type
ORDER BY c.cpm ASC;

-- name: UpdateAd :one
UPDATE ads SET
    title = $2,
    description = $3,
    image_url = $4,
    target_url = $5,
    category = $6,
    total_budget = $7,
    remaining_budget = $8,
    cpm = $9,
    status = $10,
    ad_type = $11,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetReferencedMediaURLs :many
SELECT image_url
FROM ads
WHERE image_url IS NOT NULL
  AND image_url <> ''
  AND status <> 'deleted';
