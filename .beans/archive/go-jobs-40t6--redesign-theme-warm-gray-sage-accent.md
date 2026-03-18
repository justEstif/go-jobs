---
# go-jobs-40t6
title: 'Redesign theme: warm-gray + sage accent'
status: completed
type: task
priority: normal
created_at: 2026-03-18T12:20:18Z
updated_at: 2026-03-18T12:21:20Z
---

Replace cool slate theme with warm-gray surfaces and sage green accent. Update input.css and layout.templ for new font pairing (Fraunces + DM Sans). Update web-design.md to reflect new design context.

## Summary of Changes\n\n- ****: Replaced cool-slate (hue 240) palette with warm-gray surfaces (hue 65). Primary changed from slate-blue to sage green `oklch(48% 0.10 155)`. Secondary updated to warm amber-gray. Accent updated to lighter sage. Neutral surfaces all warm-tinted. Border-radius slightly increased (fields 0.375rem, boxes 0.5rem).\n- ****: Added Google Fonts preconnect + Fraunces (display serif) + DM Sans (humanist sans) font loading. Injected inline `<style>` block applying font families to body and headings, with antialiasing, kerning, and tabular-nums utility.\n- ****: Updated design context to reflect warm-gray + sage direction, new font pairing, and status indicator conventions.
