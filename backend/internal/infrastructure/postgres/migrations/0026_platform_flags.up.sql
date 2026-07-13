CREATE TABLE platform_flags (
    key        text        PRIMARY KEY,
    enabled    boolean     NOT NULL DEFAULT false,
    updated_at timestamptz NOT NULL DEFAULT now()
);
