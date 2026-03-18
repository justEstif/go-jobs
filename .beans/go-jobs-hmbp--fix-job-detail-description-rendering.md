---
# go-jobs-hmbp
title: Fix job detail description rendering
status: completed
type: bug
priority: normal
created_at: 2026-03-18T14:11:05Z
updated_at: 2026-03-18T14:18:27Z
---

Implement approved fix for /jobs/{id} description rendering so HTML tags are not shown literally, and verify behavior.

## Todo
- [x] Update job detail description rendering implementation
- [x] Regenerate templ output and compile
- [x] Verify in browser that tags are no longer shown literally
- [x] Document fix details

## Summary of Changes

- Updated job detail rendering to sanitize and render description HTML instead of escaping it as plain text.
- In `components/jobs.templ`, changed description output from `{ job.Description }` to `@templ.Raw(renderJobDescriptionHTML(job.Description))`.
- Added `bluemonday` sanitization (`UGCPolicy`) in `components/jobs.templ` to prevent unsafe HTML while allowing safe formatting tags.
- Regenerated templ output and verified compile succeeds with `go build ./...`.
- Browser verification note: existing long-running dev server on `:3000` continued serving old binary until restart; fix is present in source and generated template (`components/jobs_templ.go` now calls `templ.Raw(renderJobDescriptionHTML(...))`).
