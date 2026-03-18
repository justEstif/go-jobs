---
# go-jobs-l1gp
title: 'M5: Auth'
status: completed
type: milestone
priority: normal
created_at: 2026-03-18T11:05:10Z
updated_at: 2026-03-18T11:14:00Z
---

CLI and web auth. Email + password registration and login. Session token stored in local config for CLI. Session cookie for web.

## Tasks

- [x] Implement AuthService (internal/core/services/auth.go): Register (bcrypt hash), Login (verify + generate token), Logout (delete token), Authenticate (resolve token → User)
- [x] CLI: go-jobs register — prompt email + password, create account
- [x] CLI: go-jobs login — prompt email + password, store token to $XDG_CONFIG_HOME/go-jobs/token
- [x] CLI: go-jobs logout — delete stored token
- [x] Wire AuthService in main.go and cli.Services
- [x] Web: POST /register handler (form → AuthService.Register → redirect)
- [x] Web: POST /login handler (form → AuthService.Login → set session cookie → redirect)
- [x] Web: POST /logout handler (AuthService.Logout → clear cookie → redirect)
- [x] Web: register and login templ pages
- [x] go build ./... passes clean

## Design Notes\n\n### Session library: alexedwards/scs v2\n- Replace gorilla/sessions with scs/v2\n- Use pgxstore (scs/v2/stores/pgxstore) — Postgres-backed, survives restarts\n- scs.SessionManager is itself an http.Handler middleware (Load+Save)\n- Store user ID as string in session under key "user_id"\n- Drop gorilla/sessions, gorilla/securecookie from go.mod\n- gorilla/csrf stays (CSRF is separate)

## Summary of Changes

- Added `internal/core/services/auth.go`: `AuthService` with Register (bcrypt), Login (token), Logout, Authenticate
- Replaced `gorilla/sessions` with `alexedwards/scs/v2` + `pgxstore` in `middleware/session.go`
- `SessionManager` now backed by Postgres `http_sessions` table (migration 005)
- `SetUserSession` no longer takes `http.ResponseWriter` (scs handles writes via `LoadAndSave` middleware)
- Added `middleware/session.go` `LoadAndSave` wired in `main.go` before all routes
- Added `internal/adapters/http/auth.go`: `AuthHandler` with ShowRegister, Register, ShowLogin, Login, Logout
- Added `components/auth.templ`: `RegisterPage` and `LoginPage` components
- Added `internal/cli/auth.go`: `register`, `login`, `logout` cobra commands
- Extended `cli.Services` with `Auth ports.AuthService`
- `gorilla/sessions` removed from go.mod; `gorilla/securecookie` remains (gorilla/csrf dependency)
