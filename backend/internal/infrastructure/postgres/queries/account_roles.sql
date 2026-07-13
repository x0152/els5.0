-- name: GetAccountRoleByAccount :one
SELECT account_id, role, entity_id
FROM v_account_role
WHERE account_id = $1;
