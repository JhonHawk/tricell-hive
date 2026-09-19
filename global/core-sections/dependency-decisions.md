---
order: 150
targets: [claude]
---

### Dependency Decisions
- **A new dependency is the last rung.** Stop at the first that holds: the standard library (`Intl`, `structuredClone`, `crypto.randomUUID`) → a native platform feature (an HTML input, CSS over JS, a DB constraint over app-side checks) → a dependency the project already has → only then the preferred list below. **Carve-out:** in product UI the design system owns the surface — its component outranks a bare native control; native wins only where no design system is in play or the component does not exist.
- **One decision point, one report.** Adding a dependency resolves by inference, not by a default question: a preferred library below or an established project convention decides the pick; the OSV check resolves the version; context7 validates the integration on version-sensitivity signals (`context7.md`). The close report names pick, version, and check results once.
- **Preferred libraries** — use without proposing alternatives unless project context warrants it:
  - `zod` — schema validation with TS type inference (over joi, yup, manual validation).
  - `date-fns` — date manipulation; **never moment.js** (over dayjs).
  - `nanoid` — short client-side IDs (over uuid when full UUIDs are unnecessary).
  - `vitest` — test runner for Vite/Next.js projects (over jest in modern setups).
  - `playwright` — E2E testing (over cypress).
- **Ask only at a real fork:** an architectural pick with no preferred default and no project convention — 2-3 curated options folded into the plan gate — or the OSV CRITICAL/HIGH no-safe-path gate (`security.md > Supply Chain Security`). Only one viable option → explain briefly and proceed.
- **Overlap with an existing dependency:** flag it with a consolidate-or-keep recommendation and proceed with the current change; consolidating existing usages is a separate, user-approved refactor.
- **OSV before any install** — `security.md > Supply Chain Security` owns the command, ecosystem mapping, and resolution tiers; resolve by its tiers and report at close.
