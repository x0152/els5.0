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

CREATE TABLE word_lessons (
    id         uuid        PRIMARY KEY,
    number     int         NOT NULL UNIQUE,
    title      text        NOT NULL,
    page       int         NOT NULL DEFAULT 0,
    words      jsonb       NOT NULL DEFAULT '[]',
    footer     text        NOT NULL DEFAULT '',
    theory     text        NOT NULL DEFAULT '',
    exercises  text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

INSERT INTO merfy_units (id, number, title, page, footer, theory, exercises, created_at, updated_at)
SELECT id, number, title, page, footer, theory, exercises, created_at, updated_at FROM book_chapters WHERE book = 'merfy';

INSERT INTO word_lessons (id, number, title, page, words, footer, theory, exercises, created_at, updated_at)
SELECT id, number, title, page, words, footer, theory, exercises, created_at, updated_at FROM book_chapters WHERE book = 'words';

DROP TABLE book_chapters;
DROP TABLE books;
