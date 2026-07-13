ALTER TABLE films ADD COLUMN kind         text NOT NULL DEFAULT 'film';
ALTER TABLE films ADD COLUMN series_title text NOT NULL DEFAULT '';
ALTER TABLE films ADD COLUMN season       int  NOT NULL DEFAULT 0;
ALTER TABLE films ADD COLUMN episode      int  NOT NULL DEFAULT 0;
