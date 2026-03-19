---
# go-jobs-i04j
title: Add static company seed list to supplement SimplifySeeder
status: completed
type: feature
priority: normal
created_at: 2026-03-19T13:02:41Z
updated_at: 2026-03-19T13:06:34Z
---

Add hundreds of Greenhouse/Lever/Ashby companies found via gh CLI scraping of multiple job repos (Summer2025, Summer2026, New Grad 2024-2026, etc.) as a static seed list that gets merged with the dynamic SimplifySeeder output.

## Summary of Changes\n\n- Created  with ~270 curated companies across all three ATS platforms\n- Updated  in  to merge  into the dedup map after parsing READMEs\n- Greenhouse: ~200 entries (Anthropic, Jane Street, Figma, Grammarly, SpaceX, Twilio, Vercel, Robinhood, etc.)\n- Lever: 39 entries (Palantir, Anduril, Zoox, Palantir, Whoop, etc.)\n- Ashby: 80 entries (OpenAI, Notion, Ramp, Snowflake, Replit, Cohere, Vanta, etc.)\n- Dynamic README sources still take priority; static entries only fill gaps
