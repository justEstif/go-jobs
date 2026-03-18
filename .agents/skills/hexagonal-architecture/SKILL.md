---
name: hexagonal-architecture
description: Apply ports and adapters (hexagonal architecture) principles when designing or reviewing architecture. Use when the user is starting a new project, planning module boundaries, asking where logic should live, discussing coupling or testability, reviewing existing structure for architectural feedback, or asking questions like "should I abstract this", "where does this code go", "is this too coupled", or "how do I structure this". Also triggers when the user shares a project structure for feedback, mentions ports/adapters/driving/driven, or is designing a boundary between system components.
---

# Ports and Adapters (Hexagonal Architecture)

Apply the pattern to enforce a clean boundary between business logic and everything around it. Focus on boundaries that matter — not dogmatic purity.

## Core vocabulary

| Term    | Synonym      | Meaning                                                              |
| ------- | ------------ | -------------------------------------------------------------------- |
| Core    | The fortress | Pure business logic — owns domain types, rules, and port definitions |
| Port    | The gate     | Interface defined by the core — specifies _what_, not _how_          |
| Adapter | The bridge   | Concrete implementation connecting a technology to a port            |

Ports come in two directions — see [references/ports-and-adapters.md](references/ports-and-adapters.md).

## Mode selection

- **Design mode**: User is planning a new system or module boundary
- **Review mode**: User shares existing structure and wants architectural feedback
- **Both**: Redesigning an existing system

---

## Design mode

### 1. Identify the core

What are the business rules? What domain concepts live here? The core should be expressible without mentioning any technology.

### 2. Identify all external connections

What calls into the system? What does the system call out to? Each connection is a candidate port.

### 3. Define ports by purpose

Name and define each port in terms of what the domain needs — not what any technology provides. A port named after a business concept (`UserRepository`, `PaymentGateway`) is right. A port named after a technology (`PostgresStore`, `StripeClient`) belongs in the adapter layer.

### 4. Apply the port decision checklist

Before defining a port, ask:

1. Will this ever need to be swapped? (test vs prod, vendor change)
2. Will this be tested in isolation?
3. Does crossing this boundary encode a business rule?

If all three are no — the abstraction may not earn its cost.

### 5. Check both directions

Both driving and driven connections need ports. See [references/ports-and-adapters.md](references/ports-and-adapters.md).

---

## Review mode

### 1. Locate the core boundary

Where does business logic live? Is it clearly separated from delivery (HTTP, CLI) and infrastructure (DB, external APIs)?

### 2. Evaluate the boundary

| Question                                    | What to look for                                                           |
| ------------------------------------------- | -------------------------------------------------------------------------- |
| Does the core depend on infra?              | Dependencies should point inward — infra knows about core, not the reverse |
| Are ports defined by purpose or technology? | Technology-named ports in core signal leakage                              |
| Are both directions abstracted?             | Driving side (how callers reach core) matters as much as driven side       |
| Is business logic leaking outward?          | Domain rules appearing in handlers, routes, or adapters                    |
| Are ports hiding real complexity?           | Shallow ports add ceremony without value                                   |

### 3. Scan for red flags

See [references/red-flags.md](references/red-flags.md).

### 4. Report findings

```
**[Problem]** — severity: high|medium|low
Location: which layer or boundary
Problem: what is wrong and why it matters architecturally
Suggestion: concrete fix at the boundary level
```

Distinguish conscious tradeoffs from actual violations. Not every impurity costs the same.

### 5. Score the architecture (0–10)

- **9–10**: Clean core boundary, ports defined by purpose, both directions abstracted, fully testable without real infra
- **7–8**: Solid separation with minor leakage or a shallow port or two
- **5–6**: Core boundary exists but business logic leaks into delivery or infra
- **3–4**: Significant coupling — domain concepts tied to infrastructure concerns
- **0–2**: No meaningful separation — business logic, delivery, and infra entangled throughout

State the score and the specific boundary changes needed to reach the next level.

---

## Pragmatic application

The pattern is a map, not a mandate. Common conscious tradeoffs:

- Cross-cutting concerns (auth, logging) living outside the core boundary are often fine
- Simple operations may not need a use case layer between delivery and infra
- Ports over trivial single-method operations may not hide enough to justify the indirection

Always weigh the cost of the abstraction against the complexity it hides.
