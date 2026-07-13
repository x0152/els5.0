ALTER TABLE raw_events DROP CONSTRAINT raw_events_status_check;
ALTER TABLE raw_events ADD CONSTRAINT raw_events_status_check
    CHECK (status IN ('pending', 'processing', 'processed', 'failed'));
ALTER TABLE raw_events ADD COLUMN claimed_at timestamptz;
