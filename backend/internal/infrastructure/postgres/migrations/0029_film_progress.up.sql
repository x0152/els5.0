CREATE TABLE film_progress (
    owner_id    uuid        NOT NULL,
    film_id     uuid        NOT NULL REFERENCES films (id) ON DELETE CASCADE,
    position_ms int         NOT NULL DEFAULT 0,
    updated_at  timestamptz NOT NULL,
    PRIMARY KEY (owner_id, film_id)
);
