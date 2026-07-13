-- 0032 briefly renamed the book to 'corebook'; the final slug is 'essentialbook'.
INSERT INTO books (slug, title)
SELECT 'essentialbook', title FROM books WHERE slug = 'corebook'
ON CONFLICT (slug) DO NOTHING;

DELETE FROM book_chapters e
WHERE e.book = 'essentialbook'
  AND EXISTS (SELECT 1 FROM book_chapters c WHERE c.book = 'corebook' AND c.number = e.number);

UPDATE book_chapters SET book = 'essentialbook' WHERE book = 'corebook';
DELETE FROM books WHERE slug = 'corebook';

UPDATE practice_variants SET kind = 'essentialbook' WHERE kind = 'corebook';

DELETE FROM practice_progress p
WHERE p.kind = 'corebook'
  AND EXISTS (
    SELECT 1 FROM practice_progress e
    WHERE e.kind = 'essentialbook' AND e.account_id = p.account_id
      AND e.number = p.number AND e.variant_key = p.variant_key
  );
UPDATE practice_progress SET kind = 'essentialbook' WHERE kind = 'corebook';
