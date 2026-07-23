CREATE TABLE onboarding_tours (
    account_id uuid        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    tour_id    text        NOT NULL,
    done_at    timestamptz NOT NULL,
    PRIMARY KEY (account_id, tour_id)
);
