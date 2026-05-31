-- name: CreateUser :one
INSERT INTO users (
  name, email, password, role
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: UpdateUserRole :one
UPDATE users SET role = $2 WHERE id = $1 RETURNING *;

-- name: GetUserByEmailVerificationToken :one
SELECT * FROM users
WHERE email_verification_token = $1 LIMIT 1;

-- name: GetUserByPasswordResetToken :one
SELECT * FROM users
WHERE password_reset_token = $1 LIMIT 1;

-- name: UpdateUserEmailVerified :one
UPDATE users SET email_verified = TRUE, email_verification_token = NULL, email_verification_sent_at = NULL WHERE id = $1 RETURNING *;

-- name: UpdateUserEmailVerificationToken :one
UPDATE users SET email_verification_token = $2, email_verification_sent_at = NOW() WHERE id = $1 RETURNING *;

-- name: UpdateUserPasswordResetToken :one
UPDATE users SET password_reset_token = $2, password_reset_sent_at = NOW() WHERE id = $1 RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users SET password = $2, password_reset_token = NULL, password_reset_sent_at = NULL WHERE id = $1 RETURNING *;
