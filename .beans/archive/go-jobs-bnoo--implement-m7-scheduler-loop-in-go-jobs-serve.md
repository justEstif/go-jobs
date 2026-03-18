---
# go-jobs-bnoo
title: Implement M7 scheduler loop in go-jobs serve
status: completed
type: task
priority: normal
created_at: 2026-03-18T13:37:43Z
updated_at: 2026-03-18T13:48:58Z
---

Build the M7 background runner that periodically executes scrape and enrichment while the web server is running.

## Todo
- [x] Review existing scheduler adapter and serve wiring
- [x] Trigger scrape then enrich on configurable interval
- [x] Ensure graceful shutdown via context cancellation
- [x] Run go test ./... compile verification for scheduler wiring
- [x] Document interval configuration and runtime behavior

## Summary of Changes

Added a serve-mode scheduler that runs scrape then enrich on startup and on every configured interval, with graceful shutdown handling and runtime configuration via environment variables.
