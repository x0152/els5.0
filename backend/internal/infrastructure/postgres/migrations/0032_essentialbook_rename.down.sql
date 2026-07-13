INSERT INTO books (slug, title)
SELECT 'words', title FROM books WHERE slug = 'essentialbook'
ON CONFLICT (slug) DO NOTHING;

DELETE FROM book_chapters w
WHERE w.book = 'words'
  AND EXISTS (SELECT 1 FROM book_chapters c WHERE c.book = 'essentialbook' AND c.number = w.number);

UPDATE book_chapters SET book = 'words' WHERE book = 'essentialbook';
DELETE FROM books WHERE slug = 'essentialbook';

UPDATE practice_variants SET kind = 'words' WHERE kind = 'essentialbook';

DELETE FROM practice_progress c
WHERE c.kind = 'essentialbook'
  AND EXISTS (
    SELECT 1 FROM practice_progress w
    WHERE w.kind = 'words' AND w.account_id = c.account_id
      AND w.number = c.number AND w.variant_key = c.variant_key
  );
UPDATE practice_progress SET kind = 'words' WHERE kind = 'essentialbook';
