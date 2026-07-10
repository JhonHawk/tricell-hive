---
name: flow-hygiene
description: >
  Audit and repair the documentary hygiene of a client workspace: misplaced files,
  decisions never promoted to the specs repo, stale scratch, broken PROJECT.md pointers,
  specs-repo nonconformance. Use at milestone closes (epic done, pre-delivery) or whenever
  a workspace feels hard to navigate. audit reports; apply executes approved actions;
  migrate bootstraps an existing project (current or past) into the session convention.
---

# /flow-hygiene — workspace hygiene

The flow contract keeps flow-skill writes clean; conversational sessions are where
disorder accumulates. This skill is the compensating control: not a gate, an on-demand
sweep. Default subcommand: `audit`.

This skill is also the **adoption path for pre-pack workspaces**: `/flow-kickoff` refuses
existing workspaces by design, so a project that predates the pack enters the flow by
running `audit` + `apply` here — that retrofits the ledger and the workspace CLAUDE.md,
after which any phase gate (`/flow-plan`, `/flow-build`, `/flow-deploy`) works normally.

It also audits **session conformance** (`project-structure.md > Session capture layer`):
loose artifacts that belong grouped in a session (`move`), unpromoted session decisions
(`promote`), stale concluded sessions and raw committed by mistake (`expire`), non-ISO
session folders and broken `Session:` back-references (`conform`/`repair`),
**uncommitted session artifacts** in a versioned home — captured plans, findings, or
reports sitting untracked (`commit`, the standing-authorized `chore(sessions): <slug>`;
backstop of the capture flow) — plus **stale captured plans** (`Status: planned` older
than ~2 weeks with no matching execution: recommend adopt via `/flow-build`, conclude, or
expire), and **same-tier duplicate folders** (several `evidence`/`reports`/`spec`) to
consolidate (`move`) — always preserving the raw(`_support`, gitignored) /
curated(specs, versioned) split. An **initiative** folder (`specs-structure.md > Session & initiative conventions`)
is legitimate structure — never flag its nested executions as loose; conversely, a flat
session grown into a dense multi-part effort is a `move` finding to promote into an
initiative. The one-shot **reset/bootstrap** of an existing repo is the **`migrate`**
subcommand below (procedure: `~/.agents/skills/flow-core/references/migration-playbook.md`).

## `audit`

1. OPEN per the contract (`~/.agents/skills/flow-core/SKILL.md`). If PROJECT.md or the
   workspace ambient pair (AGENTS.md canonical + CLAUDE.md importing it via `@AGENTS.md`
   — what kickoff normally creates) is missing or incomplete, those are the first
   findings — propose creating them: the ledger from the template (reconstructing phase
   status from observable state: specs repo, git history, the tracker), the ambient pair
   per the kickoff skill's Phase 2 block. A ledger missing the `Tracker` / `Tracker
   access` fields is a `repair` finding: **detect the tracker from observable signals**
   — ticket-key patterns in commits and branches (`FAC-48` ~ Linear, `ATSCL-2406` ~
   Jira), conventions written in AGENTS.md/CLAUDE.md, which MCP/CLI responds (`acli`,
   Linear MCP) — and propose the field values with the evidence; the user confirms.
   A workspace with only a CLAUDE.md (pre-pack pattern) gets the inversion proposed:
   content moves to AGENTS.md, CLAUDE.md becomes the import.
2. Dispatch **workspace-custodian** (fresh context, read-only) with: the workspace root,
   the specs repo path, and the intent: "your proposals will be presented verbatim to the
   user by flow-hygiene apply — make every action executable as written".
   Do not audit from the main thread; the custodian's fresh eyes are the point.
3. **Assign stable IDs** to every finding, prefixed by classification (`R1, R2…` repair,
   `C1…` conform, `M1…` move, `P1…` promote, `E1…` expire). The IDs live in the saved
   artifacts — `apply` may run days later in a session with no memory of this one.
4. Save TWO artifacts, then present:
   - **Actions manifest** → `<project>/_support/workspace/hygiene-audit-<YYYY-MM-DD>.md`:
     one table row per finding — `ID | action | path | what/why (one line) | risk
     (safe/destructive)`. This is the machine-readable source `apply` consumes; keep it
     free of prose.
   - **Review report** → render via the flow-report skill to
     `<project>/_support/workspace/hygiene-audit-<YYYY-MM-DD>.html`: the manifest table
     plus the custodian's reasoning per finding, grouped by classification. Open it in
     the default browser (`open <path>` on macOS) — this is where the user reviews the
     audit; the chat only carries the summary.
   - In chat: counts per classification + the ID list (e.g. "repair 2: R1 R2 · expire 3:
     E1-E3") and the two paths. If everything is clean, say so and skip the artifacts.

## `apply`

1. Load the most recent **actions manifest** (`hygiene-audit-<date>.md`) — never re-read
   the HTML into context; it exists for the human. If no manifest exists or it's stale,
   run `audit` first. When the audit came from a prior session, re-open the HTML in the
   browser so the user reviews before approving.
2. Gate via AskUserQuestion. **The question may only reference IDs the user just saw**
   (in chat or in the just-opened report) — never invent groupings of undefined IDs.
   Group the options by RISK, not by classification: approve everything `safe` in one
   option; `destructive` actions (expire, moves that break references, legacy
   migrations) are reviewed per item or per small batch, and deletions execute only
   after **typed confirmation** — the user writes the literal word `eliminar` plus the
   IDs as presented (an option click is not enough to destroy files; hesitation means
   archive instead). Nothing executes without explicit approval.
3. Execute only what was approved:
   - `move`/`conform`: `git mv` when inside a repo, plain `mv` otherwise
   - `promote`: write the summary into `<project>-specs/decisions/` (dated file per the
     specs structure), then update the source's ledger row to point at it
   - `expire`: delete only with approval; when the user hesitated, archive instead
   - `repair`: edit PROJECT.md rows as proposed
4. CLOSE per the contract: update PROJECT.md (including pruning rows the actions made
   stale), report what was executed vs skipped, and delete the consumed audit artifacts
   (manifest + HTML). If actions were skipped, keep both and note in the manifest which
   IDs remain open instead.

## `migrate`

One-shot bootstrap of an existing project (current OR past) INTO the convention — the broader
sibling of `audit`+`apply`, adding specs-repo creation and active/archived tiering. The full
procedure (phases, classification, safety) is
`~/.agents/skills/flow-core/references/migration-playbook.md` — read it first; it is the source of
truth for the steps below.

1. OPEN per the contract. Detect **shape** (workspace/standalone; `<project>-specs/` present?) and
   **activity tier** — ACTIVE = full migration; ARCHIVED/inactive = minimal structure + sweep loose
   docs into `sessions/previously/` (dated by mtime/git), no archaeology. Unsure → ask; default ARCHIVED.
2. Dispatch **workspace-custodian** (fresh context, read-only) seeded with the migration playbook and
   the project root: produce a **migration manifest** covering specs-repo creation (if needed),
   per-artifact routing (intention / execution / raw / `previously/`), ISO date renames, and
   index/back-reference creation. Assign stable IDs prefixed by action (`N` new-repo, `M` move,
   `P` promote, `R` rename/repair, `E` expire, `C` conform). Raw file dumps stay in the subagent.
3. Save the manifest → `<project>/_support/workspace/migration-<YYYY-MM-DD>.md` (machine-readable
   table: `ID | action | source | destination | reason | risk`) and a review report via flow-report →
   `…/migration-<YYYY-MM-DD>.html`; open the HTML. In chat: counts per action + the ID list + the two
   paths. If the project already conforms, say so and skip the artifacts.
4. Gate via AskUserQuestion, grouped by RISK (not action): specs-repo creation, deletes, and
   reference-breaking moves are reviewed per item; safe moves/renames batch. Deletions execute only
   after typed `eliminar` + the IDs. Nothing runs without approval.
5. Execute only what was approved: `git init <project>-specs` (if approved) → structure + `.gitignore`
   + `.engram/config.json` + README map; `git mv` inside a repo / `mv` across the `_support/`→specs
   boundary; write promoted files, indexes, and back-references; rename to ISO; archive (never delete)
   on any doubt. Preserve content verbatim and the raw/curated split. **Workspace conventions
   refresh:** if the workspace AGENTS.md predates the artifact-first template (it suggests invoking
   `/flow-plan` → `/flow-build` for substantive work instead of declaring the session-artifact
   conventions), rewrite that block to the current `flow-kickoff` template — a coherent revision,
   not a string swap — and add the `sessions/**` pre-authorization to the workspace
   `.claude/settings.json` if absent.
6. CLOSE per the contract: update the ledger pointer (it does NOT index sessions), report executed vs
   skipped (counts to `sessions/` / `previously/` / intention / raw / pruned), and — if a specs repo
   was created — make its initial commit (single-branch new repo; the `migrate` invocation authorizes
   it; no remote unless the user adds one). Re-running `migrate` on the same project is safe:
   already-conforming artifacts produce no findings, so iterate freely.
