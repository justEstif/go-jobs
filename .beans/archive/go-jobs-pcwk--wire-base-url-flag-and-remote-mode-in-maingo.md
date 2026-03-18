---
# go-jobs-pcwk
title: Wire --base-url flag and remote mode in main.go
status: completed
type: task
priority: normal
created_at: 2026-03-18T21:36:57Z
updated_at: 2026-03-18T21:47:24Z
parent: go-jobs-qlpf
---

Add --base-url persistent flag to root cobra command. In main.go, detect remote mode and inject httpclient adapters for client-side CLI commands instead of in-process service implementations.

## Behaviour
- serve, scrape, enrich always use local/direct adapters (require DB)
- search, pipeline, tracker, auth commands use httpclient adapters when base-url is set
- When no base-url: current behaviour unchanged

## Summary of Changes

- internal/cli/root.go: added --base-url persistent flag (for help/completion)
- internal/cli/token.go: exported ReadStoredToken() so main.go can pre-read the token
- cmd/jobs/main.go:
  - resolveBaseURL(): pre-scans os.Args for --base-url / --base-url=val, falls back to BASE_URL env
  - firstCommand(): extracts the cobra subcommand name, skipping flags and --base-url value
  - runRemoteCLI(): wires httpclient adapters (no DB boot), calls cobra Execute
  - main(): detects remote mode early; routes to runRemoteCLI or full local stack

Behaviour:
  - serve/scrape/enrich always use local mode (DB required)
  - All other commands use httpclient adapters when --base-url or BASE_URL is set
  - No base URL → existing in-process behaviour unchanged
