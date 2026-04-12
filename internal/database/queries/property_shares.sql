-- name: ListPropertySharesWithUsers :many
SELECT ps.user_id, u.email, u.name, ps.permission
FROM property_shares ps
JOIN users u ON u.id = ps.user_id
JOIN properties p ON p.id = ps.property_id
WHERE ps.property_id = $1 AND p.user_id = $2
ORDER BY ps.created_at ASC;

-- name: CreatePropertyShare :one
INSERT INTO property_shares (property_id, user_id, permission)
VALUES ($1, $2, $3)
RETURNING id, property_id, user_id, permission, created_at;

-- name: DeletePropertyShare :exec
DELETE FROM property_shares ps
WHERE ps.property_id = $1
  AND ps.user_id = $2
  AND ps.property_id IN (SELECT prp.id FROM properties prp WHERE prp.user_id = $3);

-- name: UpdatePropertyShare :one
UPDATE property_shares ps
SET permission = $3
WHERE ps.property_id = $1
  AND ps.user_id = $2
  AND ps.property_id IN (SELECT prp.id FROM properties prp WHERE prp.user_id = $4)
RETURNING ps.id, ps.property_id, ps.user_id, ps.permission, ps.created_at;

-- name: GetPropertyShare :one
SELECT id, property_id, user_id, permission, created_at
FROM property_shares
WHERE property_id = $1 AND user_id = $2;
