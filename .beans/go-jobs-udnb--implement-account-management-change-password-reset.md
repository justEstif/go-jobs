---
# go-jobs-udnb
title: 'Implement account management: change password, reset tracker, delete account'
status: completed
type: feature
priority: normal
created_at: 2026-03-19T17:25:21Z
updated_at: 2026-03-19T17:31:04Z
---

Add three account management features to the settings page:

1. Change password — verify current password, set new one  
2. Reset tracker — clear all user_jobs (pipeline state)
3. Delete account — confirm with password, delete user row (cascades to sessions, user_jobs, user_companies, coach_cache)

## Tasks
- [x] Add SQL queries: UpdatePassword, DeleteUser, DeleteUserJobs
- [x] Run sqlc generate
- [x] Add new methods to driven ports (UserRepository, UserJobRepository)
- [x] Add new methods to driving port (UserService)
- [x] Implement service layer methods (ChangePassword, ResetTracker, DeleteAccount)
- [x] Implement postgres adapter methods
- [x] Add SettingsHandler endpoints (POST /settings/password, /settings/reset-tracker, /settings/delete-account)
- [x] Update SettingsPage template with forms (confirmation dialogs for destructive ops)
- [x] Wire routes in main.go
- [x] Run templ generate + go build
- [x] Test in browser

## Summary of Changes
Completed. Change password, reset tracker, and delete account all implemented end-to-end.
