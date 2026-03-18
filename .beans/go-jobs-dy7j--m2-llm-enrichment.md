---
# go-jobs-dy7j
title: 'M2: LLM Enrichment'
status: todo
type: milestone
created_at: 2026-03-18T10:34:22Z
updated_at: 2026-03-18T10:34:22Z
---

Tiered enrichment pipeline: ATS metadata extraction (tier 1), rule-based tagging (tier 2), and LLM enrichment via OpenAI/Anthropic/Google SDKs (tier 3). CLI enrich command. Enrichment source tracked per job.

## Tasks

- [ ] Enrichment adapter: ats.go (tier 1 — extract from RawJob ATS metadata fields)
- [ ] Enrichment adapter: rules.go (tier 2 — keyword/regex on title + description)
- [ ] Enrichment adapter: llm.go (tier 3 — LLM structured output, user-provided API key)
- [ ] Enrichment adapter: enricher.go (orchestrates tiers 1→2→3)
- [ ] Add OpenAI, Anthropic, Google SDKs to go.mod
- [ ] CLI enrich command (internal/cli/enrich.go)
- [ ] Wire enricher in main.go
- [ ] go build ./... passes clean
