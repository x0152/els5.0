CREATE TABLE quest_missions (
    id         uuid        PRIMARY KEY,
    user_id    text        NOT NULL,
    status     text        NOT NULL DEFAULT 'generating',
    error      text        NOT NULL DEFAULT '',
    payload    jsonb       NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX quest_missions_user_idx ON quest_missions (user_id, created_at DESC);

CREATE TABLE quest_profiles (
    user_id    text        PRIMARY KEY,
    payload    jsonb       NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);
