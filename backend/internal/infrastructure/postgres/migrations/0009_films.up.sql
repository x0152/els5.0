CREATE TABLE films (
    id          uuid        PRIMARY KEY,
    title       text        NOT NULL,
    video_path  text        NOT NULL,
    poster_path text        NOT NULL DEFAULT '',
    duration_ms int         NOT NULL DEFAULT 0,
    subtitles   jsonb       NOT NULL DEFAULT '[]',
    created_at  timestamptz NOT NULL
);
