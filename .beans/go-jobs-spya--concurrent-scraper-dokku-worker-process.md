---
# go-jobs-spya
title: Concurrent scraper + Dokku worker process
status: completed
type: task
priority: normal
created_at: 2026-03-19T14:48:27Z
updated_at: 2026-03-19T14:49:35Z
---

1. Add concurrent worker pool to ScrapeService.Run (golang.org/x/sync/errgroup + semaphore, ~20 workers)
2. Add --loop flag to `jobs scrape` CLI command for running as a long-lived daemon
3. Create Procfile for Dokku with web and worker entries

## Summary of Changes

- ****: replaced sequential company loop with a concurrent worker pool using  counters +  (20 workers). Companies are scraped in parallel; DB upserts per-company are still sequential within each goroutine to avoid interleaving.
- ****: added  and  flags. With , the command runs as a long-lived daemon re-scraping on the interval, with graceful SIGTERM shutdown via .
- ****: created with  (serve) and  (scrape --loop) process types for Dokku.
