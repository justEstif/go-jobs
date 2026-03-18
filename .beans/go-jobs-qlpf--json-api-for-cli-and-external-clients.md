---
# go-jobs-qlpf
title: JSON API for CLI and external clients
status: todo
type: epic
created_at: 2026-03-18T21:36:40Z
updated_at: 2026-03-18T21:36:40Z
---

Add a /api/v1/ JSON route tree as a second driving adapter alongside the existing HTML web UI routes. Both share the same core service layer — no logic duplication. Enables the CLI to work against a remote server, and opens the door for third-party clients.

## Outcomes
- All CLI-accessible operations have corresponding JSON API endpoints under /api/v1/
- CLI can target a remote go-jobs server via --base-url flag or BASE_URL env var
- HTTP client adapter package implements the relevant ports for remote CLI mode
- Architecture docs updated to reflect dual driving adapter pattern
