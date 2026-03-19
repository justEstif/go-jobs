---
# go-jobs-qlam
title: Make CLI installable
status: completed
type: epic
priority: normal
created_at: 2026-03-18T18:52:51Z
updated_at: 2026-03-19T16:06:14Z
---

Enable users to install the go-jobs CLI via common package/distribution flows.

## Outcomes
- Users can install the CLI on macOS/Linux with a documented one-liner or package manager flow.
- Release artifacts are produced consistently for supported platforms.
- Installation and upgrade steps are documented and verifiable.

## Todo
- [ ] Decide distribution channels (Homebrew tap, GitHub Releases, direct binaries).
- [ ] Add reproducible cross-platform build/release pipeline.
- [ ] Add install/upgrade/uninstall documentation for each channel.
- [ ] Validate fresh install flow in a clean environment.
- [x] Add follow-up beans for implementation tasks under this epic.



## Summary of Changes

Completed CLI installable epic:
- npm distribution set up with  package
- GitHub Actions release workflow builds for 5 platforms
- Embedded static assets via go:embed
- Added documentation with npm badge, live demo badge, and install instructions
