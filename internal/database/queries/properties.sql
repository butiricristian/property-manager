-- name: ListProperties :many
SELECT * FROM properties ORDER BY created_at DESC;

-- name: GetProperty :one
SELECT * FROM properties WHERE id = $1;

-- name: CreateProperty :one
INSERT INTO properties (name, address)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateProperty :one
UPDATE properties
SET name = $2, address = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProperty :exec
DELETE FROM properties WHERE id = $1;
