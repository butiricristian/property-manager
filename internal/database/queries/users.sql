-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, created_at;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, created_at, name, avatar_url FROM users WHERE id = $1;

-- name: GetUserByGoogleID :one
SELECT id, email, created_at, name, avatar_url FROM users WHERE google_id = $1;

-- name: CreateGoogleUser :one
INSERT INTO users (email, google_id, name, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING id, email, created_at, name, avatar_url;

-- name: LinkGoogleAccount :exec
UPDATE users SET google_id = $2, name = $3, avatar_url = $4 WHERE id = $1;
