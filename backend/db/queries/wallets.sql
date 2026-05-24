-- name: GetWalletByUserID :one
SELECT * FROM wallets WHERE user_id = $1 LIMIT 1;

-- name: CreateWallet :exec
INSERT INTO wallets (user_id, balance) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: UpdateWalletBalance :one
UPDATE wallets 
SET balance = balance + $2, updated_at = NOW() 
WHERE user_id = $1 
RETURNING *;
