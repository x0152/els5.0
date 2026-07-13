DROP TABLE IF EXISTS films;

CREATE TABLE films (
    id             uuid        PRIMARY KEY,
    title          text        NOT NULL,
    poster_path    text        NOT NULL DEFAULT '',
    duration_ms    int         NOT NULL DEFAULT 0,
    status         text        NOT NULL DEFAULT 'processing',
    error          text        NOT NULL DEFAULT '',
    audio_variants jsonb       NOT NULL DEFAULT '[]',
    subtitles      jsonb       NOT NULL DEFAULT '[]',
    created_at     timestamptz NOT NULL
);
