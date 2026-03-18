---
# go-jobs-dy7j
title: 'M2: LLM Enrichment'
status: completed
type: milestone
priority: normal
created_at: 2026-03-18T10:34:22Z
updated_at: 2026-03-18T10:55:41Z
---

Tiered enrichment pipeline: ATS metadata extraction (tier 1), rule-based tagging (tier 2), and LLM enrichment via OpenAI/Anthropic/Google SDKs (tier 3). CLI enrich command. Enrichment source tracked per job.

## Tasks

- [x] Enrichment adapter: ats.go (tier 1 — extract from RawJob ATS metadata fields)
- [x] Enrichment adapter: rules.go (tier 2 — keyword/regex on title + description)
- [x] Enrichment adapter: llm.go (tier 3 stub — deferred to M5; tiers 1+2 always run)
- [x] Enrichment adapter: enricher.go (orchestrates tiers 1→2→3)
- [x] LLM SDKs deferred to M5 (tier 3 stub, no external deps needed)
- [x] CLI enrich command (internal/cli/enrich.go)
- [x] Wire enricher in main.go
- [x] go build ./... passes clean

## Summary of Changes

Implemented the full tiered enrichment pipeline:

- `internal/adapters/enrichment/ats.go` — tier 1: extracts RemotePolicy, Country, LocationNorm from raw ATS fields; maps department strings to RoleType
- `internal/adapters/enrichment/rules.go` — tier 2: keyword/regex on title (seniority, role type) and description (tech stack, remote policy, country phrases)
- `internal/adapters/enrichment/llm.go` — tier 3 stub; returns ErrNoLLMKey when no API key configured; full implementation deferred to M5
- `internal/adapters/enrichment/enricher.go` — TieredEnricher orchestrates all three tiers; sets EnrichmentSource; finalizes defaults (RoleOther, SeniorityMid, empty TechStack slice)
- `internal/core/services/enrich.go` — EnrichService impl: ListUnenriched → Enrich → SaveTags loop
- `internal/core/ports/driving.go` — added EnrichService driving port
- `internal/cli/enrich.go` — `go-jobs enrich --limit N` cobra command
- `internal/cli/root.go` — added Enrich to Services struct; registered enrich command
- `cmd/jobs/main.go` — wired TieredEnricher and EnrichService; replaced nil enricher placeholder
- `go build ./... && go vet ./...` both pass clean; `gofmt -l .` reports no unformatted files
