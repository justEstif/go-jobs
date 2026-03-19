---
# go-jobs-yb5s
title: Refactor service boundaries per Ousterhout design review
status: completed
type: task
priority: high
created_at: 2026-03-19T18:49:33Z
updated_at: 2026-03-19T18:56:23Z
---

Implement all 7 refactoring items from the design review to improve codebase score from 5/10 toward 8/10.

## Tasks

- [x] 1. Move ChangePassword and DeleteAccount from UserService to AuthService
- [x] 2. Move ResetTracker from UserService to ApplicationService
- [x] 3. Add IsCompanyHidden(ctx, userID, companyID) to CompanyService (new SQL query)
- [x] 4. Move ErrInvalidCredentials from services package to ports/errors.go
- [x] 5. Remove unused coach field from JobSearchHandler
- [x] 6. Add IsCoachReady() method to domain.User
- [x] 7. Update all call sites (handlers, composition roots, ports)


## Summary of Changes

All 7 design review items implemented:

1. **ChangePassword/DeleteAccount → AuthService** — Moved from UserService to AuthService since they verify passwords with bcrypt (auth logic). Removed bcrypt dependency from user.go. UserService no longer needs UserJobRepository.
2. **ResetTracker → ApplicationService** — Moved since it operates on UserJobRepository which ApplicationService already owns.
3. **IsCompanyHidden → CompanyService** — Added new SQL query (`IsCompanyHidden :one`) and repo method for efficient single-row check. Job detail handler no longer fetches ALL hidden companies.
4. **ErrInvalidCredentials → ports/errors.go** — Moved from services package to ports package. HTTP handlers now reference `ports.ErrInvalidCredentials` instead of importing the services package.
5. **Removed unused coach field from JobSearchHandler** — Handler dropped from 5 deps to 4. Constructor simplified.
6. **IsCoachReady() on domain.User** — Encapsulates `Resume != "" && LLMProvider != ""` check. Template simplified from `hasResume, hasLLM` params to single `coachReady` bool.
7. **All call sites updated** — Both composition roots (cmd/jobs, cmd/cli), SettingsHandler (now takes AuthService + ApplicationService instead of UserService), httpclient adapters (stub methods added), and templ templates.
