CREATE TABLE onboarding_acks (
    account_id uuid        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    item_id    text        NOT NULL,
    acked_at   timestamptz NOT NULL,
    PRIMARY KEY (account_id, item_id)
);
