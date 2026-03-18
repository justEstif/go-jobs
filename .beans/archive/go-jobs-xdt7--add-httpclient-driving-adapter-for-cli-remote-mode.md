---
# go-jobs-xdt7
title: Add httpclient driving adapter for CLI remote mode
status: completed
type: task
priority: normal
created_at: 2026-03-18T21:36:57Z
updated_at: 2026-03-18T21:43:49Z
parent: go-jobs-qlpf
---

Implement internal/adapters/httpclient/ package with HTTP client implementations of the ports the CLI uses. Injected by main.go when --base-url is set or BASE_URL env var is present.

## Ports to implement
- ports.JobSearchService  → GET /api/v1/jobs
- ports.ApplicationService → POST/GET /api/v1/jobs/... and /api/v1/pipeline
- ports.AuthService → POST /api/v1/auth/...
- ports.SessionRepository → not needed in remote mode (token is read from file and passed as Bearer header)

## Config precedence
1. --base-url CLI flag
2. BASE_URL env var
3. Default: empty (use in-process adapters)

## Summary of Changes

Added internal/adapters/httpclient/ package with four files:

- client.go — shared Client struct (baseURL, token, http.Client) with get/post/do helpers and JSON error decoding
- auth.go — AuthClient implements ports.AuthService + ports.SessionRepository; GET /api/v1/auth/me resolves token → User
- search.go — SearchClient implements ports.JobSearchService; maps SearchFilters to query params
- application.go — ApplicationClient implements ports.ApplicationService; routes SetStatus to /interested, /apply, or /status based on value

Also added GET /api/v1/auth/me endpoint to the API (needed by GetUserByToken).
Compile-time interface checks in iface_check_test.go.
