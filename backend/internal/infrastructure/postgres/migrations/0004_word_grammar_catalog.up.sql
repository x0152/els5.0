CREATE TABLE words (
    id         uuid        PRIMARY KEY,
    key        text        NOT NULL UNIQUE,
    lemma      text        NOT NULL,
    pos        text        NOT NULL DEFAULT '',
    type       text        NOT NULL DEFAULT 'word',
    enriched   boolean     NOT NULL DEFAULT false,
    is_stop    boolean     NOT NULL DEFAULT false,
    cefr       text        NOT NULL DEFAULT '',
    frequency  double precision,
    metadata   jsonb       NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE grammar_rules (
    id         uuid        PRIMARY KEY,
    key        text        NOT NULL UNIQUE,
    parent_key text        NOT NULL DEFAULT '',
    title      text        NOT NULL DEFAULT '',
    cefr_level text        NOT NULL DEFAULT '',
    enriched   boolean     NOT NULL DEFAULT false,
    metadata   jsonb       NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

ALTER TABLE events
    ADD COLUMN word_id         uuid REFERENCES words(id),
    ADD COLUMN grammar_rule_id uuid REFERENCES grammar_rules(id);

CREATE INDEX events_word_idx ON events (word_id) WHERE word_id IS NOT NULL;
CREATE INDEX events_grammar_rule_idx ON events (grammar_rule_id) WHERE grammar_rule_id IS NOT NULL;
