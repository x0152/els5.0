UPDATE raw_events SET status = 'pending' WHERE status = 'processing';
ALTER TABLE raw_events DROP COLUMN claimed_at;
ALTER TABLE raw_events DROP CONSTRAINT raw_events_status_check;
ALTER TABLE raw_events ADD CONSTRAINT raw_events_status_check
    CHECK (status IN ('pending', 'processed', 'failed'));
