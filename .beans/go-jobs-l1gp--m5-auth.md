---
# go-jobs-l1gp
title: 'M5: Auth'
status: todo
type: milestone
created_at: 2026-03-18T11:05:10Z
updated_at: 2026-03-18T11:05:10Z
---

CLI and web auth. Email + password registration and login. Session token stored in local config for CLI. Session cookie for web.

## Tasks

- [ ] Implement AuthService (internal/core/services/auth.go): Register (bcrypt hash), Login (verify + generate token), Logout (delete token), Authenticate (resolve token → User)
- [ ] CLI: go-jobs register — prompt email + password, create account
- [ ] CLI: go-jobs login — prompt email + password, store token to $XDG_CONFIG_HOME/go-jobs/token
- [ ] CLI: go-jobs logout — delete stored token
- [ ] Wire AuthService in main.go and cli.Services
- [ ] Web: POST /register handler (form → AuthService.Register → redirect)
- [ ] Web: POST /login handler (form → AuthService.Login → set session cookie → redirect)
- [ ] Web: POST /logout handler (AuthService.Logout → clear cookie → redirect)
- [ ] Web: register and login templ pages
- [ ] go build ./... passes clean
