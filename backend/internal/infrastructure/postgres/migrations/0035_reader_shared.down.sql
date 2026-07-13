ALTER TABLE reader_collections DROP CONSTRAINT reader_collections_pkey;
ALTER TABLE reader_collections ADD COLUMN owner_id text NOT NULL DEFAULT '';
ALTER TABLE reader_collections ADD PRIMARY KEY (owner_id, title);

ALTER TABLE reader_books ADD COLUMN position integer NOT NULL DEFAULT 0;

UPDATE reader_books b
SET position = p.position
FROM reader_progress p
WHERE p.book_id = b.id AND p.owner_id = b.owner_id;

DROP TABLE reader_progress;
