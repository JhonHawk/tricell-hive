---
alwaysApply: true
---

## Research-Driven Decisions

Cross-reference architecture decisions, patterns, and trade-offs against industry consensus — not only these rules:

- **Cross-reference with industry consensus.** Don't limit recommendations to what these rules say. Cross-reference against established sources (e.g., *Clean Code*, *Designing Data-Intensive Applications*, *Patterns of Enterprise Application Architecture*).
- **Verify library and framework patterns.** When a recommendation touches a specific library, framework, or version-sensitive pattern, query current docs before answering — `tools/context7.md` owns the mandatory protocol. Libraries evolve; what was correct last year may be deprecated today.
- **Pin the version before researching — installed for existing deps, latest for new ones.** On an already-installed library, read the installed version (lockfile/manifest) and anchor research to it; do not look up latest. Adding a new dependency or weighing an upgrade, determine the current latest stable and its project compatibility. Either way, never anchor to the version in training data (it lags releases, and majors change syntax wholesale). `tools/context7.md > Anchor to the right version` owns the operational steps.
- **Flag when a rule conflicts with industry consensus.** If a rule in this config contradicts well-established practices or has a better alternative backed by evidence, surface it. Explain the trade-off and let the user decide.
- **Cite your reasoning.** When recommending a pattern or architecture, reference why — the principle, the source, or the evidence. "Prefer composition over inheritance" is stronger as "Prefer composition over inheritance (GoF, Effective Java §18) because it avoids fragile base class problems and makes testing easier."
- **Stay current.** Patterns like CQRS, event sourcing, hexagonal architecture, and serverless have nuances that evolve with the ecosystem. Don't apply patterns dogmatically — evaluate fit for the specific context, scale, and team size.
