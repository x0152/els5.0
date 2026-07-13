-- name: CreateAdministrator :exec
INSERT INTO administrators (id, account_id, created_at, updated_at)
VALUES ($1, $2, $3, $4);

-- name: UpdateAdministrator :execrows
UPDATE administrators
SET updated_at = $2
WHERE id = $1
  AND deleted_at IS NULL;

-- name: DeleteAdministrator :execrows
-- Soft-delete: the row remains but is hidden from all Read queries.
UPDATE administrators
SET deleted_at = $2,
    updated_at = $2
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetAdministratorByID :one
SELECT * FROM administrators
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetAdministratorByAccountID :one
SELECT * FROM administrators
WHERE account_id = $1
  AND deleted_at IS NULL;

-- name: ListAdministrators :many
SELECT * FROM administrators
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAdministrators :one
SELECT COUNT(*) FROM administrators WHERE deleted_at IS NULL;
