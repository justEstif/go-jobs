## Design Context

### Users
Active job seekers targeting startups — technical, self-sufficient, value efficiency over visual polish. They use go-jobs to get signal over volume: curated aggregation from Greenhouse, Lever, and Ashby with structured enrichment and built-in application tracking. They may use it via web UI or CLI. Context is focused and task-driven; they're not browsing leisurely.

### Brand Personality
Warm, precise, tool-first. The interface should feel like a well-made notebook — human and considered, not sterile. Three words: **warm, focused, reliable**.

### Aesthetic Direction
Warm-gray surfaces (OKLCH hue ~65, chroma 0.006–0.012) with a single sage green accent (`oklch(48% 0.10 155)`) used sparingly: CTAs, focus rings, active states, links. All surfaces are warm-tinted — never pure white or gray. Light mode only. Rounded fields (0.375rem), rounded boxes (0.5rem). Depth enabled, no grain.

Typography: **Fraunces** (display, serif, optical size, weight 300–600) for headings; **DM Sans** (body, humanist sans, weight 300–600) for UI text. Headings use `letter-spacing: -0.02em` and `font-weight: 300` for an editorial, airy quality. Numeric data uses `font-variant-numeric: tabular-nums`.

Status indicators: left-border accents on job rows — sage for Interested, warning-amber for Applied, muted for closed/rejected states.

Anti-references: avoid cool blue-gray palettes, glassmorphism, gradient text, neon accents. This is not sterile SaaS enterprise software — it has warmth and character while staying information-dense.

### Design Principles
1. **Content first** — UI chrome earns no attention; job listings and status states do.
2. **Restraint over decoration** — one accent color (sage), used at ~10%. Everything else is warm neutral.
3. **Warm-tinted neutrals** — hue 65 tint on all surfaces for subconscious warmth; never pure white or gray.
4. **Editorial hierarchy** — Fraunces headings create clear, memorable visual anchors; DM Sans body keeps it readable and friendly.
5. **Tool, not product** — interactions feel fast, keyboard-friendly, breathing room without excess whitespace.
