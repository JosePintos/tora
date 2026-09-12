-- name: CreateDeck :one
INSERT INTO decks (
    id,
    owner_id,
    name,
    created_at
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetDeckByID :one
SELECT *
FROM decks
WHERE id = $1;

-- name: ListDecksByOwner :many
SELECT *
FROM decks
WHERE owner_id = $1
ORDER BY created_at;