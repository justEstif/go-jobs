# Hexagonal Architecture Red Flags

## Inward dependency violation

**Signal**: The core depends on something outside itself — an infrastructure concern, a delivery mechanism, or a specific technology.

**Why it matters**: The core can no longer be understood or tested independently. Every change to infra risks touching business logic.

**Fix**: Define a port in the core. Invert the dependency — the adapter depends on the core, not the reverse.

---

## Business logic in the delivery layer

**Signal**: Domain rules, validation, or conditional business behaviour living in HTTP handlers, CLI commands, Server Actions, or similar delivery mechanisms.

**Why it matters**: Logic becomes inaccessible to other callers. The same rule can't be reused without duplicating it. Testing requires going through the delivery mechanism.

**Fix**: Move the rule into a use case or domain service in the core. The delivery layer calls the use case and maps the result to its own format.

---

## Technology-named ports in the core

**Signal**: Port interfaces named after specific technologies or vendors rather than business concepts.

**Why it matters**: The name reveals an implementation assumption. The port is already thinking about _how_, not _what_ — the abstraction is leaking.

**Fix**: Rename the port to reflect the business concept it serves. The adapter can know what technology it uses; the port should not.

---

## Shallow port

**Signal**: A port interface that barely abstracts anything — callers still effectively need to know how the implementation works.

**Why it matters**: Adds interface surface area and indirection without hiding complexity. Creates the illusion of decoupling without the reality.

**Fix**: Either deepen the port by giving it more to hide, or remove it. A port that isn't hiding real complexity is noise.

---

## Only one direction abstracted

**Signal**: Driven ports exist (e.g. repositories) but the driving side is unabstracted — business logic is called directly from delivery code with no intervening interface.

**Why it matters**: The test suite can't call business logic without going through delivery. Swapping the delivery mechanism requires touching business logic.

**Fix**: Define driving ports for non-trivial use cases. The delivery layer calls the use case interface; tests call the same interface directly.

---

## Domain concepts leaking into adapters

**Signal**: Business rules or domain decisions made inside an adapter — format choices that embed domain meaning, validation that belongs in the core, branching on domain concepts.

**Why it matters**: Business logic is now split between core and adapters. Changing a rule requires finding all the places it lives.

**Fix**: Adapters translate — they do not decide. Push the decision into the core, pass the result to the adapter to execute.

---

## Cross-adapter knowledge

**Signal**: Two or more adapters share knowledge of the same format, structure, or protocol — changing one requires changing the other.

**Why it matters**: The shared knowledge is an implicit coupling that bypasses the port boundary.

**Fix**: Consolidate the shared knowledge into a single adapter, or extract it into one place that both adapters depend on.

---

## Quick reference

| Red flag                    | Core symptom                   | Fix direction                                   |
| --------------------------- | ------------------------------ | ----------------------------------------------- |
| Inward dependency violation | Core imports infra             | Invert — define a port, adapter depends on core |
| Business logic in delivery  | Rules in handlers/routes       | Move to use case in core                        |
| Technology-named ports      | Port reveals implementation    | Rename to business concept                      |
| Shallow port                | No real complexity hidden      | Deepen or remove                                |
| Only driven side abstracted | Delivery touches core directly | Define driving ports                            |
| Domain logic in adapters    | Rules split across layers      | Adapters translate, core decides                |
| Cross-adapter knowledge     | Change one, touch another      | Single owner for shared knowledge               |
