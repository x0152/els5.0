DROP INDEX IF EXISTS raw_events_user_session_idx;

ALTER TABLE raw_events
    DROP COLUMN version,
    DROP COLUMN type,
    DROP COLUMN note,
    DROP COLUMN session,
    ADD COLUMN skill   text NOT NULL DEFAULT '',
    ADD COLUMN target  text NOT NULL DEFAULT '',
    ADD COLUMN outcome text NOT NULL DEFAULT '';
