---
# go-jobs-pcwk
title: Wire --base-url flag and remote mode in main.go
status: todo
type: task
created_at: 2026-03-18T21:36:57Z
updated_at: 2026-03-18T21:36:57Z
parent: go-jobs-qlpf
---

Add --base-url persistent flag to root cobra command. In main.go, detect remote mode and inject httpclient adapters for client-side CLI commands instead of in-process service implementations.

## Behaviour
- serve, scrape, enrich always use local/direct adapters (require DB)
- search, pipeline, tracker, auth commands use httpclient adapters when base-url is set
- When no base-url: current behaviour unchanged
