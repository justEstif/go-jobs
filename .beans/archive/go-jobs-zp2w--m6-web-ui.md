---
# go-jobs-zp2w
title: 'M6: Web UI'
status: completed
type: milestone
priority: normal
created_at: 2026-03-18T11:15:15Z
updated_at: 2026-03-18T12:13:22Z
---

HTTP server serving the full web UI using templ + htmx + DaisyUI.

Browse and search jobs with URL-based filter state. Job cards show 'added X days ago' and highlight jobs added since the user's last visit. Job detail view. Interested/apply/status/notes buttons. Pipeline view. Company hide/show toggle. Scraper status indicator. Auth-aware navbar.

## Tasks

### Design & theme
- [x] Create custom DaisyUI theme and overall web design (color palette, typography, spacing, card style, filter panel layout)

### Shared / layout
- [x] Update Navbar to be auth-aware (Sign in / Sign out links; logout as POST form with CSRF token)
- [x] Add scraper status component (placeholder; full LatestRun wiring deferred)

### Jobs browse & search
- [x] GET / — jobs list page: search bar + filter panel + job cards (replace hello-world home)
- [x] Job card component: title, company, location, tags, "added X days ago", new-since-last-visit highlight
- [x] Filter panel component: role type, seniority, remote policy, country, tech stack — all multi-select; filter state in URL query params
- [x] URL query param binding: parse filters from r.URL.Query() → domain.SearchFilters; render page server-side
- [x] JobSearchHandler struct wired to JobSearchService and ApplicationService
- [x] TouchLastVisited on each authenticated browse request

### Job detail
- [x] GET /jobs/{id} — job detail page: full description, company link, all tags, tracker actions
- [x] Tracker action buttons: Interested / Apply / Status dropdown / Notes (htmx POST, swap the button area)

### Tracker actions (htmx endpoints)
- [x] POST /jobs/{id}/interested — SetStatus(interested); return updated action bar partial
- [x] POST /jobs/{id}/apply — SetStatus(applied); return updated action bar partial
- [x] POST /jobs/{id}/status — SetStatus from form value; return updated action bar partial
- [x] POST /jobs/{id}/notes — SetNotes from form value; return updated notes partial

### Pipeline view
- [x] GET /pipeline — tracked jobs grouped by status (uses ApplicationService.ListPipeline)
- [x] Pipeline page component: columns or grouped list per status with job cards

### Company management
- [x] GET /companies — list all companies with hide/show toggle per company
- [x] POST /companies/{id}/hide — CompanyService.HideCompany; htmx swap toggle button
- [x] POST /companies/{id}/show — CompanyService.ShowCompany; htmx swap toggle button

### Wiring
- [x] Instantiate CompanyService, UserService in main.go
- [x] Add RequireAuth to tracker, pipeline, companies routes
- [x] go build ./... passes clean

## Summary of Changes

### New files
- **components/jobs.templ** — JobsListPage, FilterPanel, JobCard, JobDetailPage, ActionBarPartial, NotesPartial, ScraperStatusPlaceholder + all helpers (daysAgo, filter option lists, status badge classes)
- **components/pipeline.templ** — PipelinePage with pipelineGroup and pipelineJobRow; stable pipeline status order
- **components/companies.templ** — CompaniesPage + CompanyTogglePartial (htmx swap target)
- **internal/adapters/http/jobs.go** — JobSearchHandler (List + Detail); parseSearchFilters from URL query params
- **internal/adapters/http/tracker.go** — TrackerHandler (Interested, Apply, SetStatus, SetNotes); each returns htmx partial
- **internal/adapters/http/pipeline.go** — PipelineHandler.List
- **internal/adapters/http/companies.go** — CompanyHandler (List, Hide, Show)
- **internal/core/services/company.go** — CompanyService impl; ListCompanies (filters hidden), ListAllWithHiddenIDs (for /companies page)
- **internal/core/services/user.go** — UserService impl; SetLLMKey, GetByID, TouchLastVisited

### Modified files
- **components/layout.templ** — Layout now takes (loggedIn bool, csrfToken string) to pass to Navbar
- **components/navbar.templ** — Auth-aware: shows Sign in/Register vs Jobs/Pipeline/Companies/Sign out; logout is POST form
- **components/auth.templ** — RegisterPage/LoginPage updated signatures; includes CSRF token in form
- **components/about.templ** — Updated to new Layout signature
- **components/contact.templ** — Updated to new Layout signature
- **components/home.templ** — Deleted (/ now served by JobSearchHandler)
- **internal/adapters/http/auth.go** — Updated calls to RegisterPage/LoginPage with csrfToken
- **internal/adapters/http/about.go** — Passes loggedIn + csrfToken
- **internal/adapters/http/contact.go** — Passes loggedIn + csrfToken
- **internal/core/ports/driven.go** — Added TouchLastVisited to UserRepository interface
- **internal/core/ports/driving.go** — Added ListAllWithHiddenIDs to CompanyService interface
- **cmd/jobs/main.go** — Wired UserService, CompanyService, UserCompanyRepo; all new HTTP handlers; RequireAuth group for tracker/pipeline/companies routes
