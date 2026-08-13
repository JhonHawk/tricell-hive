# Documentary migration playbook

Canonical procedure for bringing an existing project (current OR past) into the session-based
documentary convention. Consumed by **`/flow-adopt`**; also self-contained enough to
paste to any project's agent (including non-Claude-Code harnesses) where the skill is not deployed.

The full convention is `specs-structure.md > Session & initiative conventions` (summary:
`project-structure.md > Session capture layer`); if this playbook and the convention ever disagree,
the convention wins. Always run **audit → manifest → approve → execute**: read-only first,
nothing destructive without typed confirmation.

---

## The convention (compact recap)

Three layers, by versioning need and axis:

- **Intention** (durable, versioned, BY TYPE) — `product/` (business rules in force), `decisions/`
  (dated ADRs), `contracts/`, `epics/`, `conventions/`. Business rules and prior technical
  analysis. Stays intact.
- **Execution** (versioned journal, BY TIME) — `sessions/YYYY-MM-DD-<slug>/`: the real work record
  (`<slug>-plan.md`, `<slug>-findings.md`, analysis, reports — top-level files carry the slug, not the date). Immutable once concluded.
- **Raw** (NOT versioned) — logs, dumps, raw screenshots, video, build output →
  `_support/workspace|evidence/YYYY-MM-DD-<slug>/`, gitignored; referenced by path from the session.

Placement rules: versioned home is `<project>-specs/` (workspace) or `<repo>/` (standalone, whose
`_support/` is versioned except `workspace/`); dates are ISO `YYYY-MM-DD` (prefix); the index is the
co-located versioned `sessions/README.md` (NOT a new `INDEX.md`, NOT the ledger); back-reference by
slug (`Implements:` ↔ `Implementado en: sessions/<slug>`); structure optional/proportional
(type-subfolders only for 2+ artifacts of a type) — homologate HOW, not WHAT.

---

## Phase 0 — Detect shape + activity tier (read-only)

1. **Shape.** Workspace (folder of N child git repos, maybe a `<project>-specs/` sibling) or
   standalone repo? Does `<project>-specs/` exist? Ledger `_support/PROJECT.md`? `.engram/config.json`
   (`<group>-<project>`)? Resolve homes — sessions: `<project>-specs/sessions/` ·
   `<repo>/_support/sessions/` · or a new specs repo; raw: `_support/workspace|evidence/`.
2. **Activity tier** (decides depth):
   - **ACTIVE** (ongoing/recent work): full migration — create the specs repo if missing, structure
     sessions, promote durable content, build indexes, reconstruct recent obvious session clusters.
   - **ARCHIVED / INACTIVE** (kept for reference): light touch — minimal structure, sweep existing
     docs into `sessions/previously/` (dated by mtime/git), index, STOP. No archaeology, no
     reconstructed history, no rewriting old content.
   - Unsure → ask the user; default ARCHIVED (least invasive).

## Phase 1 — Inventory (read-only)

- **Include:** `_support/` + subfolders, loose files at the root, `docs/`, `reports/`, `evidence/`,
  `analysis/`, `spec/`, `plan/`, `todos/`, `backup/`, `tmp/`, `context-ia/`, `manuals/`, and any
  legacy-named folder.
- **Exclude:** source (`src/`, `app/`; config repos: `global/`, `harness/`, `.claude/`), `.git/`,
  `node_modules/`, build outputs, lockfiles, and already-conforming specs layers.
- Per artifact: path, type, date (name/frontmatter/`git log`/mtime), size, reproducible? (raw vs durable).
- Flag SAME-tier duplication (consolidate); never flag the legitimate raw(`_support`)/curated(versioned) split.

## Phase 2 — Create the specs repo if missing (ACTIVE only, conditional)

Workspace + no `<project>-specs/` + ACTIVE (or user asks): `git init <project>-specs` (no remote
needed); folders that will have content (`sessions/` + promotion targets); `.gitignore`
(`.DS_Store`, `*.swp`, `*.swo`, `*~`); `.engram/config.json` (`{ "project_name": "<group>-<project>" }`
— may be globally gitignored; leave uncommitted unless a remote will share it); `README.md` map.
Standalone repos and ARCHIVED workspaces skip this — sessions live in `<repo>/_support/sessions/`
(committed) or `<project>/_support/sessions/` (staging).

## Phase 3 — Classify & route each artifact

> **Classify by CONTENT, never by filename or the folder it was found in.** The decisive split is
> WHAT/WHY (intention) vs HOW/WHEN-it-was-done (execution).

- **Implementation / execution plan** (task-by-task steps, `- [ ]` checkboxes, a "how to execute"
  doc — e.g. a `*-plan.md`, an `Implementation Plan` header, `executing-plans` /
  `subagent-driven-development` markers) → this is **EXECUTION, not intention**: route to
  `sessions/YYYY-MM-DD-<slug>/<slug>-plan.md` (its own session), even when it is paired with a
  design/spec/decision of the same topic+date. The paired design/spec/decision stays in the
  intention layer; the two cross-reference by slug. A dated plan is a point-in-time execution
  record — it does NOT belong in `decisions/` (the #1 misclassification: a plan read as "prior
  analysis").
- **Durable intention** — a *decision/ADR* (context/decision/consequences), *design/spec* (the WHAT
  and WHY), convention, or contract → its by-type home (`decisions/YYYY-MM-DD-<slug>.md`,
  `conventions/<topic>.md`, `epics/<id>/`, `contracts/`), with a back-reference to the producing
  session if known. NOT task-by-task plans (those are execution, above).
- **Reproducible raw** → raw home (gitignored), dated slug; clearly expired → propose deletion.
- **Reconstructable session bundle** → `sessions/YYYY-MM-DD-<slug>/` (intention-revealing slug); add a
  metadata header (date, slug, state, implements, immutable) and an `Implements:` line.
- **Not reconstructable / orphan** → `sessions/previously/` (quarantine), sub-grouped by date/origin;
  raw landing here splits out to the gitignored raw home.
- **Promote vs keep:** reusable-beyond-one-effort (a convention, a sealed decision) is PROMOTED to the
  intention layer; effort-specific narrative stays in the session. One canonical copy; the other side
  references it by slug (no duplication).

## Phase 4 — Homologate dates

Rename non-ISO date tokens (`DD-MM-YYYY`, `DD_MM_YYYY`, `DDMMYYYY`, `MM-DD-YYYY`, …) to ISO
`YYYY-MM-DD`, using the artifact's own date. Folders carry the date; in-place living files do not.

## Phase 5 — Indexes & back-references

- Create/update `sessions/README.md` (versioned, co-located): `Date | Slug | State | Implements | Docs`.
- Create/update the repo `README.md` map (intention + execution layers, pointers).
- Seal back-references by slug. Do NOT add session state to the ledger (`_support/PROJECT.md`).

## Phase 6 — Execute (with approval)

- **MANIFEST** table: `id | action | source | destination | reason | risk`, grouped by RISK. Nothing
  runs without approval. Higher-risk: specs-repo creation, deletes, moves that break references.
- `git mv` inside a repo; plain `mv` across a repo boundary (e.g. `_support/` → a new specs repo).
- Delete non-reproducible only after typed confirmation (`eliminar` + ids); doubt → archive.
- Preserve the raw/curated split; preserve content verbatim when moving/promoting (migrate, don't rewrite).
- If a specs repo was created: stage and make its initial commit at the end.

## Output

MANIFEST + per-finding reasoning, then a SUMMARY: counts to `sessions/`, `previously/`, intention
homes, raw, pruned; whether a specs repo was created; what was NOT touched and why. ARCHIVED projects
get a short summary (structure + sweep + index).

## Product-layer adoption (specs repos that predate `product/`)

A separate, deliberate migration — user-approved on its own, never a side effect of the
documentary sweep above. For a specs repo with epics but no `product/` layer
(`specs-structure.md > product/`):

1. **Derive the map** (`product/README.md`) from the reviewed/delivered epics and the
   requirements doc: portals/surfaces, actors and roles, end-to-end flow, module list.
2. **Extract the vistas.** Per module, write each vista's rules in force from the epics'
   PRODUCT.md content (rules live once, in the vista; the epic keeps the delta). Reconstruct
   `Influenciada por` per epic with its gate date from the README index / git history.
3. **Relocate technical content** already written into PRODUCT.md: to the epic's TECH.md
   (draft, if foundation exists) or to parked Open Questions (`technical — resolves in
   TECH.md`). Preserve every decision — relocate, don't rewrite; loss here is the migration's
   main risk, so diff-review each epic.
4. **Reorder the Starlight sidebar** if a site exists: product topology first (map → modules →
   vistas), epics/decisions/requirements as the appendix; status out of sidebar labels, into
   the in-page rubric.

## Safety invariants

- Read-only audit first; execution gated by explicit approval; deletes need typed confirmation.
- Never put raw into the versioned layer; never put durable rules into `sessions/`.
- Never reconstruct historical sessions for ARCHIVED projects (no fabricated dates/attribution).
- Never commit secrets; commit only secret templates.
