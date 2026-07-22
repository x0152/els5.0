CREATE TABLE onboarding_metrics (
    account_id uuid        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    metric     text        NOT NULL,
    value      int         NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (account_id, metric)
);
