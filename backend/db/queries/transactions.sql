-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id, amount, type, tx_hash, metadata
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: ListTransactionsByUser :many
SELECT * FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
