DROP TABLE IF EXISTS reader_collections;
ALTER TABLE reader_books DROP COLUMN IF EXISTS group_title;
ALTER TABLE reader_books DROP COLUMN IF EXISTS kind;
