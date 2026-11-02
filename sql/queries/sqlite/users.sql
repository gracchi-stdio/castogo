-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ?;

-- name: GetUserByEmail :one 
SELECT *
FROM users
WHERE email = ?;

-- name: CreateUser :one 
INSERT INTO users (
    email, password_hash
) VALUES (?, ?) RETURNING *;

-- name: UpdateUser :one 
UPDATE users
SET email = ?, password_hash = ?, updated_at = datetime('now')
WHERE id = ? RETURNING *;

-- name: DeleteUser :exec 
DELETE FROM users
WHERE id = ?;
