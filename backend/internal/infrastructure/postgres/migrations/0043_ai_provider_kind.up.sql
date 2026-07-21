ALTER TABLE ai_providers
    ADD COLUMN kind   text  NOT NULL DEFAULT 'openai',
    ADD COLUMN params jsonb NOT NULL DEFAULT '{}';
