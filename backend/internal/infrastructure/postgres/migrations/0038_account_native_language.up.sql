ALTER TABLE accounts ADD COLUMN native_language   text    NOT NULL DEFAULT 'Russian';
ALTER TABLE accounts ADD COLUMN show_translations boolean NOT NULL DEFAULT true;
