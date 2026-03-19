---
# go-jobs-qn3e
title: 'Fix npm package: bin name mismatch and install issues'
status: in-progress
type: bug
priority: normal
created_at: 2026-03-19T18:47:59Z
updated_at: 2026-03-19T18:49:59Z
---

The CLI crashes on startup when no BASE_URL or --base-url is set because it immediately tries to connect to PostgreSQL. Fix: default to https://jobs.estifanos.cc remote mode. Also clean up: Windows install uses unavailable 'unzip', PLATFORM_MAP.files is stale dead code.
