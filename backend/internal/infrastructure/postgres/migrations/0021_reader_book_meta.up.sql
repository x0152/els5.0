ALTER TABLE reader_books
    ADD COLUMN description text NOT NULL DEFAULT '',
    ADD COLUMN cover_path  text NOT NULL DEFAULT '';
