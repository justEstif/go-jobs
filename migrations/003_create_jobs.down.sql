DROP INDEX IF EXISTS idx_scrape_runs_started_at;
DROP TABLE IF EXISTS scrape_runs;
DROP TABLE IF EXISTS job_tags;
DROP INDEX IF EXISTS idx_jobs_source;
DROP INDEX IF EXISTS idx_jobs_first_seen;
DROP INDEX IF EXISTS idx_jobs_active;
DROP INDEX IF EXISTS idx_jobs_company_id;
DROP TABLE IF EXISTS jobs;
