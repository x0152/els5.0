ALTER TABLE reader_books
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS cover_path;
