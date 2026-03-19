---
# go-jobs-i5t0
title: 'Optimize release workflow: split CLI and server builds'
status: in-progress
type: task
created_at: 2026-03-19T18:58:05Z
updated_at: 2026-03-19T18:58:05Z
---

CLI binary (cmd/cli) doesn't import web, components, or HTTP adapters — it doesn't need Tailwind, DaisyUI, or templ generate. Split the build matrix into a lightweight CLI job (all 5 platforms) and a server job (linux/darwin only, needs CSS + templ).
