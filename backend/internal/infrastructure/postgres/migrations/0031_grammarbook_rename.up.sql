INSERT INTO books (slug, title)
SELECT 'grammarbook', title FROM books WHERE slug = 'merfy'
ON CONFLICT (slug) DO NOTHING;

-- A backend restarted before this migration may have seeded 'grammarbook' chapters already:
-- drop those fresh copies and keep the original 'merfy' rows (they carry user edits).
DELETE FROM book_chapters g
WHERE g.book = 'grammarbook'
  AND EXISTS (SELECT 1 FROM book_chapters m WHERE m.book = 'merfy' AND m.number = g.number);

UPDATE book_chapters SET book = 'grammarbook' WHERE book = 'merfy';
DELETE FROM books WHERE slug = 'merfy';

UPDATE practice_variants SET kind = 'grammarbook' WHERE kind = 'merfy';

DELETE FROM practice_progress p
WHERE p.kind = 'merfy'
  AND EXISTS (
    SELECT 1 FROM practice_progress g
    WHERE g.kind = 'grammarbook' AND g.account_id = p.account_id
      AND g.number = p.number AND g.variant_key = p.variant_key
  );
UPDATE practice_progress SET kind = 'grammarbook' WHERE kind = 'merfy';

-- Empty leftovers from interim book names that never shipped.
DELETE FROM books b
WHERE b.slug IN ('vocabulary', 'collocations')
  AND NOT EXISTS (SELECT 1 FROM book_chapters c WHERE c.book = b.slug);

