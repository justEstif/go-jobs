---
# go-jobs-ldc0
title: npm distribution for go-jobs CLI
status: completed
type: task
priority: normal
created_at: 2026-03-18T21:55:32Z
updated_at: 2026-03-18T21:57:02Z
parent: go-jobs-qlam
---

Distribute the go-jobs CLI as @justestif/go-jobs on npm. Binary command name is 'jobs'. Website is jobs.estifanos.cc.

## Todo
- [x] Rename binary and cobra root Use field from 'go-jobs' to 'jobs'
- [x] Update mise.toml build task output to bin/jobs
- [x] Create npm/go-jobs/package.json
- [x] Create npm/go-jobs/install.js (downloads binary from GitHub Releases on postinstall)
- [x] Create npm/go-jobs/bin.js (thin exec wrapper)
- [x] Create .github/workflows/release.yml (build → GitHub Release → npm publish)
- [x] Update AGENTS.md with install instructions
- [ ] Manual step: set NPM_TOKEN secret in GitHub repo settings
