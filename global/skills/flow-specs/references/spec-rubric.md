# Spec quality rubric

Scored 1-5 per dimension by `spec-quality-reviewer`. Plain markdown on purpose: the same
rubric is consumable from any harness (Claude Code agent, Codex prompt). The agent brings
the judgment; this file brings the domain checklist.

| # | Dimension | What a 5 looks like |
|---|---|---|
| 1 | Problem & objective | Problem, affected actor, and "why now" explicit; no solution language in the problem statement |
| 2 | Persona & operating context | Who uses it, in what role, on what device/surface, under what tenant/org boundary |
| 3 | Scope fence | Both included AND excluded scope written (non-goals section exists and is real) |
| 4 | Main + alternate flows | Happy path plus the alternates that change behavior (retries, cancellations, concurrency) |
| 5 | Permissions & tenancy | Roles, visibility rules, and tenant/org limits stated per flow — not assumed |
| 6 | UI states | Empty, loading, error, success defined for every surface the epic touches |
| 7 | Data & contracts | Required data, its ownership, and affected contracts (OpenAPI/events) *identified by name* — never defined: shapes, fields, and schemas are TECH.md's job |
| 8 | Gherkin verifiability | Every AC executable as Given/When/Then; no "should work correctly" criteria |
| 9 | Success signal | Measurable acceptance: what QA runs, what the client signs off against |
| 10 | Risks & dependencies | Known risks, cross-epic dependencies, and open questions with owners |

## Hard checks (binary — not scored, every failure is at least a `gap`)

- **No technical leakage in PRODUCT.md.** Table/column shapes, endpoint definitions,
  token/session mechanics, algorithms, hashing/library choices, or any content whose home
  is TECH.md — each occurrence is at least a `gap` and the finding MUST name the
  relocation target (parked Open Question `technical — resolves in TECH.md`, or a task AC
  when the detail is genuinely verifiable behavior). *Naming* an affected contract or
  constraint is fine (dimension 7); *defining* it is the leak. A spec that resolves
  technical blockers inline instead of parking them fails this check regardless of how
  well-argued the resolutions are.
- **Affected vistas declared and the map consistent.** The `## Affected vistas` section
  exists, every vista it names exists in `product/` (or its map entry is part of this
  epic's draft), and the epic's business rules are written for the vista pages — not
  restated as product state inside PRODUCT.md. A missing map entry, or rules that live
  only in the epic, is a `gap`.
- **Code-layer identifiers are English.** Every identifier the spec *defines* — OpenAPI
  `path`/`property`, schema field, table/column/FK name, event payload key, request/response
  shape — must be English, even when the surrounding prose is Spanish. A Spanish identifier
  here is a `gap` (it persists into a migration/column once implemented, then costs a
  migration to fix). Governed by `CLAUDE.md > Code Identifier Language`.
  - **Not a finding:** Spanish *domain values* (enum literals, RBAC/permission keys like
    `finanzas.operacion.caja`). Flag these only when **inconsistent** with the domain's
    established precedent — the mixed-enum anti-pattern (some values English, some Spanish
    in one domain) — never for merely being Spanish.

## Severity mapping

- Dimension scored 1-2 with implementation pending → at least one `blocker` or `gap`
  finding must explain it.
- "Implicit business rule" findings (stated nowhere, assumed everywhere) are always at
  least `gap`, regardless of dimension scores.
