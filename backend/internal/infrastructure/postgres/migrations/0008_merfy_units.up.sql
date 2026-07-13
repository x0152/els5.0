CREATE TABLE merfy_units (
    id         uuid        PRIMARY KEY,
    number     int         NOT NULL UNIQUE,
    title      text        NOT NULL,
    page       int         NOT NULL DEFAULT 0,
    footer     text        NOT NULL DEFAULT '',
    theory     text        NOT NULL DEFAULT '',
    exercises  text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);
