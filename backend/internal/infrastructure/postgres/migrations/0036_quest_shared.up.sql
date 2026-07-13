ALTER TABLE quest_missions ADD COLUMN author_id text;
UPDATE quest_missions SET author_id = user_id;
ALTER TABLE quest_missions ALTER COLUMN author_id SET NOT NULL;

ALTER TABLE quest_missions DROP CONSTRAINT quest_missions_pkey;
ALTER TABLE quest_missions ADD PRIMARY KEY (id, user_id);

-- Forks reference the author row: deleting the original cascades to player copies.
ALTER TABLE quest_missions ADD CONSTRAINT quest_missions_origin_fk
    FOREIGN KEY (id, author_id) REFERENCES quest_missions (id, user_id) ON DELETE CASCADE;
