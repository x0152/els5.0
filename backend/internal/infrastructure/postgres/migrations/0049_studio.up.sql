CREATE TABLE studio_areas (
    id         uuid        PRIMARY KEY,
    account_id uuid        NOT NULL,
    title      text        NOT NULL,
    emoji      text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL
);

CREATE INDEX studio_areas_account_idx ON studio_areas (account_id, created_at);

CREATE TABLE studio_items (
    id            uuid        PRIMARY KEY,
    area_id       uuid        NOT NULL REFERENCES studio_areas (id) ON DELETE CASCADE,
    account_id    uuid        NOT NULL,
    text          text        NOT NULL,
    transcription text        NOT NULL DEFAULT '',
    translation   text        NOT NULL DEFAULT '',
    example       text        NOT NULL DEFAULT '',
    task          text        NOT NULL DEFAULT '',
    listened      boolean     NOT NULL DEFAULT false,
    spoken        boolean     NOT NULL DEFAULT false,
    written       boolean     NOT NULL DEFAULT false,
    created_at    timestamptz NOT NULL
);

CREATE INDEX studio_items_area_idx ON studio_items (area_id, created_at);
