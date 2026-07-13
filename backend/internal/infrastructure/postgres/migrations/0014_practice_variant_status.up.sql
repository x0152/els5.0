ALTER TABLE practice_variants ADD COLUMN status text NOT NULL DEFAULT 'ready';
ALTER TABLE practice_variants ADD COLUMN error  text NOT NULL DEFAULT '';
