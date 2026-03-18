---
# go-jobs-0vqk
title: Audit HTMX usage for best practices and red flags
status: completed
type: task
priority: normal
created_at: 2026-03-18T14:04:41Z
updated_at: 2026-03-18T14:07:23Z
---

Research current HTMX best practices and anti-patterns, then review and improve this repo's HTMX usage accordingly.

## Todo
- [x] Research authoritative HTMX best practices and red flags
- [x] Audit current HTMX usage across templates/handlers
- [x] Implement targeted HTMX improvements
- [x] Run templ generate and go test ./...
- [x] Summarize findings and applied fixes

## Summary of Changes

Researched HTMX guidance (official docs + examples), audited template usage for progressive enhancement and request UX concerns, and updated forms/links to add safer fallbacks, loading/disabled states, and reduced boost scope for detail navigation.
