-- name: CreateCard :one
INSERT INTO cards (id, deck_id, front, back, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListCardsByDeck :many
SELECT *
FROM cards
WHERE deck_id = $1
-- id breaks ties so the order is deterministic even when timestamps are equal.
ORDER BY created_at ASC, id ASC;

-- name: ListCardsByDecks :many
-- Cards of several decks in one query (used by the DataLoader for Deck.cards).
-- Same order as ListCardsByDeck within each deck.
SELECT *
FROM cards
WHERE deck_id = ANY(@deck_ids::uuid[])
ORDER BY deck_id, created_at ASC, id ASC;
