-- name: CreateOAuthAccount :exec
INSERT INTO oauth_accounts (user_id, provider, provider_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: GetOAuthAccountByProvider :one
SELECT * FROM oauth_accounts
WHERE provider = $1 AND provider_id = $2
LIMIT 1;

-- name: ListOAuthAccountsByUserID :many
SELECT * FROM oauth_accounts WHERE user_id = $1;

-- name: DeleteOAuthAccount :exec
DELETE FROM oauth_accounts WHERE id = $1;
