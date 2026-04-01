-- name: ListIntegrationsByUser :many
SELECT * FROM integrations WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetIntegration :one
SELECT * FROM integrations WHERE id = $1;

-- name: GetIntegrationByPlatform :one
SELECT * FROM integrations WHERE user_id = $1 AND platform = $2;

-- name: CreateIntegration :one
INSERT INTO integrations (user_id, platform, credentials, enabled)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateIntegration :one
UPDATE integrations
SET credentials = $2, enabled = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteIntegration :exec
DELETE FROM integrations WHERE id = $1;

-- name: UpdateIntegrationSyncStatus :exec
UPDATE integrations
SET last_synced_at = $2, last_sync_error = $3, updated_at = NOW()
WHERE id = $1;

-- name: ListEnabledIntegrations :many
SELECT * FROM integrations WHERE enabled = true;
