ALTER TABLE users
    ADD COLUMN IF NOT EXISTS resume       TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS llm_model    TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS llm_base_url TEXT NOT NULL DEFAULT '';

-- Cache for LLM analysis results. Avoids re-spending on identical requests.
-- Users can force a refresh ("Re-analyze") which overwrites the cached row.
CREATE TABLE IF NOT EXISTS coach_cache (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id     UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    kind       TEXT NOT NULL,  -- 'analyze' or 'case_study'
    result     TEXT NOT NULL,
    model_used TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, job_id, kind)
);
