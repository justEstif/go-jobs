## Design Context

### Users
Active job seekers targeting startups — technical, self-sufficient, value efficiency over visual polish. They use go-jobs to get signal over volume: curated aggregation from Greenhouse, Lever, and Ashby with structured enrichment and built-in application tracking. They may use it via web UI or CLI. Context is focused and task-driven; they're not browsing leisurely.

### Brand Personality
Quiet, precise, tool-first. The interface should disappear — chrome recedes so job listings and actions become the visual focus. Three words: **minimal, focused, reliable**.

### Aesthetic Direction
Near-monochrome with cool-tinted slate grays. A single precise slate-blue accent (`oklch(42% 0.10 255)`) used sparingly: CTAs, focus rings, active states, links. All surfaces use OKLCH with subtle chroma (0.005–0.015 at hue 240) — never pure white or gray. Light mode only. Slightly rounded fields (0.25rem), slightly more rounded boxes (0.375rem). Depth enabled, no grain.

Anti-references: avoid purple-blue gradients, glassmorphism, gradient text, neon accents, "developer dark mode with glowing accents" aesthetic. This is a focused productivity tool, not a portfolio or marketing page.

### Design Principles
1. **Content first** — UI chrome earns no attention; job listings and status states do.
2. **Restraint over decoration** — one accent color, used at 10%. Everything else is neutral.
3. **Tinted neutrals** — cool-slate tint on all surfaces for cohesion; never pure white or gray.
4. **Clarity without loudness** — hierarchy through weight and space, not color saturation.
5. **Tool, not product** — interactions feel fast, keyboard-friendly, information-dense.
