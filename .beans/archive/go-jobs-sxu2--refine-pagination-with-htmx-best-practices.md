---
# go-jobs-sxu2
title: Refine pagination with HTMX best practices
status: completed
type: task
priority: normal
created_at: 2026-03-18T14:03:03Z
updated_at: 2026-03-18T14:03:32Z
---

Apply HTMX best-practice wiring to pagination/search interactions while preserving progressive enhancement.

## Todo
- [x] Add hx-boost + push-url behavior to jobs browse region
- [x] Ensure partial swap target uses stable container id
- [x] Add visible loading indicator for htmx requests
- [x] Verify templates regenerate and build/tests pass
- [x] Add summary of behavior changes

## Summary of Changes

Updated jobs browse pagination/search interactions to use hx-boost with URL history support and container-scoped swaps, plus a lightweight loading indicator for htmx requests.
