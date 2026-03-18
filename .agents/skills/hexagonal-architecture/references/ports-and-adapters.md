# Ports and Adapters Reference

## The two directions

Every port is either driving or driven. Both directions need ports — abstracting only one side is the most common incomplete application of the pattern.

| Direction | Also called | Who initiates                    | Examples                                            |
| --------- | ----------- | -------------------------------- | --------------------------------------------------- |
| Driving   | Primary     | External caller drives the core  | HTTP handler, CLI, test suite, AI agent, cron job   |
| Driven    | Secondary   | The core drives external systems | Database, mailer, search index, external API, queue |

**Driving ports** define how callers use the core — service and use case interfaces.
**Driven ports** define what the core needs from the world — repository, notification, enrichment interfaces.

## What makes a good port

A port earns its cost when callers are completely unaware of what implements it. The interface should be expressible in domain language alone — no technology concepts, no provider details, no format knowledge.

A port is shallow when removing it and calling the implementation directly would change nothing meaningful for callers. Shallow ports add indirection without hiding complexity — they are a red flag, not a feature.

## The test suite is a driving adapter

Because the core only depends on port interfaces, a test suite is just another driving adapter. It calls the same driving ports as production delivery mechanisms, and wires lightweight in-memory implementations of driven ports instead of real infrastructure. Testability without real infra is a structural consequence of the pattern — not something bolted on afterward.

## The composition root

All concrete adapter implementations are wired together in one place — the composition root (typically the application entry point). This is the only place that knows about all concrete types. Everything else depends only on interfaces. Dependency injection flows inward: the composition root instantiates adapters and passes them into the core.
