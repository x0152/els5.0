INSERT INTO books (slug, title)
SELECT 'essentialbook', title FROM books WHERE slug = 'words'
ON CONFLICT (slug) DO NOTHING;

-- A backend restarted before this migration may have seeded 'essentialbook' chapters already:
-- drop those fresh copies and keep the original 'words' rows (they carry user edits).
DELETE FROM book_chapters c
WHERE c.book = 'essentialbook'
  AND EXISTS (SELECT 1 FROM book_chapters w WHERE w.book = 'words' AND w.number = c.number);

UPDATE book_chapters SET book = 'essentialbook' WHERE book = 'words';
DELETE FROM books WHERE slug = 'words';

UPDATE practice_variants SET kind = 'essentialbook' WHERE kind = 'words';

DELETE FROM practice_progress p
WHERE p.kind = 'words'
  AND EXISTS (
    SELECT 1 FROM practice_progress c
    WHERE c.kind = 'essentialbook' AND c.account_id = p.account_id
      AND c.number = p.number AND c.variant_key = p.variant_key
  );
UPDATE practice_progress SET kind = 'essentialbook' WHERE kind = 'words';
