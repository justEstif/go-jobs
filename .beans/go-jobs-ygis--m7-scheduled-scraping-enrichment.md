---
# go-jobs-ygis
title: 'M7: Scheduled Scraping + Enrichment'
status: completed
type: milestone
priority: normal
created_at: 2026-03-18T14:55:44Z
updated_at: 2026-03-18T14:59:05Z
---

Implement MVP milestone M7 from docs/MVP.md.\n\n## Scope\n- [x] Add background scheduler that runs scrape + enrich on configurable interval\n- [x] Add CLI command go-jobs serve to start web UI + scheduler\n- [x] Ensure jobs are marked inactive when removed from source (verify behavior remains intact)\n- [x] Run build/tests for verification\n- [x] Add summary of changes

## Summary of Changes\n\n- Added a first-class go-jobs serve Cobra command in internal/cli/serve.go.\n- Wired Serve into CLI dependency injection in internal/cli/root.go and composition root in cmd/jobs/main.go.\n- Refactored HTTP server plus scheduler startup into runHTTPServer so both default no-args mode and go-jobs serve use the same startup and shutdown path.\n- Kept scheduled scrape and enrich behavior configurable via SCRAPE_INTERVAL and ENRICH_LIMIT.\n- Verified existing inactive-job behavior remains in internal/core/services/scrape.go via MarkInactive.\n- Validation: go build ./... and go test ./... both pass.









