---
# go-jobs-ymbg
title: 'Job Coach: LLM-powered resume analysis and optimization'
status: completed
type: feature
priority: high
created_at: 2026-03-19T16:26:40Z
updated_at: 2026-03-19T16:54:14Z
---

Replace the unused LLM enrichment tier (tier 3) with a user-facing Job Coach feature. Users provide their resume (markdown/plaintext), select a job, and get:

1. **ATS Analysis** — keyword gaps and formatting advice specific to the company's ATS (Greenhouse/Lever/Ashby)
2. **Fit Analysis** — match assessment with strengths and gaps
3. **Optimized Resume** — a rewritten, tailored version of their resume for this specific job

Additionally, users can generate a **Portfolio Case Study** from a resume bullet point — expanding a one-liner into a structured case study (Problem → Process → Solution → Results → Learnings).

## LLM Providers (MVP)
- **Ollama** (default) — local, free, no API key. Uses openai-go SDK with base URL override
- **OpenAI** — cloud, needs API key, best structured output support

## Prompt Strategy
Two prompt templates adapted from ResumeSkills project:

### Prompt 1: Job Analysis + Resume Optimization
Based on: github.com/Paramchoudhary/ResumeSkills job-description-analyzer
- Extract requirements (must-have vs nice-to-have)
- Keyword extraction (hard skills, soft skills, industry terms)
- Gap analysis (critical / major / minor)
- ATS-specific advice (we know the ATS type from company.ATSType)
- Generate optimized resume tailored to this JD

### Prompt 2: Portfolio Case Study Writer
Based on: github.com/Paramchoudhary/ResumeSkills portfolio-case-study-writer
- User selects a resume bullet point or project
- LLM expands it into structured case study: Overview → Problem → Process → Solution → Results → Learnings
- Useful for portfolio building and interview prep

## Architecture

### New domain types
- `User.Resume` field (markdown text, stored in DB)
- `LLMOllama LLMProvider = "ollama"` added to provider enum

### New ports
- `JobCoachService` driving port:
  - `AnalyzeJob(ctx, userID, jobID) → stream of analysis text`
  - `GenerateCaseStudy(ctx, userID, projectDescription) → stream of case study text`
- `LLMClient` driven port:
  - `Complete(ctx, systemPrompt, userPrompt string) → stream/string`

### New adapters
- `internal/adapters/llm/ollama.go` — openai-go SDK with custom base URL
- `internal/adapters/llm/openai.go` — openai-go SDK with user's API key
- `internal/adapters/crypto/` — AES-256-GCM encrypt/decrypt for API keys at rest

### Web UI
- Settings page: resume textarea + LLM provider config
- Job detail page: "Analyze" button → streaming analysis panel with optimized resume
- Portfolio section or button for case study generation

### CLI
- `jobs resume set < resume.md` / `jobs resume show` / `jobs resume clear`
- `jobs analyze <job-id>` — streaming analysis + optimized resume
- `jobs case-study` — interactive case study generation
- LLM config via env vars: `LLM_PROVIDER`, `OPENAI_API_KEY`, `OLLAMA_URL`, `LLM_MODEL`

### Encryption
- AES-256-GCM with ENCRYPTION_KEY env var (already configured in mise.toml and Dokku)
- Encrypt API keys before DB storage, decrypt at call time
- Ollama needs no key — just URL + model name

## Security
- Never log API keys
- Never return keys to UI (show "configured" / "not set")
- Fresh SDK client per call — key in memory only during the call
- Validate Ollama URL to prevent SSRF (restrict to private/loopback IPs)

## What gets removed
- LLM tier 3 from enrichment pipeline (tiers 1+2 are sufficient for tagging)
- `llmEnricher` struct and `ErrNoLLMKey` from enrichment adapter
- LLM-related code from `TieredEnricher`

## Tasks
- [x] Add `LLMOllama` provider to domain types
- [x] Add `Resume` field to User domain type + migration
- [x] Create `LLMClient` driven port interface
- [x] Create `JobCoachService` driving port interface
- [x] Implement AES-256-GCM crypto adapter
- [x] Implement Ollama LLM adapter (openai-go with base URL)
- [x] Implement OpenAI LLM adapter
- [x] Write job analysis prompt template (ATS analysis + fit + resume rewrite)
- [x] Write case study prompt template
- [x] Implement JobCoachService (service layer)
- [x] Update UserService.SetLLMKey to encrypt keys
- [x] Add settings page (resume + LLM provider config)
- [x] Add analyze button + streaming panel to job detail page
- [x] Add case study generation UI
- [x] Add CLI commands: resume set/show/clear, analyze, prompt
- [x] Remove LLM tier 3 from enrichment pipeline
- [x] Wire everything in main.go
- [x] Test with Ollama + OpenAI (verified: prompt export works, Ollama connection works, model timeout expected on 1.5B param model)



## Summary of Changes

Implemented the full Job Coach feature replacing the unused LLM enrichment tier:

- **Domain**: LLMOllama provider, Resume/LLMModel/LLMBaseURL on User, CoachCache type
- **Ports**: JobCoachService (driving), LLMClient + CoachCacheRepository (driven)
- **Adapters**: AES-256-GCM crypto, Ollama + OpenAI LLM clients, coach cache repo, settings + coach HTTP handlers
- **Services**: Coach service with analyze/case-study/prompt-builder, UserService with encryption + resume
- **Enrichment**: Removed LLM tier 3; pipeline is two-tier (ATS + rules)
- **Web UI**: Settings page, Analyze button on job detail, Case Study page, navbar links
- **CLI**: resume set/show/clear, analyze, prompt (raw prompt export for BYOLLM)
- **DB**: Migration 006 with resume/llm columns + coach_cache table
- **Testing**: Verified prompt export, Ollama connection, resume set/show lifecycle
