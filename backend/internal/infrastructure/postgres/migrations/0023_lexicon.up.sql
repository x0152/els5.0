CREATE TABLE base_forms (
    text     varchar(255) PRIMARY KEY,
    pos      varchar(50),
    is_stop  boolean     NOT NULL DEFAULT false,
    language varchar(10) NOT NULL DEFAULT 'en',
    meaning  text
);

CREATE TABLE media_segments (
    id          uuid        PRIMARY KEY,
    media_id    uuid        NOT NULL,
    kind        varchar(20) NOT NULL,
    segment_idx integer     NOT NULL,
    start_pos   integer     NOT NULL DEFAULT 0,
    end_pos     integer     NOT NULL DEFAULT 0,
    text        text        NOT NULL DEFAULT '',
    metadata    jsonb       NOT NULL DEFAULT '{}'
);

CREATE INDEX media_segments_media_kind_idx ON media_segments (media_id, kind, segment_idx);

CREATE TABLE units (
    id            uuid         PRIMARY KEY,
    media_id      uuid         NOT NULL,
    segment_id    uuid         REFERENCES media_segments(id) ON DELETE SET NULL,
    unit_type     varchar(50)  NOT NULL,
    base_form     varchar(255) NOT NULL REFERENCES base_forms(text),
    pos           varchar(50),
    sentence_idx  integer      NOT NULL,
    unit_metadata jsonb        NOT NULL DEFAULT '{}',
    language      varchar(10)  NOT NULL DEFAULT 'en'
);

CREATE INDEX units_media_id_idx ON units (media_id);
CREATE INDEX units_segment_id_idx ON units (segment_id);
CREATE INDEX units_base_form_idx ON units (base_form);

CREATE TABLE unit_spans (
    id        uuid         PRIMARY KEY,
    unit_id   uuid         NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    position  integer      NOT NULL,
    span_type varchar(10)  NOT NULL,
    start     integer      NOT NULL,
    "end"     integer      NOT NULL,
    text      varchar(255) NOT NULL
);

CREATE INDEX unit_spans_unit_id_idx ON unit_spans (unit_id);
CREATE INDEX unit_spans_start_end_idx ON unit_spans (start, "end");
