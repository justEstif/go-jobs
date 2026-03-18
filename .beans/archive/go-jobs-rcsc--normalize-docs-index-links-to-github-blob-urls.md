---
# go-jobs-rcsc
title: Normalize docs index links to GitHub blob URLs
status: completed
type: task
priority: normal
created_at: 2026-03-18T17:47:16Z
updated_at: 2026-03-18T17:48:01Z
---

Update all documentation links in docs/index.html to use GitHub blob URLs on main branch (not raw links), aligned with README doc paths.\n\n## Checklist\n- [x] Update docs/index.html links to blob URLs\n- [x] Verify all href values in docs/index.html\n- [x] Add summary and complete bean\n\n## Summary of Changes\n\n- Replaced relative doc links in docs/index.html with GitHub blob links for README, ARCHITECTURE, interfaces, self-hosting, and MVP docs.\n- Kept non-doc links intact (#top, stylesheet CDN, repository root).\n- Normalized repository owner casing in repository links to justEstif.
