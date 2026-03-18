---
# go-jobs-cd2c
title: Fix design issues identified in architecture review
status: completed
type: task
priority: normal
created_at: 2026-03-18T10:01:14Z
updated_at: 2026-03-18T10:04:46Z
---

Apply corrections from the dual-lens (hexagonal + Ousterhout) design review of ARCHITECTURE.md and docs/interfaces.md:

1. Fix SearchFilters implicit UserID/OnlyNew coupling → separate UserSearchContext struct
2. Sync ARCHITECTURE.md ports tables with actual interfaces.md definitions
3. Extract SessionRepository from UserRepository
4. Add ListByIDs to JobRepository (fix ListByStatus N+1 gap)
5. Add ScrapeStatus typed string (consistency fix)
6. Clarify UserJobRepository.Upsert partial-update semantics
7. Remove stale GetOrCreate rationale from ARCHITECTURE.md

## Summary of Changes

### docs/interfaces.md
- **SearchFilters** — removed UserID and OnlyNew; introduced **UserSearchContext** struct with UserID + OnlyNew. Structural coupling eliminated: the precondition is now impossible to violate.
- **JobSearchService.Search** and **JobRepository.Search** — signature updated to .
- **JobRepository.GetByIDs** — new method for bulk job hydration; resolves the ApplicationService→UserJobRepository N+1 gap.
- **UserRepository** — stripped of token methods; now owns only user record persistence (4 methods).
- **SessionRepository** — new driven port for opaque session token management (SaveToken, DeleteToken, GetUserByToken). AuthService comment updated to document its dependency on both.
- **ScrapeStatus** — new typed string with ScrapeStatusRunning/Completed/Failed constants; ScrapeRun.Status changed from string to ScrapeStatus.
- **UserJobRepository.Upsert** — expanded comment documenting the read-modify-write contract, zero-value semantics for Notes, and AppliedAt immutability guarantee.
- **Notes section** — updated UserSearchContext and auth notes to reflect new types.

### ARCHITECTURE.md
- **Driving ports table** — corrected all 6 divergences from interfaces.md; added AuthService row; fixed ApplicationService methods.
- **Driven ports table** — added SessionRepository, ScrapeRunRepository, CompanySeeder rows; corrected UserRepository, UserJobRepository, JobRepository signatures.
- **Port decision rationale** — removed stale GetOrCreate rationale; added entries for SessionRepository, AuthService, CompanySeeder, JobRepository.GetByIDs.
- **System overview diagram** — added AuthService, HTTP middleware, SessionRepository, ScrapeRunRepository boxes.
- **Folder structure** — updated driven.go comment to mention SessionRepository.
