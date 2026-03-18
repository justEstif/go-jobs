---
# go-jobs-zp2w
title: 'M6: Web UI'
status: todo
type: milestone
priority: normal
created_at: 2026-03-18T11:15:15Z
updated_at: 2026-03-18T11:17:03Z
---

HTTP server serving the full web UI using templ + htmx + DaisyUI.

Browse and search jobs with URL-based filter state. Job cards show 'added X days ago' and highlight jobs added since the user's last visit. Job detail view. Interested/apply/status/notes buttons. Pipeline view. Company hide/show toggle. Scraper status indicator. Auth-aware navbar.

## Tasks

### Design & theme
- [ ] Create custom DaisyUI theme and overall web design (color palette, typography, spacing, card style, filter panel layout)

### Shared / layout
- [ ] Update Navbar to be auth-aware (Sign in / Sign out links; logout as POST form with CSRF token)
- [ ] Add scraper status component ("last updated X ago" from ScrapeService.LatestRun)

### Jobs browse & search
- [ ] GET / — jobs list page: search bar + filter panel + job cards (replace hello-world home)
- [ ] Job card component: title, company, location, tags, "added X days ago", new-since-last-visit highlight
- [ ] Filter panel component: role type, seniority, remote policy, country, tech stack — all multi-select; filter state in URL query params
- [ ] URL query param binding: parse filters from r.URL.Query() → domain.SearchFilters; render page server-side
- [ ] JobSearchHandler struct wired to JobSearchService and ApplicationService
- [ ] TouchLastVisited on each authenticated browse request

### Job detail
- [ ] GET /jobs/{id} — job detail page: full description, company link, all tags, tracker actions
- [ ] Tracker action buttons: Interested / Apply / Status dropdown / Notes (htmx POST, swap the button area)

### Tracker actions (htmx endpoints)
- [ ] POST /jobs/{id}/interested — SetStatus(interested); return updated action bar partial
- [ ] POST /jobs/{id}/apply — SetStatus(applied); return updated action bar partial
- [ ] POST /jobs/{id}/status — SetStatus from form value; return updated action bar partial
- [ ] POST /jobs/{id}/notes — SetNotes from form value; return updated notes partial

### Pipeline view
- [ ] GET /pipeline — tracked jobs grouped by status (uses ApplicationService.ListPipeline)
- [ ] Pipeline page component: columns or grouped list per status with job cards

### Company management
- [ ] GET /companies — list all companies with hide/show toggle per company
- [ ] POST /companies/{id}/hide — CompanyService.HideCompany; htmx swap toggle button
- [ ] POST /companies/{id}/show — CompanyService.ShowCompany; htmx swap toggle button

### Wiring
- [ ] Instantiate CompanyService, UserService in main.go
- [ ] Add RequireAuth to tracker, pipeline, companies routes
- [ ] go build ./... passes clean
