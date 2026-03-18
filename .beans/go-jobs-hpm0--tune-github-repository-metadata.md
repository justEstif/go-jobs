---
# go-jobs-hpm0
title: Tune GitHub repository metadata
status: in-progress
type: task
priority: normal
created_at: 2026-03-18T17:27:36Z
updated_at: 2026-03-18T17:56:03Z
---

Update go-jobs GitHub repository metadata (description, homepage, and topics) using gh CLI.\n\n## Checklist\n- [x] Inspect current metadata fields\n- [ ] Update description, homepage, and topics\n- [ ] Verify metadata changes\n- [ ] Commit bean update\n\n## Blocker\n\n- Metadata write calls fail from current gh auth context: gh repo edit justEstif/go-jobs returns 404, and gh api PATCH repos/justEstif/go-jobs also fails.\n- Read access works (gh repo view succeeds), so this appears to be a repository-settings permission/auth mismatch for write operations.

\n## Attempt (latest)\n\n- Ran  from repository root.\n- Command failed with .\n- Verified read access and target repo identity with {"homepageUrl":"","nameWithOwner":"justEstif/go-jobs","url":"https://github.com/justEstif/go-jobs"} (returns ).\n- github.com
  ✓ Logged in to github.com account e-beyene (keyring)
  - Active account: true
  - Git operations protocol: ssh
  - Token: gho_************************************
  - Token scopes: 'admin:public_key', 'gist', 'read:org', 'repo' confirms token scopes include , so this remains a repository-settings permission mismatch for the authenticated account.
