-- name: ListAdCategories :many
SELECT category, label, multiplier FROM ad_categories ORDER BY category;

-- name: GetCategoryMultiplier :one
SELECT multiplier FROM ad_categories WHERE category = $1 LIMIT 1;
