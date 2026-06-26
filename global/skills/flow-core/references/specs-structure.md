# Specs repo structure — conventions

Used by `/flow-specs init` to create `<project>/<project>-specs/` and by `/flow-hygiene` to
check conformance. Derived from Warp's specs organization
(github.com/warpdotdev/warp/tree/master/specs, extracted 2026-06-11) and adapted to this
workflow: tracker épicas → tasks with Gherkin ACs, plus the conventions/decisions layers the
flow pack needs.

What Warp gets right and we keep: **one folder per work item, named by its tracker ID**, a
**product spec separated from the tech spec**, specs written as the source of truth that
implementation must match, and tech specs that cite real code paths (`file.rs:24-145`) so
they stay verifiable against the codebase.

## Repository layout

```
<project>-specs/
├── README.md                 # Index: epic table (ID, name, status, links) + pointer to sessions/README.md — the map
│                             # ── INTENTION (rules / prior analysis) ──
├── conventions/
│   └── naming.md             # Instantiated naming table (see naming-template.md)
├── contracts/                # OpenAPI specs, event schemas — cross-repo source of truth
├── decisions/
│   └── YYYY-MM-DD-<slug>.md  # Promoted decisions (from a session / ledger / temp reports)
├── evidence/
│   └── <epic-id>/            # in-vivo reports (test-report-template.md), text-only — no images; raw stays in _support/
├── releases/                # Client-facing release notes per promotion (release-notes-template.md): YYYY-MM-DD-<env>.md
├── epics/
│   └── <EPIC-ID>-<slug>/     # e.g. E07-mensajeria or PROJ-360-cicd
│       ├── PRODUCT.md        # What and why — written at F3, gate: /flow-specs review
│       ├── TECH.md           # How — written when foundation exists; cites real code paths
│       └── tasks.md          # Task list mirroring the tracker, Gherkin ACs per task
│                             # ── EXECUTION (what actually happened) ──
└── sessions/                 # Execution journal (by time) — see project-structure.md "Session capture layer"
    ├── README.md             # Sessions index (versioned, co-located): slug · date · state · implements
    ├── previously/           # reset quarantine (loose/legacy artifacts swept in at bootstrap)
    └── YYYY-MM-DD-<slug>/     # <slug>-plan.md (declares `Implements:`), <slug>-findings.md, optional analysis/ reports/
```

## PRODUCT.md skeleton (Warp-derived)

```markdown
# <Epic name> — Product Spec
Tracker: <epic URL or — if untracked> · Mock: <route in the mocks repo, if it exists>

## Summary            <!-- 3-5 lines: the change, for whom, why now -->
## Problem            <!-- the pain with evidence; no solution language here -->
## Goals              <!-- bullet list, each verifiable -->
## Non-goals          <!-- explicit scope fence — what this epic deliberately won't do -->
## User Experience    <!-- subsections per flow/surface; states (empty/loading/error),
                           roles and tenant boundaries, edge cases -->
## Success Criteria   <!-- measurable; what QA/the client accepts against -->
## Validation         <!-- how it will be verified: E2E journeys, in-vivo gates -->
## Open Questions     <!-- each blocks something specific; owner named -->
```

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
  is a defect (`/flow-hygiene` flags it).
- **PRODUCT.md before TECH.md, TECH.md before code.** A PRODUCT.md merges only after
  passing `/flow-specs review`. TECH.md cites real paths — it cannot be written before the
  foundation exists, and it goes stale loudly (paths stop resolving) rather than silently.
- **README.md is the index, not a document.** One table: epic ID, name, status
  (draft / reviewed / in development / delivered), links. Same role the ledger plays for
  the workspace — if it's not in the index, it's invisible.
- **Status lives in the spec header**, mirrored to the README table. Vocabulary:
  `draft → reviewed → in development → delivered` (+ `parked`).
- **Decisions are dated files**, `YYYY-MM-DD-<slug>.md`: context, decision, consequences,
  and a back-reference to the session that produced or changed them
  (`Implementado en: sessions/<slug>`, path noted even if the raw expired).
- **Sessions are the execution layer.** `sessions/YYYY-MM-DD-<slug>/` records what was
  actually done (plan, findings, reports); raw (logs/dumps/screenshots) stays gitignored under
  `_support/`, never here. The intention layer (decisions/contracts/epics/conventions) is
  updated BY execution and back-references the session slug (`Session:` in `tasks.md`).
  `sessions/README.md` is the versioned, co-located index — NOT the ledger, which is
  non-versioned and lives outside this repo. Full rule:
  `project-structure.md > Session capture layer`. Reset/bootstrap of an existing repo sweeps
  loose artifacts into `sessions/previously/`.
- **Evidence here is summarized** — the curated proof a delivery points to (E2E report
  summary, sign-off notes). Raw traces/screenshots stay in `_support/evidence/`.
- **Release notes are dated client-facing records**, `releases/YYYY-MM-DD-<env>.md` — the
  durable history of what each QA/prod promotion delivered (client language, usted). The
  draft is reviewed before sending; the committed file is the record. Never `_support/workspace/`
  (gitignored), which is the bug this rule fixes.
- Documents in the client's language; Gherkin keywords stay in English
  (`Given/When/Then`) for tooling compatibility.
- **Optional Docusaurus presentation layer.** A specs repo MAY carry a Docusaurus site
  (opted in at `/flow-specs init`) rendering the same content as a navigable site. Its
  scaffold (`package.json`, `docusaurus.config.*`, `sidebars.*`, `src/`, `static/`) is
  conformant and sits *over* the structure above — it never replaces it: `conventions/`,
  `contracts/`, `decisions/`, `epics/`, `sessions/` stay the source of truth at their
  paths. `/flow-hygiene` treats the scaffold as expected, not as misplaced files.
