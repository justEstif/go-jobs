---
# go-jobs-hfpq
title: Add remote API endpoint mode for CLI
status: todo
type: feature
priority: normal
created_at: 2026-03-18T18:56:22Z
updated_at: 2026-03-18T18:57:03Z
parent: go-jobs-qlam
---

Introduce a networked CLI mode so users can target a hosted go-jobs server instead of requiring direct DB access.

## Goal
Make CLI usable as a standalone client against an HTTP API endpoint.

## Proposed behavior
- Dev default: read base URL from existing BASE_URL env var (from mise), defaulting to http://127.0.0.1:3000 when unset.
- Support CLI override flag (e.g. --base-url) that takes precedence over env.
- Keep existing local/in-process behavior available behind explicit mode if needed during migration.

## Todo
- [ ] Define CLI connection config contract (env var names, precedence, defaults).
- [ ] Add HTTP client adapter for CLI commands used by search/auth/tracker flows.
- [ ] Add endpoint/base-url flag wiring in root command.
- [ ] Update docs with local dev and hosted usage examples.
- [ ] Add tests for config precedence and request routing.
