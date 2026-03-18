CREATE TABLE IF NOT EXISTS user_jobs (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id     UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    status     TEXT NOT NULL,
    status_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    applied_at TIMESTAMPTZ,
    notes      TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (user_id, job_id)
);

CREATE INDEX IF NOT EXISTS idx_user_jobs_user_status ON user_jobs(user_id, status);
CREATE INDEX IF NOT EXISTS idx_user_jobs_job_id ON user_jobs(job_id);

CREATE TABLE IF NOT EXISTS user_companies (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    hidden     BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (user_id, company_id)
);

CREATE INDEX IF NOT EXISTS idx_user_companies_user_hidden ON user_companies(user_id, hidden);
