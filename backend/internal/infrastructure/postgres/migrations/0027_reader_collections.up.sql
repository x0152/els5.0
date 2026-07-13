ALTER TABLE reader_books ADD COLUMN kind        text NOT NULL DEFAULT 'book';
ALTER TABLE reader_books ADD COLUMN group_title text NOT NULL DEFAULT '';

CREATE TABLE reader_collections (
    owner_id    text NOT NULL,
    title       text NOT NULL,
    description text NOT NULL DEFAULT '',
    cover_path  text NOT NULL DEFAULT '',
    PRIMARY KEY (owner_id, title)
);
