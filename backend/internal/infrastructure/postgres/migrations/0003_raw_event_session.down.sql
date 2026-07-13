DROP INDEX IF EXISTS raw_events_user_session_idx;

ALTER TABLE raw_events DROP COLUMN IF EXISTS session;
