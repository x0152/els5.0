CREATE TABLE ai_sessions (
    id                 uuid        PRIMARY KEY,
    account_id         uuid        NOT NULL UNIQUE,
    model              text        NOT NULL DEFAULT '',
    context_started_at timestamptz NOT NULL,
    created_at         timestamptz NOT NULL,
    updated_at         timestamptz NOT NULL
);

CREATE TABLE ai_messages (
    id                uuid        PRIMARY KEY,
    session_id        uuid        NOT NULL REFERENCES ai_sessions(id) ON DELETE CASCADE,
    role              text        NOT NULL,
    content           text        NOT NULL DEFAULT '',
    tool_calls        jsonb       NOT NULL DEFAULT '[]',
    tool_call_id      text        NOT NULL DEFAULT '',
    tool_name         text        NOT NULL DEFAULT '',
    model             text        NOT NULL DEFAULT '',
    finish_reason     text        NOT NULL DEFAULT '',
    reasoning_content text        NOT NULL DEFAULT '',
    prompt_tokens     int         NOT NULL DEFAULT 0,
    completion_tokens int         NOT NULL DEFAULT 0,
    total_tokens      int         NOT NULL DEFAULT 0,
    created_at        timestamptz NOT NULL
);

CREATE INDEX ai_messages_session_created_idx ON ai_messages (session_id, created_at);
