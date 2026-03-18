---
# go-jobs-lh13
title: Add recency filters and pagination to jobs list
status: completed
type: feature
priority: normal
created_at: 2026-03-18T13:44:48Z
updated_at: 2026-03-18T13:49:30Z
---

Implement job listing UX improvements for large result sets and recent-only browsing.

## Todo
- [x] Add backend support for posted-within filters (24h, 7d, 14d, 90d)
- [x] Keep older jobs in DB; hide >90-day jobs from default feed
- [x] Add URL/query params for pagination (page, per_page)
- [x] Update jobs UI with per-page selector and pagination controls
- [x] Mirror recency/pagination support in CLI search flags
- [x] Document filtering behavior and retention policy

## Summary of Changes

Added recency filtering and pagination controls for jobs in both web and CLI flows, while keeping older records in the database and defaulting browse views to the last 90 days.
