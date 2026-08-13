# Specs repo structure — conventions

Used by `/flow-specs init` to create `<project>/<project>-specs/` and by `/flow-workspace` to
check conformance. Core shape: **one folder per work item, named by its tracker ID**; the
**product spec separated from the tech spec**; specs as the source of truth implementation
must match, with tech specs citing real code paths (`file.rs:24-145`) so they stay
verifiable; and a **product layer** (`product/`) holding the business truth in force — the
map of portals, actors, and modules, plus the current business rules per vista. Epics are
*deltas* against that layer; the product layer is the accumulated *state*. A business reader
reads `product/`; a developer plans from `epics/`.

## Repository layout

```
<project>-specs/
├── README.md                 # Index: epic table (ID, name, status, links) + pointer to sessions/README.md
│                             # ── INTENTION (rules / prior analysis) ──
├── product/                  # Business truth IN FORCE — the product map + current rules per vista
│   ├── README.md             # The map: portals/surfaces, actors & roles, end-to-end flow
│   └── <module>/             # e.g. usuarios/, finanzas/ — one folder per business module
│       ├── index.md          # What the module is, for which actor(s); lists its vistas
│       └── <vista>.md        # Current business rules of one vista (screen/surface)
├── conventions/
│   └── naming.md             # Instantiated naming table (see naming-template.md)
├── contracts/                # OpenAPI specs, event schemas — cross-repo source of truth
├── decisions/
│   └── YYYY-MM-DD-<slug>.md  # Promoted decisions (from a session / ledger / temp reports)
├── evidence/
│   └── <epic-id>/            # in-vivo reports (test-report-template.md), text-only — no images; raw stays in _support/
├── releases/                # Client-facing release notes per promotion (release-notes-template.md): YYYY-MM-DD-<env>.md
├── epics/
│   └── <EPIC-ID>-<slug>/     # e.g. E07-mensajeria or TRI-360-cicd
│       ├── PRODUCT.md        # The DELTA: what changes and why — written at the specs stage, gate: /flow-specs review
│       ├── TECH.md           # How — written when foundation exists; cites real code paths
│       └── tasks.md          # Task list mirroring the tracker, Gherkin ACs per task — derived post-gate
│                             # ── EXECUTION (what actually happened) ──
├── audit/                    # Multi-lens audit instance (audit-playbook.md): README.md = protocol instance + history table
│   └── reports/              #   one YYYY-MM-DD.html per run — stable home so baselines compare across runs
└── sessions/                 # Execution journal (by time) — full convention below
    ├── README.md             # Sessions index (versioned, co-located): slug · date · state · implements
    ├── previously/           # reset quarantine (loose/legacy artifacts swept in at bootstrap)
    └── YYYY-MM-DD-<slug>/     # <slug>-plan.md (declares `Implements:`), <slug>-findings.md, optional analysis/ reports/
```

## product/ — the product map and vistas (business truth in force)

The product layer answers "what is this product, and what are its rules **today**" — it is
what a business reader (client, PO, new team member) reads, and what the Starlight sidebar
renders. It grows only with what has been *defined*: a product born with just auth and user
management has one module with its vistas, nothing else.

- **`product/README.md` — the map.** Portals/surfaces (e.g. client portal vs backoffice),
  actors and roles (who enters which portal), the end-to-end flow of the platform, and the
  module list. Reference shape: the "Introducción" layer of a user manual — written as
  *intention* before anything is built.
- **`product/<module>/index.md` — the module.** Short: what capability it groups, for which
  actor(s), its vistas with one line each. Never replaces the vista pages.
- **`product/<module>/<vista>.md` — the vista.** The business rules IN FORCE for one
  screen/surface — the accumulated result of every epic that shaped it, not a per-epic
  restatement. Two epics touching the same vista converge here instead of contradicting
  each other in two documents.

Vista skeleton — headings in the **client's language** (business-facing; Spanish shown as
the example):

```markdown
# <Vista> — <Módulo>

## Propósito y actores      <!-- qué resuelve, quién la opera (rol), en qué portal -->
## Reglas de negocio        <!-- numeradas; las VIGENTES tras el último gate de negocio -->
## Estados y transiciones   <!-- estados de la entidad/flujo; mermaid opcional -->
## Casos borde              <!-- límites, vacíos, concurrencia -->
## Influenciada por         <!-- trazabilidad, ver regla abajo -->
```

**Traceability (`Influenciada por`).** One line per epic that created or modified the
vista: `- E04 — define los 4 roles operativos (gate 2026-06-18)`. The epic's *status* is
NOT duplicated here — it lives once, in the README epic index (the site's SpecRubric
renders it by lookup); the vista entry records only who contributed what, and when it
passed the business gate.

## PRODUCT.md skeleton (Warp-derived; the epic's DELTA, not product state)

```markdown
# <Epic name> — Product Spec
Tracker: <epic URL or — if untracked> · Mock: <route in the mocks repo, if it exists>

## Summary            <!-- 3-5 lines: the change, for whom, why now -->
## Problem            <!-- the pain with evidence; no solution language here -->
## Affected vistas    <!-- MANDATORY: which product/ vistas this epic CREATES or MODIFIES,
                           with the one-line delta per vista. The rules themselves land in
                           the vista pages at gate pass — never restated here -->
## Goals              <!-- bullet list, each verifiable -->
## Non-goals          <!-- explicit scope fence — what this epic deliberately won't do -->
## User Experience    <!-- only what the DELTA changes; current behavior lives in product/ -->
## Success Criteria   <!-- measurable; what QA/the client accepts against -->
## Validation         <!-- how it will be verified: E2E journeys, in-vivo gates -->
## Open Questions     <!-- each blocks something specific; owner named. Technical questions
                           surfaced at the business gate are PARKED here (marked
                           `technical — resolves in TECH.md`), never answered in this file -->
```

**PRODUCT.md carries no technical content.** Schemas, endpoints, table/column shapes,
token/session mechanics, algorithms, and library choices belong to TECH.md (foundation
material, now `/flow-start`). A
technical question that surfaces while drafting or reviewing the epic is *parked* as an
Open Question with an owner — resolving it inside PRODUCT.md is the defect this rule
exists to prevent (the review gate hardens it: `spec-rubric.md` hard checks).

## TECH.md skeleton (Warp-derived)

```markdown
# <Epic name> — Tech Spec
Product spec: ./PRODUCT.md

## Problem            <!-- the technical constraint set, not a repeat of PRODUCT -->
## Relevant Code      <!-- real paths with line refs across affected repos -->
## Current State      <!-- how it works today, per repo/path -->
## Proposed Changes   <!-- numbered; each independently reviewable -->
## End-to-End Flow    <!-- request/data flow across services after the change -->
## Implementation Plan<!-- phases that map to dev-session scopes -->
## Risks and Mitigations
## Testing and Validation
```

## tasks.md skeleton

```markdown
# <Epic name> — Tasks
Tracker epic: <URL or — if untracked>

## <TASK-ID> — <task name>
Status: pending | in progress | done · Repo: <repo> · Session: <YYYY-MM-DD-slug or —>

```gherkin
Given <precondition>
When <action>
Then <verifiable outcome>
```
```

## Conventions

- **Epic folder = tracker epic.** The ID prefix comes from the tracker declared in the
  ledger (Linear team key, Jira project key); with `Tracker: none`, IDs are self-assigned
  (`E07`, …) and tasks.md is the source of truth. If an epic is re-scoped in the tracker,
  the spec updates in the same change — divergence between the tracker and the specs repo
  is a defect (`/flow-workspace` flags it).
- **The map precedes epics.** An epic may only reference vistas/modules that exist in
  `product/` — when the epic introduces a new one, creating the map entry is part of the
  epic's draft. At business-gate pass (`/flow-specs review`), the epic's rules are applied
  to the vista pages and the `Influenciada por` entries are added: the vista absorbs the
  *decided* rules, even before they are built — the epic index status tells the reader
  what is decided vs delivered.
- **PRODUCT.md before TECH.md, TECH.md before code — and each has its own gate.** A
  PRODUCT.md merges only after passing `/flow-specs review` (the **business gate**: rules,
  scope, verifiability — technical findings get parked, not resolved). TECH.md cites real
  paths — it cannot be written before the foundation exists, and it goes stale loudly
  (paths stop resolving) rather than silently; it closes the epic's parked technical
  questions, verified at the `flow-plan` plan gate (the **technical gate**).
- **tasks.md is derived AFTER the business gate.** Task decomposition (units of work,
  repo assignment, per-task Gherkin) is delivery planning over *gated* rules — writing
  it pre-gate means every business finding invalidates already-synced tasks. Business
  acceptance scenarios (happy + negative) belong in PRODUCT.md/the vistas and ARE gated;
  the tracker never sees pre-gate tasks.
- **README.md is the index, not a document.** One table: epic ID, name, status
  (draft / reviewed / in development / delivered), links. Same role the ledger plays for
  the workspace — if it's not in the index, it's invisible.
- **Status lives in the spec header**, mirrored to the README table. Vocabulary:
  `draft → reviewed → in development → delivered` (+ `parked`). Convention files
  (`conventions/`, or repo-scoped ones in `<repo>/_support/docs/`) carry their own lifecycle
  instead: `proposed → adopted → superseded`.
- **Decisions are dated files**, `YYYY-MM-DD-<slug>.md`: context, decision, consequences,
  and a back-reference to the session that produced or changed them
  (`Implementado en: sessions/<slug>`, path noted even if the raw expired).
- **Sessions are the execution layer.** `sessions/YYYY-MM-DD-<slug>/` records what was
  actually done (plan, findings, reports); raw (logs/dumps/screenshots) stays gitignored under
  `_support/`, never here. The intention layer (product/decisions/contracts/epics/conventions)
  is updated BY execution and back-references the session slug (`Session:` in `tasks.md`).
  `sessions/README.md` is the versioned, co-located index — NOT the ledger, which is
  non-versioned and lives outside this repo. Full convention: "Session & initiative
  conventions" below (the always-loaded summary is `project-structure.md > Session capture
  layer`). Reset/bootstrap of an existing repo sweeps loose artifacts into
  `sessions/previously/`.
- **Evidence here is summarized** — the curated proof a delivery points to (E2E report
  summary, sign-off notes). Raw traces/screenshots stay in `_support/evidence/`.
- **Release notes are dated client-facing records**, `releases/YYYY-MM-DD-<env>.md` — the
  durable history of what each QA/prod promotion delivered (client language, usted). The
  draft is reviewed before sending; the committed file is the record. Never `_support/workspace/`
  (gitignored), which is the bug this rule fixes.
- Documents in the client's language; Gherkin keywords stay in English
  (`Given/When/Then`) for tooling compatibility.
- **Optional Astro Starlight presentation layer.** A specs repo MAY carry a Starlight
  site (opted in at `/flow-specs init`, scaffolded via the `starlight-docs-site` skill's
  `spec-site` profile) rendering the same content as a navigable site. Its scaffold
  (`package.json`, `astro.config.*`, `src/`, `public/`) is conformant and sits *over* the
  structure above — it never replaces it: `product/`, `conventions/`, `contracts/`,
  `decisions/`, `epics/`, `sessions/` stay the source of truth at their paths.
  `/flow-workspace` treats the scaffold as expected, not as misplaced files.
  - **The sidebar IS the product map** — Introducción (map: qué es, actores y roles,
    flujo completo) → one group per module with its vistas → an appendix group (épicas,
    decisiones, requisitos) as reference material, last. Delivery taxonomy (epic IDs,
    E01–E12) is never the navigation axis.
  - **Workflow status never rides in sidebar labels** ("draft (bloqueada en Q-03)" in a
    label is the anti-pattern). Status renders in-page via the profile's `SpecRubric`
    component, looked up from the epic index.

## Session & initiative conventions (canonical)

The convention (`workflow/session-capture.md`, loaded via the `workspace-conventions` skill) carries the
summary — two axes, detection rule, raw-out-of-git, lifecycle. This section is the full
convention; it applies wherever sessions live (specs repo, standalone `_support/sessions/`).

**Naming.** Folder: `YYYY-MM-DD-<kebab-slug>` (ISO date prefix; lexicographic =
chronological). Multiple sessions the same day → distinct intention-revealing slugs; a
numeric tiebreaker (`-2`) only on a real slug collision. **Session top-level files carry
the SLUG, not the date** — `<slug>-plan.md`, `<slug>-findings.md`, `<slug>-report.html`:
the slug makes a hit self-identifying in basename-only surfaces (quick-open, editor tabs,
filename/semantic search) where the folder path isn't shown; the date stays the folder's
(a date on a living file asserts a fixity it doesn't have). Nested subfolder files
(`analysis/…`, `reports/…`) stay short — their path is already specific. This is a
deliberate, scoped exception to "internal files unprefixed": sessions are high-volume and
referenced individually; a one-off deliverable folder is not.

**Structure is optional and proportional** — no fixed skeleton. Create a type-subfolder
(`reports/`, `analysis/`, `internal/`) only at 2+ artifacts of that type. A trivial
session is a `<slug>-plan.md` plus a couple of loose files. Homologate HOW artifacts are
grouped, not WHAT files exist.

**Initiative (multi-session grouping).** When one effort exceeds a single session — a
dense plan split into parts, executed across several sessions or days — group it under
`sessions/<start-date>-<slug>/` holding `README.md` (index), `findings/` (research),
`plan/` (master plan + numbered parts `00-NN`), and the dated execution sub-sessions
INSIDE it (`YYYY-MM-DD-<sub>/`, each with its own `reports/`). The initiative container
carries its immutable **start date** (preserves the index's chronological order; the
slug-not-date rule governs files, not this container); sub-sessions carry their own dates,
so a later day nests inside the initiative instead of fragmenting into a sibling folder.
Each plan part carries its own `Status` (`plan-format.md`); the master plan's part index
lists them. **One-off work stays a flat session** — promote to an initiative only when it
grows (the move is `/flow-workspace`'s; `flow-plan write` proposes the split when scope
density warrants it).

**Back-reference (by slug).** A session's `<slug>-plan.md` declares `Implements:` the
intention it executes; the intention records the session that implemented or changed it
(`Session:` in `tasks.md`, `Implementado en: sessions/<slug>` in decisions/epics). When
execution diverges from the spec, update the spec (source of truth) and record the session
slug as the origin — `gap-resolution.md > Divergence Between Sources`.

**Lifecycle** (recorded in the sessions index, never the ledger): `in-progress` (active
`<slug>-plan.md`) → `concluded` (work done, promotion pending — `/flow-workspace` flags it)
→ `finalized` (durable outputs promoted, raw pruned). A concluded/finalized session is
immutable — a later correction supersedes with a new linked record, never an in-place
edit.
