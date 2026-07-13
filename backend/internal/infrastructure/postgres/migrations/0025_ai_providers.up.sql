CREATE TABLE ai_providers (
    feature    text        PRIMARY KEY,
    base_url   text        NOT NULL DEFAULT '',
    api_key    text        NOT NULL DEFAULT '',
    model      text        NOT NULL DEFAULT '',
    updated_at timestamptz NOT NULL DEFAULT now()
);
