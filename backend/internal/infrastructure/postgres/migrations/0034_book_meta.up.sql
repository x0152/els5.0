ALTER TABLE books
    ADD COLUMN series      text NOT NULL DEFAULT '',
    ADD COLUMN level       text NOT NULL DEFAULT '',
    ADD COLUMN description text NOT NULL DEFAULT '';

UPDATE books SET series = slug WHERE series = '';
