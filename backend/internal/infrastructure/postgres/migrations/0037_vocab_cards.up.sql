ALTER TABLE vocab_units ADD COLUMN correct_streak   int NOT NULL DEFAULT 0;
ALTER TABLE vocab_units ADD COLUMN last_answered_at timestamptz;
