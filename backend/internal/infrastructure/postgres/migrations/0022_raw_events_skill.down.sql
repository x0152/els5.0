ALTER TABLE raw_events
    DROP COLUMN skill,
    DROP COLUMN target,
    DROP COLUMN outcome,
    ADD COLUMN version int  NOT NULL DEFAULT 1,
    ADD COLUMN type    text NOT NULL DEFAULT '',
    ADD COLUMN note    text NOT NULL DEFAULT '',
    ADD COLUMN session text NOT NULL DEFAULT '';

CREATE INDEX raw_events_user_session_idx ON raw_events (user_id, session) WHERE session <> '';
