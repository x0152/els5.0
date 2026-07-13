ALTER TABLE quest_missions DROP CONSTRAINT quest_missions_origin_fk;

DELETE FROM quest_missions WHERE user_id <> author_id;

ALTER TABLE quest_missions DROP CONSTRAINT quest_missions_pkey;
ALTER TABLE quest_missions ADD PRIMARY KEY (id);

ALTER TABLE quest_missions DROP COLUMN author_id;
