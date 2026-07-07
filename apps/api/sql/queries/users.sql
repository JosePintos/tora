-- name: CreateUser :one
INSERT INTO users (
    id,
    username,
    password_hash,
    created_at
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY created_at;