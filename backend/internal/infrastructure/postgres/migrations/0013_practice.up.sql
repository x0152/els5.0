CREATE TABLE practice_variants (
    id         uuid        PRIMARY KEY,
    account_id uuid        NOT NULL,
    kind       text        NOT NULL,
    number     int         NOT NULL,
    title      text        NOT NULL DEFAULT '',
    exercises  text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL
);

CREATE INDEX practice_variants_owner_idx ON practice_variants (account_id, kind, number, created_at);

CREATE TABLE practice_progress (
    account_id  uuid        NOT NULL,
    kind        text        NOT NULL,
    number      int         NOT NULL,
    variant_key text        NOT NULL,
    answers     jsonb       NOT NULL DEFAULT '{}',
    completed   boolean     NOT NULL DEFAULT false,
    updated_at  timestamptz NOT NULL,
    PRIMARY KEY (account_id, kind, number, variant_key)
);
