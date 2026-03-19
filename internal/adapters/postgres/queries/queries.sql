-- ============================================================
-- Users
-- ============================================================

-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: UpdateUser :exec
UPDATE users
SET
    llm_api_key     = $2,
    llm_provider    = $3,
    llm_model       = $4,
    llm_base_url    = $5,
    resume          = $6,
    last_visited_at = $7
WHERE id = $1;

-- name: TouchUserLastVisited :exec
UPDATE users
SET last_visited_at = NOW()
WHERE id = $1;

-- name: UpdatePassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: DeleteUserJobs :exec
DELETE FROM user_jobs WHERE user_id = $1;

-- ============================================================
-- Sessions
-- ============================================================

-- name: SaveSession :exec
INSERT INTO sessions (token, user_id)
VALUES ($1, $2)
ON CONFLICT (token) DO UPDATE SET user_id = EXCLUDED.user_id;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = $1;

-- name: GetUserByToken :one
SELECT u.*
FROM users u
JOIN sessions s ON s.user_id = u.id
WHERE s.token = $1
LIMIT 1;

-- ============================================================
-- Companies
-- ============================================================

-- name: UpsertCompany :one
INSERT INTO companies (name, careers_url, ats_type, scrape_type, board_token, active)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (ats_type, board_token) DO UPDATE
    SET name        = EXCLUDED.name,
        careers_url = EXCLUDED.careers_url,
        scrape_type = EXCLUDED.scrape_type,
        active      = EXCLUDED.active
RETURNING *;

-- name: ListActiveCompanies :many
SELECT * FROM companies
WHERE active = TRUE
ORDER BY name;

-- name: GetCompanyByID :one
SELECT * FROM companies
WHERE id = $1
LIMIT 1;

-- name: GetCompanyByBoardToken :one
SELECT * FROM companies
WHERE ats_type = $1 AND board_token = $2
LIMIT 1;

-- ============================================================
-- Jobs
-- ============================================================

-- name: UpsertJob :one
INSERT INTO jobs (company_id, external_id, title, url, location, description, raw_html, source, first_seen, last_seen)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (company_id, external_id) DO UPDATE
    SET title       = EXCLUDED.title,
        url         = EXCLUDED.url,
        location    = EXCLUDED.location,
        description = EXCLUDED.description,
        raw_html    = EXCLUDED.raw_html,
        last_seen   = EXCLUDED.last_seen,
        active      = TRUE
RETURNING *;

-- name: GetJobByID :one
SELECT j.*, c.name AS company_name
FROM jobs j
JOIN companies c ON c.id = j.company_id
WHERE j.id = $1
LIMIT 1;

-- name: GetJobsByIDs :many
SELECT j.*, c.name AS company_name
FROM jobs j
JOIN companies c ON c.id = j.company_id
WHERE j.id = ANY($1::uuid[]);

-- name: MarkJobsInactive :execrows
UPDATE jobs
SET active = FALSE
WHERE company_id = $1
  AND external_id != ALL($2::text[])
  AND active = TRUE;

-- name: ListUnenrichedJobs :many
SELECT j.*, c.name AS company_name
FROM jobs j
JOIN companies c ON c.id = j.company_id
LEFT JOIN job_tags jt ON jt.job_id = j.id
WHERE jt.job_id IS NULL
  AND j.active = TRUE
ORDER BY j.first_seen ASC
LIMIT $1;

-- name: SearchJobs :many
-- Filters jobs using multi-dimensional criteria. All slice parameters use OR
-- semantics within the field; tech_stack uses AND semantics (job must mention
-- all specified terms). Passing an empty slice disables that filter.
--
-- Parameters:
--   $1  query        TEXT          — free-text match on title or company name ('' disables)
--   $2  role_types   TEXT[]        — OR filter on job_tags.role_type
--   $3  seniorities  TEXT[]        — OR filter on job_tags.seniority
--   $4  remote_policies TEXT[]     — OR filter on job_tags.remote_policy
--   $5  countries    TEXT[]        — OR filter on job_tags.country
--   $6  tech_stack   TEXT[]        — AND filter: job_tags.tech_stack must contain all items
--   $7  company_ids  UUID[]        — OR filter on jobs.company_id
--   $8  posted_within_days INT     — only jobs first_seen in last N days (<=0 disables)
--   $9  limit        INT
--   $10 offset       INT
SELECT
    j.id,
    j.company_id,
    c.name AS company_name,
    j.external_id,
    j.title,
    j.url,
    j.location,
    j.description,
    j.raw_html,
    j.source,
    j.first_seen,
    j.last_seen,
    j.active,
    jt.role_type,
    jt.seniority,
    jt.remote_policy,
    jt.location_norm,
    jt.country,
    jt.tech_stack,
    jt.enrichment_source,
    jt.enriched_at
FROM jobs j
JOIN companies c ON c.id = j.company_id
LEFT JOIN job_tags jt ON jt.job_id = j.id
WHERE j.active = TRUE
  AND (
    $1 = ''
    OR j.title ILIKE '%' || $1 || '%'
    OR c.name  ILIKE '%' || $1 || '%'
  )
  AND (
    array_length($2::text[], 1) IS NULL
    OR jt.role_type = ANY($2::text[])
  )
  AND (
    array_length($3::text[], 1) IS NULL
    OR jt.seniority = ANY($3::text[])
  )
  AND (
    array_length($4::text[], 1) IS NULL
    OR jt.remote_policy = ANY($4::text[])
  )
  AND (
    array_length($5::text[], 1) IS NULL
    OR jt.country = ANY($5::text[])
  )
  AND (
    array_length($6::text[], 1) IS NULL
    OR $6::text[] <@ jt.tech_stack
  )
  AND (
    array_length($7::uuid[], 1) IS NULL
    OR j.company_id = ANY($7::uuid[])
  )
  AND (
    $8 <= 0
    OR j.first_seen >= NOW() - ($8 * INTERVAL '1 day')
  )
ORDER BY j.first_seen DESC
LIMIT $9
OFFSET $10;

-- ============================================================
-- Job Tags
-- ============================================================

-- name: UpsertJobTags :exec
INSERT INTO job_tags (job_id, role_type, seniority, remote_policy, location_norm, country, tech_stack, enrichment_source, enriched_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (job_id) DO UPDATE
    SET role_type         = EXCLUDED.role_type,
        seniority         = EXCLUDED.seniority,
        remote_policy     = EXCLUDED.remote_policy,
        location_norm     = EXCLUDED.location_norm,
        country           = EXCLUDED.country,
        tech_stack        = EXCLUDED.tech_stack,
        enrichment_source = EXCLUDED.enrichment_source,
        enriched_at       = EXCLUDED.enriched_at;

-- ============================================================
-- Scrape Runs
-- ============================================================

-- name: CreateScrapeRun :exec
INSERT INTO scrape_runs (id, started_at, status)
VALUES ($1, $2, $3);

-- name: UpdateScrapeRun :exec
UPDATE scrape_runs
SET
    finished_at  = $2,
    status       = $3,
    jobs_added   = $4,
    jobs_updated = $5,
    jobs_removed = $6,
    error        = $7
WHERE id = $1;

-- name: GetLatestScrapeRun :one
SELECT * FROM scrape_runs
ORDER BY started_at DESC
LIMIT 1;

-- ============================================================
-- User Jobs
-- ============================================================

-- name: UpsertUserJob :exec
INSERT INTO user_jobs (user_id, job_id, status, status_at, applied_at, notes)
VALUES ($1, $2, $3, $4,
    CASE WHEN $3 = 'applied' THEN COALESCE($5, NOW()) ELSE $5 END,
    $6)
ON CONFLICT (user_id, job_id) DO UPDATE
    SET status     = EXCLUDED.status,
        status_at  = EXCLUDED.status_at,
        applied_at = CASE
            WHEN user_jobs.applied_at IS NOT NULL THEN user_jobs.applied_at
            WHEN EXCLUDED.status = 'applied' THEN COALESCE(EXCLUDED.applied_at, NOW())
            ELSE NULL
        END,
        notes      = EXCLUDED.notes;

-- name: GetUserJob :one
SELECT * FROM user_jobs
WHERE user_id = $1 AND job_id = $2
LIMIT 1;

-- name: ListUserJobsByStatus :many
SELECT job_id FROM user_jobs
WHERE user_id = $1 AND status = $2
ORDER BY status_at DESC;

-- name: ListAllUserJobs :many
SELECT * FROM user_jobs
WHERE user_id = $1
ORDER BY status_at DESC;

-- ============================================================
-- User Companies
-- ============================================================

-- name: SetUserCompanyHidden :exec
INSERT INTO user_companies (user_id, company_id, hidden)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, company_id) DO UPDATE
    SET hidden = EXCLUDED.hidden;

-- name: ListHiddenCompanies :many
SELECT company_id FROM user_companies
WHERE user_id = $1 AND hidden = TRUE;

-- name: IsCompanyHidden :one
SELECT EXISTS(
    SELECT 1 FROM user_companies
    WHERE user_id = $1 AND company_id = $2 AND hidden = TRUE
) AS hidden;

-- ============================================================
-- Coach Cache
-- ============================================================

-- name: GetCoachCache :one
SELECT * FROM coach_cache
WHERE user_id = $1 AND job_id = $2 AND kind = $3
LIMIT 1;

-- name: UpsertCoachCache :exec
INSERT INTO coach_cache (user_id, job_id, kind, result, model_used)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, job_id, kind) DO UPDATE
    SET result     = EXCLUDED.result,
        model_used = EXCLUDED.model_used,
        created_at = NOW();
