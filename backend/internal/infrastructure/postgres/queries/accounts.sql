-- name: CreateAccount :exec
INSERT INTO accounts (id, email, created_at, updated_at, first_name, last_name, status, picture_url, english_level, about_me, native_language, show_translations)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: UpdateAccount :execrows
-- Basic profile update: picture_url is NOT changed here — only via UpdateAccountPicture.
UPDATE accounts
SET email             = $2,
    updated_at        = $3,
    first_name        = $4,
    last_name         = $5,
    status            = $6,
    english_level     = $7,
    about_me          = $8,
    native_language   = $9,
    show_translations = $10
WHERE id = $1;

-- name: UpdateAccountPicture :one
-- Atomically updates picture_url and returns the previous value for deleting the old file.
WITH prev AS (
    SELECT picture_url AS old_picture_url
    FROM accounts
    WHERE id = $1
)
UPDATE accounts
SET picture_url = $2,
    updated_at  = $3
FROM prev
WHERE accounts.id = $1
RETURNING prev.old_picture_url AS previous_picture_url;

-- name: DeleteAccount :execrows
DELETE FROM accounts WHERE id = $1;

-- name: GetAccountByID :one
SELECT id, email, created_at, updated_at, first_name, last_name, status, picture_url, english_level, about_me, native_language, show_translations
FROM accounts
WHERE id = $1;

-- name: GetAccountByEmail :one
SELECT id, email, created_at, updated_at, first_name, last_name, status, picture_url, english_level, about_me, native_language, show_translations
FROM accounts
WHERE lower(email) = lower($1);

-- name: ExistsAccountEmail :one
SELECT EXISTS (
    SELECT 1 FROM accounts WHERE lower(email) = lower($1)
) AS exists;

-- name: GetAccountsByIDs :many
SELECT id, email, created_at, updated_at, first_name, last_name, status, picture_url, english_level, about_me, native_language, show_translations
FROM accounts
WHERE id = ANY(@ids::uuid[]);

-- name: SearchAccountsByEmail :many
SELECT id, email, created_at, updated_at, first_name, last_name, status, picture_url, english_level, about_me, native_language, show_translations
FROM accounts
WHERE email ILIKE @query
ORDER BY email ASC
LIMIT @lim;
