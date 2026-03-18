---
# go-jobs-3x4i
title: Update architecture docs for dual driving adapter pattern
status: in-progress
type: task
priority: normal
created_at: 2026-03-18T21:36:57Z
updated_at: 2026-03-18T21:47:35Z
parent: go-jobs-qlpf
---

Update docs/ARCHITECTURE.md and docs/interfaces.md to document the dual driving adapter pattern.

## What to document
- /api/v1/ JSON routes as a second driving adapter (alongside HTML handlers)
- httpclient adapter package and its role
- --base-url flag and BASE_URL env var config
- Command split table: which CLI commands require local access vs. work remotely
- Auth flow for API routes (Bearer token)
