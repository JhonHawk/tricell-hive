---
name: flow-hygiene
description: >
  Audit and repair workspace health — git state, stray files, broken pointers; convention
  is suggested, not enforced. Use at milestone closes (epic done, pre-delivery) or whenever
  a workspace feels hard to navigate. audit reports proposed actions; apply executes the
  approved ones; migrate is the opt-in path for bringing a pre-pack project into the flow.
argument-hint: "[audit | apply | migrate]"
disable-model-invocation: true
---

# /flow-hygiene — workspace health

The flow contract keeps flow-skill writes clean; conversational sessions are where disorder
accumulates. This skill is the compensating control: not a gate, an on-demand sweep guided by
judgment. Default subcommand: `audit`.

It is also the **adoption path for pre-pack workspaces**: `/flow-start` refuses fully-established
workspaces by design, so a project that predates the pack enters the flow through `migrate` here
(or `audit` + `apply` for a lighter retrofit) — reconstructing the ledger and the ambient pair,
after which any phase gate (`/flow-plan`, `/flow-build`) works normally.

## What to judge (the center)

Three questions, all judgment — never a pass/fail conformance sweep:

1. **Git state.** Uncommitted session artifacts in a versioned home (captured plans, findings,
   reports sitting untracked) → propose the standing-authorized `chore(sessions): <slug>` commit.
   Moves use `git mv` inside a repo, plain `mv` across a repo boundary. Stale branches / leftover
   worktrees are noted, not force-resolved.
2. **Loose or stray files — preserve or not?** Judge each: keep in place, relocate, or expire.
   **Archive on doubt** — any hesitation archives (never a silent delete). Stale captured plans
   (`Status: planned`, older than ~2 weeks with no matching execution) → recommend adopt via
   `/flow-build`, conclude, or expire. Deletions execute ONLY after the user types `eliminar`
   plus the IDs. Preserve the raw (`_support`, gitignored) / curated (specs, versioned) split.
3. **Broken pointers.** Session back-references that dangle, stale `PROJECT.md` rows, a missing
   `Tracker` / `Tracker access` field. Detect the tracker from observable signals — ticket-key
   patterns in commits and branches (`FAC-48` ~ Linear, `ATSCL-2406` ~ Jira), conventions in
   AGENTS.md/CLAUDE.md, which MCP/CLI responds — and propose the field values with the evidence;
   the user confirms.

**Convention is suggested, never enforced.** Canonical structure, ISO date renames, naming
schemes, initiative-vs-flat layout: SUGGEST them when they clearly help navigation, never audit
them as findings. A workspace that deviates from the convention without harm produces NO finding.

## `audit`

1. OPEN per the contract (`~/.claude/skills/flow-core/SKILL.md`). If PROJECT.md or the workspace
   ambient pair (AGENTS.md canonical + CLAUDE.md importing it via `@AGENTS.md`) is missing or
   incomplete, those are the first proposals: reconstruct the ledger from the template (phase
   status inferred from observable state — specs repo, git history, the tracker), the ambient pair
   per the flow-start bootstrap template block. A workspace with only a CLAUDE.md (pre-pack
   pattern) gets the inversion proposed: content moves to AGENTS.md, CLAUDE.md becomes the import.
2. Dispatch **workspace-custodian** (fresh context, read-only) with the workspace root, the specs
   repo path, and the intent: "your proposals will be presented verbatim to the user by
   flow-hygiene apply — make every action executable as written". The custodian's fresh eyes are
   the point; don't audit from the main thread.
3. Save the **actions manifest** → `<project>/_support/workspace/hygiene-audit-<YYYY-MM-DD>.md`:
   one simple list, a row per proposed action — `action · path · why (one line) · risk
   (safe/destructive)`. This is the machine-readable source `apply` consumes; keep it free of
   prose. Give each row a short stable ID (any scheme) so `apply` can reference it days later in a
   session with no memory of this one.
4. When the audit is substantial (per `communication-format.md`), ALSO render a review report via
   the flow-report skill to `…/hygiene-audit-<YYYY-MM-DD>.html` and open it (`open <path>` on
   macOS); a small audit needs only the chat summary.
5. In chat: action counts by risk + the ID list and the manifest path. Clean → say so, skip the
   artifacts.

## `apply`

1. Load the most recent **actions manifest** (`hygiene-audit-<date>.md`) — never re-read the HTML
   into context; it exists for the human. No manifest or it's stale → run `audit` first. Audit from
   a prior session → re-open the HTML (if any) so the user reviews first.
2. Gate via AskUserQuestion, **grouped by RISK, not by action type** — the question may only
   reference IDs the user just saw. Approve everything `safe` in one option; `destructive` actions
   (expirations, moves that break references, legacy migrations) are reviewed per item or small
   batch, and deletions execute only after **typed confirmation** — the user writes the literal
   word `eliminar` plus the IDs (a click is not enough to destroy files; hesitation means archive
   instead). Nothing executes without explicit approval.
3. Execute only what was approved:
   - `move`: `git mv` inside a repo, plain `mv` otherwise
   - `promote`: write the summary into `<project>-specs/decisions/` (dated file), then point the
     source's ledger row at it
   - `expire`: delete only with typed approval; on hesitation, archive instead
   - `repair`: edit PROJECT.md rows as proposed
4. CLOSE per the contract: update PROJECT.md (pruning rows the actions made stale), report executed
   vs skipped, and delete the consumed audit artifacts. If actions were skipped, keep them and note
   in the manifest which IDs remain open.

## `migrate`

The opt-in path for bringing an existing project (current OR past) into the flow — the broader
sibling of `audit` + `apply`, adding specs-repo creation and active/archived tiering. Mechanics
live in `~/.claude/skills/flow-core/references/migration-playbook.md` — **read it first; it is the
source of truth.** Work by its principles, not a rigid checklist:

1. OPEN per the contract. Detect **shape** (workspace/standalone; `<project>-specs/` present?) and
   **activity tier** by judgment — ACTIVE = full adoption; ARCHIVED/inactive = minimal structure +
   sweep loose docs into `sessions/previously/` (dated by mtime/git), no archaeology. Unsure → ask;
   default ARCHIVED.
2. Dispatch **workspace-custodian** (fresh, read-only) seeded with the playbook and the project
   root to produce a **migration manifest** (same simple shape: `action · source · destination ·
   why · risk`), covering specs-repo creation if needed, per-artifact routing, and index /
   back-reference creation. Save it → `<project>/_support/workspace/migration-<YYYY-MM-DD>.md`;
   render an HTML review via flow-report when substantial and open it. Already-conforming → say so
   and skip.
3. Gate via AskUserQuestion grouped by RISK: specs-repo creation, deletes, and reference-breaking
   moves per item; safe moves/renames batch. Deletions run only after typed `eliminar` + IDs.
4. Execute only what was approved: `git init <project>-specs` (if approved) → skeleton the ledger,
   the ambient pair, and the specs-repo structure guided by judgment (not a fixed bootstrap
   sequence); `git mv` inside a repo / `mv` across the `_support/`→specs boundary; write promoted
   files, indexes, and back-references; archive (never delete) on doubt; preserve content verbatim
   and the raw/curated split. **Ambient refresh:** if the workspace AGENTS.md predates the
   flow-start template, rewrite that block to it (a coherent revision, not a string swap) and add
   the `sessions/**` pre-authorization to `.claude/settings.json` if absent.
5. CLOSE per the contract: update the ledger pointer (it does NOT index sessions — that stays in
   the co-located `sessions/README.md`), report executed vs skipped, and — if a specs repo was
   created — make its initial commit (this invocation authorizes it; no remote unless the user adds
   one). Re-running `migrate` is safe: already-conforming artifacts produce no findings.
