-- name: ListInventoryByProperty :many
SELECT * FROM inventory_items
WHERE property_id = $1
ORDER BY acquired_date DESC;

-- name: GetInventoryItem :one
SELECT * FROM inventory_items WHERE id = $1;

-- name: CreateInventoryItem :one
INSERT INTO inventory_items (property_id, name, category, quantity, acquired_date, lifetime_months, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateInventoryItem :one
UPDATE inventory_items
SET name = $2, category = $3, quantity = $4, acquired_date = $5, lifetime_months = $6, notes = $7
WHERE id = $1
RETURNING *;

-- name: DeleteInventoryItem :exec
DELETE FROM inventory_items WHERE id = $1;
