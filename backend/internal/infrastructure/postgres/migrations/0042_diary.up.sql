CREATE TABLE diary_entries (
    id            uuid        PRIMARY KEY,
    account_id    uuid        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    entry_date    date        NOT NULL,
    question      text        NOT NULL DEFAULT '',
    text          text        NOT NULL,
    reply         text        NOT NULL DEFAULT '',
    next_question text        NOT NULL DEFAULT '',
    native_sample text        NOT NULL DEFAULT '',
    corrections   jsonb       NOT NULL DEFAULT '[]',
    created_at    timestamptz NOT NULL,
    UNIQUE (account_id, entry_date)
);

CREATE INDEX diary_entries_account_date_idx ON diary_entries (account_id, entry_date DESC);
