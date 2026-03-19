---
# go-jobs-qn3e
title: 'Fix npm package: bin name mismatch and install issues'
status: completed
type: bug
priority: normal
created_at: 2026-03-19T18:47:59Z
updated_at: 2026-03-19T19:09:20Z
---

The CLI crashes on startup when no BASE_URL or --base-url is set because it immediately tries to connect to PostgreSQL. Fix: default to https://jobs.estifanos.cc remote mode. Also clean up: Windows install uses unavailable 'unzip', PLATFORM_MAP.files is stale dead code.

## Summary of Changes\n\n- Default CLI to remote mode (https://jobs.estifanos.cc) so npm-installed CLI works without PostgreSQL\n- Local mode available via `--base-url local` sentinel\n- Replaced Windows `unzip` with PowerShell extraction in install.js\n- Removed stale `files` dead code from PLATFORM_MAP
