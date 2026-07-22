ALTER TABLE studio_items
    ADD COLUMN explanation        text NOT NULL DEFAULT '',
    ADD COLUMN explanation_native text NOT NULL DEFAULT '';
