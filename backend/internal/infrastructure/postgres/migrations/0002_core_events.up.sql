CREATE TABLE raw_events (
    id           uuid        PRIMARY KEY,
    user_id      uuid        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    status       text        NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processed', 'failed')),
    version      int         NOT NULL DEFAULT 1,
    type         text        NOT NULL,
    text         text        NOT NULL DEFAULT '',
    context      text        NOT NULL DEFAULT '',
    note         text        NOT NULL DEFAULT '',
    source       jsonb       NOT NULL DEFAULT '{}',
    meta         jsonb       NOT NULL DEFAULT '{}',
    occurred_at  timestamptz NOT NULL,
    created_at   timestamptz NOT NULL,
    processed_at timestamptz,
    error        text        NOT NULL DEFAULT ''
);

CREATE INDEX raw_events_user_status_idx ON raw_events (user_id, status, created_at DESC);
CREATE INDEX raw_events_pending_idx ON raw_events (created_at) WHERE status = 'pending';

CREATE TABLE events (
    id           uuid        PRIMARY KEY,
    raw_event_id uuid        NOT NULL REFERENCES raw_events(id) ON DELETE CASCADE,
    user_id      uuid        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    type         text        NOT NULL DEFAULT '',
    action       text        NOT NULL DEFAULT '',
    lemma        text        NOT NULL DEFAULT '',
    pos          text        NOT NULL DEFAULT '',
    grammar_key  text        NOT NULL DEFAULT '',
    outcome      text        NOT NULL DEFAULT '',
    error        jsonb,
    context      text        NOT NULL DEFAULT '',
    occurred_at  timestamptz NOT NULL,
    created_at   timestamptz NOT NULL
);

CREATE INDEX events_user_idx ON events (user_id, created_at DESC);
CREATE INDEX events_user_lemma_idx ON events (user_id, lemma);
CREATE INDEX events_user_grammar_idx ON events (user_id, grammar_key);
