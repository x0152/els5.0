ALTER TABLE vocab_units ADD COLUMN frequency int  NOT NULL DEFAULT 3;
ALTER TABLE vocab_units ADD COLUMN cefr      text NOT NULL DEFAULT 'B1';
