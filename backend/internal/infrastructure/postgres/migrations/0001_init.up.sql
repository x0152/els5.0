-- Base platform schema: accounts (IAM), credentials, administrators
-- and a view for computing the account role. Everything specific to the previous platform
-- (clients, experts, customers, projects, tasks, timesheets, calendar) was removed.
--
-- Column order matters: sqlc-generated code uses SELECT *,
-- and structs in sqlc/models.go expect exactly this order.

CREATE TABLE accounts (
    id          uuid        PRIMARY KEY,
    email       text        NOT NULL,
    created_at  timestamptz NOT NULL,
    updated_at  timestamptz NOT NULL,
    first_name  text        NOT NULL DEFAULT '',
    last_name   text        NOT NULL DEFAULT '',
    status      text        NOT NULL DEFAULT 'active'
        CHECK (status IN ('pending_password', 'active', 'blocked', 'no_auth')),
    picture_url text        NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX accounts_email_lower_uniq ON accounts (lower(email));

CREATE TABLE credentials (
    account_id    uuid        PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    password_hash text        NOT NULL,
    updated_at    timestamptz NOT NULL
);

CREATE TABLE administrators (
    id         uuid        PRIMARY KEY,
    account_id uuid        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE UNIQUE INDEX administrators_account_id_uniq
    ON administrators (account_id)
    WHERE deleted_at IS NULL;

CREATE INDEX administrators_deleted_at_idx ON administrators (deleted_at);

-- Account role is computed from the presence of a related entity. In the base template
-- the only role is administrator. New roles are added as UNION ALL.
CREATE VIEW v_account_role AS
SELECT account_id, 'admin'::text AS role, id AS entity_id
FROM administrators
WHERE deleted_at IS NULL;
