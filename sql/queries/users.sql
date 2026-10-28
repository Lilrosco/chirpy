-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: UpdateUser :one
UPDATE
    users
SET
    updated_at = NOW(), email = $1, hashed_password = $2
WHERE
    id = $3
RETURNING *;

-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    email = $1;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT
    *
FROM
    users;

-- name: UpgradeToChirpyRed :one
UPDATE
    users
SET
    updated_at = NOW(), is_chirpy_red = true
WHERE
    id = $1
RETURNING *;

-- name: DowngradeFromChirpyRed :one
UPDATE
    users
SET
    updated_at = NOW(), is_chirpy_red = false
WHERE
    id = $1
RETURNING *;
