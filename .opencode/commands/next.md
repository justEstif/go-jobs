---
description: Pick the next ready bean, mark it in-progress, and start work
---

Next ready beans:
!`beans list --json --ready`

Uncommitted changes (if any):
!`git status --short`

Do the following:
1. Take the **first** bean from the list above
2. Mark it in-progress: `beans update <id> -s in-progress`
3. If there are uncommitted changes, run `/ship` first before starting
4. Read the bean body carefully with `beans show --json <id>`
5. Implement exactly what it describes, following AGENTS.md conventions
6. When done: mark it completed and run `/ship`
