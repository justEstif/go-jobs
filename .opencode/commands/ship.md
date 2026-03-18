---
description: Stage all changes, commit with a generated message, and push
---

Stage all changes and ship them:

Current git status:
!`git status --short`

Staged diff:
!`git add -A && git diff --cached --stat`

Current branch:
!`git rev-parse --abbrev-ref HEAD`

Write a concise commit message (imperative mood, ≤72 char subject line) that accurately describes the staged changes, then run:
```
git commit -m "<your message>"
git push origin <branch>
```
