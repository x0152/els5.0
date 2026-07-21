CREATE TABLE workout_plans (
    film_id    uuid        PRIMARY KEY REFERENCES films(id) ON DELETE CASCADE,
    status     text        NOT NULL,
    error      text        NOT NULL DEFAULT '',
    segments   jsonb       NOT NULL DEFAULT '[]',
    created_at timestamptz NOT NULL
);

CREATE TABLE workout_lessons (
    id           uuid        PRIMARY KEY,
    account_id   uuid        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    number       int         NOT NULL,
    film_id      uuid        REFERENCES films(id) ON DELETE SET NULL,
    start_ms     int         NOT NULL DEFAULT 0,
    end_ms       int         NOT NULL DEFAULT 0,
    status       text        NOT NULL,
    steps        jsonb       NOT NULL DEFAULT '[]',
    created_at   timestamptz NOT NULL,
    completed_at timestamptz,
    UNIQUE (account_id, number)
);

CREATE TABLE workout_items (
    id             uuid        PRIMARY KEY,
    account_id     uuid        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    kind           text        NOT NULL,
    text           text        NOT NULL,
    film_id        text        NOT NULL DEFAULT '',
    start_ms       int         NOT NULL DEFAULT 0,
    end_ms         int         NOT NULL DEFAULT 0,
    lesson_number  int         NOT NULL,
    last_score     int         NOT NULL DEFAULT 0,
    times_reviewed int         NOT NULL DEFAULT 0,
    last_lesson    int         NOT NULL DEFAULT 0,
    updated_at     timestamptz NOT NULL,
    UNIQUE (account_id, kind, text)
);

CREATE TABLE workout_positions (
    account_id   uuid        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    title        text        NOT NULL,
    film_id      uuid        NOT NULL,
    next_segment int         NOT NULL DEFAULT 0,
    used_at      timestamptz NOT NULL,
    PRIMARY KEY (account_id, title)
);
