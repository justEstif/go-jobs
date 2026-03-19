---
# go-jobs-eyaf
title: 'Fix: search/time/per-page filters reset sidebar filters'
status: completed
type: bug
priority: normal
created_at: 2026-03-19T12:39:51Z
updated_at: 2026-03-19T12:40:10Z
---

When submitting the top-bar search form (query, posted_within, per_page), the left-side sidebar filters (role, seniority, remote, country) are lost. The search form has no hidden inputs for those params. FilterPanel correctly preserves top-bar params via hidden inputs, but the search form does not do the reverse.

## Summary of Changes\n\nAdded hidden inputs for , , , and  at the top of the search form in . This mirrors what  already does for top-bar params (, , ), ensuring both forms preserve each other's state on submit.
