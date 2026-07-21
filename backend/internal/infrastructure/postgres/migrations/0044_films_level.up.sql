ALTER TABLE films
    ADD COLUMN level text NOT NULL DEFAULT '';

-- Existing uploads get a sensible default; new uploads must set the level explicitly.
UPDATE films SET level = 'B1' WHERE level = '';
