---
# go-jobs-hpm0
title: Tune GitHub repository metadata
status: in-progress
type: task
priority: normal
created_at: 2026-03-18T17:27:36Z
updated_at: 2026-03-18T17:28:31Z
---

Update go-jobs GitHub repository metadata (description, homepage, and topics) using gh CLI.\n\n## Checklist\n- [x] Inspect current metadata fields\n- [ ] Update description, homepage, and topics\n- [ ] Verify metadata changes\n- [ ] Commit bean update\n\n## Blocker\n\n- Metadata write calls fail from current gh auth context: gh repo edit justEstif/go-jobs returns 404, and gh api PATCH repos/justEstif/go-jobs also fails.\n- Read access works (gh repo view succeeds), so this appears to be a repository-settings permission/auth mismatch for write operations.
