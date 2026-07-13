ALTER TABLE raw_events ADD COLUMN client_id text NOT NULL DEFAULT '';

CREATE INDEX raw_events_user_client_idx ON raw_events (user_id, client_id) WHERE client_id <> '';

ALTER TABLE events
    ADD COLUMN client_id text  NOT NULL DEFAULT '',
    ADD COLUMN source    jsonb NOT NULL DEFAULT '{}',
    ADD COLUMN meta      jsonb NOT NULL DEFAULT '{}';
