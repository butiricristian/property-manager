-- name: ListMappingsByIntegration :many
SELECT * FROM property_mappings
WHERE integration_id = $1
ORDER BY created_at;

-- name: GetMappingByExternalID :one
SELECT * FROM property_mappings
WHERE integration_id = $1 AND external_id = $2;

-- name: CreateMapping :one
INSERT INTO property_mappings (integration_id, property_id, external_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteMapping :exec
DELETE FROM property_mappings WHERE id = $1;

-- name: ListMappingsByProperty :many
SELECT pm.id, pm.integration_id, pm.property_id, pm.external_id, pm.created_at,
       i.platform
FROM property_mappings pm
JOIN integrations i ON i.id = pm.integration_id
WHERE pm.property_id = $1
ORDER BY pm.created_at;
