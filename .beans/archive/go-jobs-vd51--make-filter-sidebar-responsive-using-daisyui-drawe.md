---
# go-jobs-vd51
title: Make filter sidebar responsive using daisyUI drawer
status: completed
type: task
priority: normal
created_at: 2026-03-18T23:26:46Z
updated_at: 2026-03-18T23:27:22Z
---

Replace the static flex sidebar in JobsListPage with a daisyUI drawer component (lg:drawer-open). Always visible on lg+, toggleable on mobile.

## Summary of Changes\n\n- Replaced  shell in  with daisyUI \n-  moves into , always visible on , overlay on mobile\n- Added hidden  checkbox input for drawer toggle state\n- Added mobile-only "Filters" button (funnel icon) above the search bar via  label\n-  aside updated to  for proper drawer styling\n- Build passes clean ()
