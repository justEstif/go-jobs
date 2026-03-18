---
# go-jobs-gpi9
title: Add configurable scrape/enrich interval settings
status: completed
type: task
priority: normal
created_at: 2026-03-18T13:37:43Z
updated_at: 2026-03-18T13:49:08Z
---

Complete M7 operability by making scrape/enrich cadence configurable in serve mode.

## Todo
- [x] Add configuration field(s) for scheduler interval
- [x] Wire config parsing from environment/CLI
- [x] Validate defaults and invalid values
- [x] Surface current interval in startup logs
- [x] Update docs with configuration examples

## Summary of Changes

Added interval and enrichment batch runtime configuration in serve mode, with defaults, validation, and startup visibility.
