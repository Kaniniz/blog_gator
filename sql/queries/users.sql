-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users 
WHERE name = $1;

-- name: DropUsers :exec
DROP TABLE users;

-- name: CreateUsersTable :exec
CREATE TABLE users (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	name TEXT UNIQUE
);

-- name: GetUsers :many
SELECT * FROM users;