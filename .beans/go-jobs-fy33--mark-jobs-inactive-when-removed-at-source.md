---
# go-jobs-fy33
title: Mark jobs inactive when removed at source
status: completed
type: task
priority: normal
created_at: 2026-03-18T13:37:43Z
updated_at: 2026-03-18T14:43:14Z
---

Implement M7 inactive job handling so jobs that disappear from providers are marked inactive without deleting history.

## Todo
- [x] Define repository/service contract for deactivation
- [x] Update scrape flow to compute active vs missing jobs
- [x] Persist inactive state and preserve prior tracker data
- [x] Add tests for disappear/reappear scenarios
- [x] Verify search excludes inactive jobs by default

## Summary of Changes

- Updated the JobRepository contract so MarkInactive returns the number of jobs deactivated.
- Updated scrape flow to collect external IDs from each batch, call MarkInactive, and count removed jobs in scrape run metrics.
- Updated the Postgres adapter and sqlc query to use execrows for deactivation counts while preserving job history by toggling active=false.
- Added scrape service unit tests for disappear and reappear scenarios, including reactivation via upsert.
- Verified inactive jobs are excluded by default search via the SearchJobs SQL predicate (WHERE j.active = TRUE).
