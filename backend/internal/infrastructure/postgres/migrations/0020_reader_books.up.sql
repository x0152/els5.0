CREATE TABLE reader_books (
    id           uuid        PRIMARY KEY,
    owner_id     text        NOT NULL,
    title        text        NOT NULL,
    author       text        NOT NULL DEFAULT '',
    content_path text        NOT NULL DEFAULT '',
    text_length  integer     NOT NULL DEFAULT 0,
    position     integer     NOT NULL DEFAULT 0,
    status       text        NOT NULL DEFAULT 'processing',
    error        text        NOT NULL DEFAULT '',
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX reader_books_owner_created_idx ON reader_books (owner_id, created_at DESC);
