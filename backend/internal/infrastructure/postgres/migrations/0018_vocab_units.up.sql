CREATE TABLE vocab_units (
    id            uuid        PRIMARY KEY,
    account_id    text        NOT NULL,
    text          text        NOT NULL,
    kind          text        NOT NULL DEFAULT 'word',
    transcription text        NOT NULL DEFAULT '',
    translation   text        NOT NULL DEFAULT '',
    definition    text        NOT NULL DEFAULT '',
    example       text        NOT NULL DEFAULT '',
    status        text        NOT NULL DEFAULT 'new',
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX vocab_units_account_text_idx ON vocab_units (account_id, lower(text));
CREATE INDEX vocab_units_account_created_idx ON vocab_units (account_id, created_at DESC);
