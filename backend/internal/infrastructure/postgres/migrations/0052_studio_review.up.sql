ALTER TABLE studio_items
    ADD COLUMN recalled       boolean     NOT NULL DEFAULT false,
    ADD COLUMN review_stage   int         NOT NULL DEFAULT 0,
    ADD COLUMN next_review_at timestamptz;
