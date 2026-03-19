---
# go-jobs-ril3
title: 'Redesign companies: hide from job detail + blocked list in settings'
status: completed
type: feature
priority: normal
created_at: 2026-03-19T17:16:04Z
updated_at: 2026-03-19T17:31:04Z
---

Rework the companies UX:

1. Add 'Hide company' button on job detail page — primary interaction point
2. Replace /companies page with a blocked-companies section on a new /settings page  
3. /settings page: blocked companies (with search + unblock), plus placeholders for password change, tracker reset, account deletion
4. Filter out companies whose Name equals their ATS provider name (e.g. 'greenhouse', 'lever', 'ashby') — these are placeholder names from scraping
5. Move LLM Provider section to top of /coach page (ahead of Resume and Case Study)
6. Update navbar: replace Companies link with Settings link

## Tasks
- [x] Add CompanyID to Job domain type (or use existing CompanyID field) for hide button
- [x] Add 'Hide company' button + htmx partial to job detail page template
- [x] Add hide endpoint awareness to JobSearchHandler (needs CompanyService)
- [x] Filter out provider-name-only companies in ListHiddenCompanies
- [x] Create SettingsPage template with blocked companies list (search + unblock)
- [x] Create SettingsHandler with Show + unblock endpoint
- [x] Update /companies routes → /settings routes
- [x] Reorder CoachPage sections: LLM Provider first, then Resume, then Case Study
- [x] Update navbar: Companies → Settings
- [x] Update main.go wiring
- [x] Run templ generate + go build
- [x] Verify in browser

## Summary of Changes
Completed. Companies page replaced with settings page. Hide company button on job detail. Placeholder companies filtered out.
