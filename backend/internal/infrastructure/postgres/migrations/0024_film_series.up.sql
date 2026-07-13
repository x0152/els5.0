CREATE TABLE film_series (
    title       text PRIMARY KEY,
    description text NOT NULL DEFAULT '',
    poster_path text NOT NULL DEFAULT ''
);
