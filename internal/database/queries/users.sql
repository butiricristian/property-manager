-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, created_at;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, created_at, name, avatar_url, user_type FROM users WHERE id = $1;

-- name: GetUserByGoogleID :one
SELECT id, email, created_at, name, avatar_url FROM users WHERE google_id = $1;

-- name: CreateGoogleUser :one
INSERT INTO users (email, google_id, name, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING id, email, created_at, name, avatar_url;

-- name: LinkGoogleAccount :exec
UPDATE users SET google_id = $2, name = $3, avatar_url = $4 WHERE id = $1;

-- name: CreateGuestUser :one
INSERT INTO users (email, user_type)
VALUES ($1, 'guest')
RETURNING id, email, created_at, user_type;

-- name: UpgradeGuestUser :one
UPDATE users
SET email = $2, password_hash = $3, user_type = 'free'
WHERE id = $1 AND user_type = 'guest'
RETURNING id, email, created_at, name, avatar_url, user_type;

-- name: GetUserType :one
SELECT user_type FROM users WHERE id = $1;

-- name: CountProperties :one
SELECT count(*) FROM properties WHERE user_id = $1;

-- name: DeleteOldGuestUsers :exec
DELETE FROM users WHERE user_type = 'guest' AND created_at < NOW() - INTERVAL '7 days';
