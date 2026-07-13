DROP INDEX IF EXISTS raw_events_user_client_idx;

ALTER TABLE raw_events DROP COLUMN IF EXISTS client_id;

ALTER TABLE events
    DROP COLUMN IF EXISTS client_id,
    DROP COLUMN IF EXISTS source,
    DROP COLUMN IF EXISTS meta;
