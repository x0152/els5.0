DROP INDEX vocab_units_account_text_idx;
CREATE UNIQUE INDEX vocab_units_account_text_kind_idx ON vocab_units (account_id, lower(text), kind);
