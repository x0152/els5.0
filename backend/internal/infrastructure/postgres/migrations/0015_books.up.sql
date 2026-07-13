CREATE TABLE books (
    slug       text        PRIMARY KEY,
    title      text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE book_chapters (
    id         uuid        PRIMARY KEY,
    book       text        NOT NULL REFERENCES books(slug) ON DELETE CASCADE,
    number     int         NOT NULL,
    title      text        NOT NULL,
    page       int         NOT NULL DEFAULT 0,
    words      jsonb       NOT NULL DEFAULT '[]',
    footer     text        NOT NULL DEFAULT '',
    theory     text        NOT NULL DEFAULT '',
    exercises  text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (book, number)
);

INSERT INTO books (slug, title) VALUES ('merfy', 'Murphy Grammar'), ('words', 'Essential Words');

INSERT INTO book_chapters (id, book, number, title, page, words, footer, theory, exercises, created_at, updated_at)
SELECT id, 'merfy', number, title, page, '[]'::jsonb, footer, theory, exercises, created_at, updated_at FROM merfy_units;

INSERT INTO book_chapters (id, book, number, title, page, words, footer, theory, exercises, created_at, updated_at)
SELECT id, 'words', number, title, page, words, footer, theory, exercises, created_at, updated_at FROM word_lessons;

DROP TABLE merfy_units;
DROP TABLE word_lessons;
