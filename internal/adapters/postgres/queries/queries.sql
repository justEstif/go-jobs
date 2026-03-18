-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
SET name = $1, email = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $3;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
