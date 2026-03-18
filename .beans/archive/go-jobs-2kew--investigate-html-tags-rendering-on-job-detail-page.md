---
# go-jobs-2kew
title: Investigate HTML tags rendering on job detail page
status: completed
type: bug
priority: normal
created_at: 2026-03-18T14:05:49Z
updated_at: 2026-03-18T14:09:55Z
---

Review homepage and a job detail page using agent-browser skill; identify why /job/{jobid} may display raw HTML tags.

## Todo
- [x] Reproduce issue in browser on homepage and one job detail page
- [x] Trace rendering path in code to identify root cause
- [x] Document findings and proposed fix approach (no code changes)

## Summary of Changes

- Reproduced the issue in browser: homepage loads correctly, and job detail pages display raw HTML tags in the Description section (for example `<p>`, `<ul>`, and `<li>` are shown as text).
- Traced rendering path: `internal/adapters/http/jobs.go` passes `job.Description` into `components.JobDetailPage`; `components/jobs.templ` renders description using `{ job.Description }`, which templ escapes by default.
- Confirmed output behavior from server response: HTML includes escaped entities like `&lt;p&gt;` (single-escaped), which is why tags appear visibly on `/jobs/{id}`.
