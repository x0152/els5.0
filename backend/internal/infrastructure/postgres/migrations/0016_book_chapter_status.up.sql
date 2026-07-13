ALTER TABLE book_chapters ADD COLUMN status text NOT NULL DEFAULT 'ready';
ALTER TABLE book_chapters ADD COLUMN error  text NOT NULL DEFAULT '';
