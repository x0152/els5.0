INSERT INTO books (slug, title)
SELECT 'merfy', title FROM books WHERE slug = 'grammarbook'
ON CONFLICT (slug) DO NOTHING;

DELETE FROM book_chapters m
WHERE m.book = 'merfy'
  AND EXISTS (SELECT 1 FROM book_chapters g WHERE g.book = 'grammarbook' AND g.number = m.number);

UPDATE book_chapters SET book = 'merfy' WHERE book = 'grammarbook';
DELETE FROM books WHERE slug = 'grammarbook';

UPDATE practice_variants SET kind = 'merfy' WHERE kind = 'grammarbook';

DELETE FROM practice_progress g
WHERE g.kind = 'grammarbook'
  AND EXISTS (
    SELECT 1 FROM practice_progress m
    WHERE m.kind = 'merfy' AND m.account_id = g.account_id
      AND m.number = g.number AND m.variant_key = g.variant_key
  );
UPDATE practice_progress SET kind = 'merfy' WHERE kind = 'grammarbook';
