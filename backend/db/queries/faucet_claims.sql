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

-- name: GetFaucetClaimByUserFiltered :many
SELECT * FROM faucet_claims
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('q')::text IS NULL OR sqlc.narg('q') = '' OR tx_hash::text ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY
  CASE WHEN sqlc.arg('sort_by')::text = 'amount' AND sqlc.arg('sort_dir')::text = 'ASC' THEN amount::text END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'amount' AND sqlc.arg('sort_dir')::text = 'DESC' THEN amount::text END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'tx_hash' AND sqlc.arg('sort_dir')::text = 'ASC' THEN tx_hash::text END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'tx_hash' AND sqlc.arg('sort_dir')::text = 'DESC' THEN tx_hash::text END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'claimed_at' AND sqlc.arg('sort_dir')::text = 'ASC' THEN claimed_at::text END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'claimed_at' AND sqlc.arg('sort_dir')::text = 'DESC' THEN claimed_at::text END DESC NULLS LAST,
  claimed_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountFaucetClaimsByUserFiltered :one
SELECT COUNT(*) FROM faucet_claims
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('q')::text IS NULL OR sqlc.narg('q') = '' OR tx_hash::text ILIKE '%' || sqlc.narg('q') || '%');
