-- name: UpsertCredentials :exec
INSERT INTO credentials (account_id, password_hash, updated_at)
VALUES ($1, $2, $3)
ON CONFLICT (account_id) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    updated_at    = EXCLUDED.updated_at;

-- name: GetCredentialsByAccountID :one
SELECT account_id, password_hash, updated_at
FROM credentials
WHERE account_id = $1;
