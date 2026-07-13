CREATE TABLE reader_progress (
    owner_id text    NOT NULL,
    book_id  uuid    NOT NULL REFERENCES reader_books (id) ON DELETE CASCADE,
    position integer NOT NULL DEFAULT 0,
    PRIMARY KEY (owner_id, book_id)
);

INSERT INTO reader_progress (owner_id, book_id, position)
SELECT owner_id, id, position FROM reader_books WHERE position > 0;

ALTER TABLE reader_books DROP COLUMN position;

DELETE FROM reader_collections a USING reader_collections b
WHERE a.title = b.title AND a.owner_id > b.owner_id;

ALTER TABLE reader_collections DROP CONSTRAINT reader_collections_pkey;
ALTER TABLE reader_collections DROP COLUMN owner_id;
ALTER TABLE reader_collections ADD PRIMARY KEY (title);
