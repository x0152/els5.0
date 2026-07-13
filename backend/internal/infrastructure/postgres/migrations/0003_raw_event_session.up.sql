ALTER TABLE raw_events ADD COLUMN session text NOT NULL DEFAULT '';

CREATE INDEX raw_events_user_session_idx ON raw_events (user_id, session) WHERE session <> '';
