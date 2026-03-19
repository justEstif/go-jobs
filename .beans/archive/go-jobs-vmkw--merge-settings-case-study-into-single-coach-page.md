---
# go-jobs-vmkw
title: Merge settings + case study into single /coach page
status: completed
type: task
priority: normal
created_at: 2026-03-19T17:06:30Z
updated_at: 2026-03-19T17:08:53Z
---

Consolidate resume config, LLM provider config, and case study form/result into a single /coach page. Remove /settings and /case-study routes. Update navbar, handlers, templates, and routes.

## Tasks
- [x] Create unified CoachPage templ template with three sections (resume, LLM, case study)
- [x] Merge SettingsHandler into CoachHandler (add user service dependency, resume/LLM save methods)
- [x] Update routes: /coach GET, /coach/resume POST, /coach/llm POST, /coach/case-study POST
- [x] Remove /settings and /case-study routes
- [x] Update navbar: replace Settings + Case Study links with single Coach link
- [x] Delete settings.go handler (merged into coach.go)
- [x] Update composition root (main.go) wiring
- [x] Run templ generate + go build

## Summary of Changes

Merged /settings and /case-study into a single /coach page.

### Routes
- `GET /coach` — unified page with resume, LLM config, and case study generator
- `POST /coach/resume` — save resume
- `POST /coach/llm` — save LLM provider config
- `POST /coach/case-study` — generate case study
- Removed: `GET /settings`, `POST /settings/resume`, `POST /settings/llm`, `GET /case-study`, `POST /case-study`

### Files changed
- `components/settings.templ` — replaced SettingsPage + CaseStudyPage with unified CoachPage
- `components/navbar.templ` — replaced Settings + Case Study links with single Coach link
- `components/jobs.templ` — updated /settings references to /coach
- `internal/adapters/http/coach.go` — merged SettingsHandler methods (Show, SaveResume, SaveLLM) into CoachHandler; now takes both coach + user service
- `cmd/jobs/main.go` — updated wiring and routes

### Files deleted
- `internal/adapters/http/settings.go` — fully merged into coach.go
