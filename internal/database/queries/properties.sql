-- name: ListProperties :many
SELECT * FROM properties WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ListAccessibleProperties :many
SELECT p.id, p.name, p.address, p.created_at, p.updated_at, p.user_id, 'admin'::text AS permission
FROM properties p
WHERE p.user_id = $1
UNION ALL
SELECT p.id, p.name, p.address, p.created_at, p.updated_at, p.user_id, ps.permission
FROM properties p
JOIN property_shares ps ON ps.property_id = p.id
WHERE ps.user_id = $1
ORDER BY created_at DESC;

-- name: GetProperty :one
SELECT * FROM properties WHERE id = $1 AND user_id = $2;

-- name: CreateProperty :one
INSERT INTO properties (name, address, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateProperty :one
UPDATE properties
SET name = $2, address = $3, updated_at = NOW()
WHERE id = $1 AND user_id = $4
RETURNING *;

-- name: DeleteProperty :exec
DELETE FROM properties WHERE id = $1 AND user_id = $2;

-- name: VerifyPropertyOwnership :one
SELECT EXISTS(SELECT 1 FROM properties WHERE id = $1 AND user_id = $2);

-- name: VerifyPropertyAccess :one
SELECT EXISTS(
    SELECT 1 FROM properties p WHERE p.id = $1 AND p.user_id = $2
    UNION ALL
    SELECT 1 FROM property_shares ps WHERE ps.property_id = $1 AND ps.user_id = $2
);
