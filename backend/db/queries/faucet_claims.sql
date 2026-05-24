-- name: CreateFaucetClaim :one
INSERT INTO faucet_claims (user_id, amount, tx_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetFaucetClaimByUser :many
SELECT * FROM faucet_claims
WHERE user_id = $1
ORDER BY claimed_at DESC
LIMIT $2 OFFSET $3;

-- name: CountFaucetClaims :one
SELECT COUNT(*) FROM faucet_claims
WHERE user_id = $1;

-- name: CountFaucetClaimsToday :one
SELECT COUNT(*) FROM faucet_claims
WHERE user_id = $1
  AND claimed_at >= NOW() - INTERVAL '24 hours';
