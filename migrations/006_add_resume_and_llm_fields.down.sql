DROP TABLE IF EXISTS coach_cache;

ALTER TABLE users
    DROP COLUMN IF EXISTS resume,
    DROP COLUMN IF EXISTS llm_model,
    DROP COLUMN IF EXISTS llm_base_url;
