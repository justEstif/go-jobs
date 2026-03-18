---
# go-jobs-fy33
title: Mark jobs inactive when removed at source
status: todo
type: task
created_at: 2026-03-18T13:37:43Z
updated_at: 2026-03-18T13:37:43Z
---

Implement M7 inactive job handling so jobs that disappear from providers are marked inactive without deleting history.

## Todo
- [ ] Define repository/service contract for deactivation
- [ ] Update scrape flow to compute active vs missing jobs
- [ ] Persist inactive state and preserve prior tracker data
- [ ] Add tests for disappear/reappear scenarios
- [ ] Verify search excludes inactive jobs by default
