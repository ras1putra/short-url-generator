-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id, amount, type, tx_hash, metadata
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: CreatePendingTransaction :one
INSERT INTO transactions (
    user_id, amount, type, tx_hash, metadata, status
) VALUES (
    $1, $2, $3, $4, $5, 'PENDING'::transaction_status
)
RETURNING *;

-- name: GetTransactionByHash :one
SELECT * FROM transactions WHERE tx_hash = $1;

-- name: UpdateTransactionStatus :one
UPDATE transactions
SET status = $2
WHERE tx_hash = $1
RETURNING *;

-- name: ListTransactionsByUser :many
SELECT * FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountTransactionsByUser :one
SELECT COUNT(*) FROM transactions WHERE user_id = $1;

-- name: ListTransactionsByUserFiltered :many
SELECT * FROM transactions
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('q')::text IS NULL OR sqlc.narg('q') = '' OR type ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY
  CASE WHEN sqlc.arg('sort_by')::text = 'amount' AND sqlc.arg('sort_dir')::text = 'asc' THEN amount END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'amount' AND sqlc.arg('sort_dir')::text = 'desc' THEN amount END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'type' AND sqlc.arg('sort_dir')::text = 'asc' THEN type END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'type' AND sqlc.arg('sort_dir')::text = 'desc' THEN type END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'status' AND sqlc.arg('sort_dir')::text = 'asc' THEN status END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'status' AND sqlc.arg('sort_dir')::text = 'desc' THEN status END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'created_at' AND sqlc.arg('sort_dir')::text = 'asc' THEN created_at::text END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_by')::text = 'created_at' AND sqlc.arg('sort_dir')::text = 'desc' THEN created_at::text END DESC NULLS LAST,
  created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountTransactionsByUserFiltered :one
SELECT COUNT(*) FROM transactions
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('q')::text IS NULL OR sqlc.narg('q') = '' OR type ILIKE '%' || sqlc.narg('q') || '%');
