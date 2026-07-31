---
name: memory-sync
description: >
  Audit/apply Engram+native memory vs ground truth when memory resurfaces finished work
  as pending, contradicts live code, at session close, or on request. Not for one-off
  "what's pending?" (inline ground-truth check). Subcommands: audit | apply.
---

# /memory-sync — memory reconciliation

Two memory layers run at once: Engram (MCP) and native file-memory
(`~/.claude/projects/<slug>/memory/`). They drift when state advances and the old fact is
never invalidated (`memory-routing.md > Invalidation` — "corrected by addition" leaves
both facts alive); this skill reconciles both layers against ground truth. Default
subcommand: `audit`.

**Ground-truth precedence** (canon: `quality/debugging.md > Reporting state from ground
truth`) — a memory is a claim verified *against* these, never trusted *over* them:
1. **The live system** — matched to the claim type: git/disk for implementation, the
   running app/DB for runtime state, context7 anchored to the installed version for
   tool/library behavior claims. **Absence-of-evidence ≠ proof of absence:** an empty
   inspection result (`pnpm config get`, a `grep` miss) does not prove "off" — confirm a
   negative against the authoritative source.
2. **The ledger** (and a *declared* tracker — see the tracker step) — status records:
   above memory, below the live system; when they disagree with git/disk, live state
   wins and the record is what to correct.
3. **Memory** — Engram + native: the *lowest* authority, a claim to confirm.

**Existence ≠ completion (the core trap).** A remembered "X is MISSING" is refuted by X simply *existing* — directly verifiable, safe to auto-invalidate. But a remembered "X is PENDING / TODO" is **not** refuted by a related artifact existing — 30 test files existing does not prove the testing task is done (they could be stubs, cover other cases, be unrelated). A pending-task claim needs a **positive completion signal**: a commit/PR that closes it, a ledger phase that has passed it, a tracker ticket marked done, or tests that actually *pass* — not just a file on disk. Existence-only → KEEP-and-flag for review, never auto-invalidate.

## Resolve the project first

- **Engram project:** `mem_current_project`, or `.engram/config.json` → `project_name` (multi-repo workspace → the unified `<group>-<project>`). Ambiguous → ask, never guess (`memory-routing.md > Workspace project identity`).
- **Native memory dir:** `~/.claude/projects/<slug>/memory/`, `<slug>` = workspace absolute path with `/` → `-`.
- **Ledger:** `<workspace>/_support/PROJECT.md` (absent in non-flow projects → ground truth is live state only).
- **Task tracker:** check whether the project *declares* one — in its `AGENTS.md`/`CLAUDE.md` or the ledger's `Tracker` field. Declared → reconcile uses it without asking each run. In use but **undeclared** (signals: ticket keys in commits/branches like `FAC-48`/`TRI-229`, a connected tracker MCP) → flag it in `audit` and propose adding the declaration to the project's `AGENTS.md`/`CLAUDE.md` (confirm before writing); once declared, later runs use it automatically — no per-run request. No tracker → skip.
- **Preflight tools:** Engram lifecycle tools are deferred — load before first use: `ToolSearch("select:mem_update,mem_compare,mem_delete,mem_review")`. (`mem_context`/`mem_search` are core.)

## `audit`

1. **Gather memory (read-only).**
   - Engram: `mem_context` (project) for the recent picture — do NOT rely on `mem_search` alone, its recall is unreliable (verified). Pull status / pending / decision observations with their ids.
   - Native: read `memory/MEMORY.md` and each file. Flag every "Remaining Work" / pending entry — **including struck-through or "DESACTUALIZADO"-annotated ones left inside a pending section** (annotated-in-place still reads as pending — the exact anti-pattern).
2. **Establish ground truth from the live system FIRST, matched to the claim type.** Never trust a *record* (ledger, tracker, Engram, native) as proof — records corroborate; the authoritative live source proves. Pick the source by claim (and **decompose a compound memory** — an upgrade *and* a resulting default — verifying each part against its own source, or the whole thing routes to git and the behavior half never reaches context7):
   - **Implementation claim** ("X was built / exists / is merged"): git is mandatory and primary — `git log --oneline --all` (grep area/ticket), `git show <sha>`, `git branch --merged` to confirm a closing commit/PR actually landed — plus `ls`/`grep` and a test run for presence/passing.
   - **Tool/library behavior or default claim** ("the cooldown is active", "the default is Z", anything version-sensitive): verify against the tool's authoritative docs for the *installed* version (context7, per `tools/context7.md`) — never against another memory or a misread inspection command.
   - **Task-state claim** ("done/pending"): the ledger and a *declared* tracker corroborate, but git proves.
   Apply existence-vs-completion (a "pending" claim needs a positive completion signal — a closing commit/PR or passing tests, not a related file existing) AND absence-of-evidence (an empty/`undefined` inspection result is not proof of "off"). Run checks in the main thread (it holds the Engram session); delegate only a wide multi-repo file sweep to an `Explore` subagent.
3. **Classify each claim with a stable ID** (IDs live in the manifest — `apply` may run later):
   - `S1…` **STALE** — *and auto-invalidable*: either "memory says MISSING / live shows it EXISTS", or "memory says PENDING / a positive completion signal proves it DONE". Record the exact evidence.
   - `F1…` **FLAG** — looks stale but only existence (not completion) could be shown, or could not be verified at all. KEEP, surface for user review — never auto.
   - `D1…` **DIVERGENT** — two memories / the two layers contradict; FLAG unless one side is positively proven by live state.
   - `O1…` **ORPHAN** — references a file/branch/ticket that no longer exists. A dangling ref does **not** make the remembered claim false (a merged-and-deleted branch leaves a true decision); auto-action is at most to fix the *reference*, never invalidate the observation.
   - `K1…` **KEEP** — verified still-true.
   For each non-KEEP: cite the memory (Engram `#id` / native `file:line`), the **positive** ground-truth evidence (command + observed result), and the proposed fix.
4. **Save the actions manifest** → `<project>/_support/workspace/memory-audit-<YYYY-MM-DD>.md`: one row per finding — `ID | class | layer | ref | claim | evidence (command + result) | proposed fix | risk (safe-update / destructive)`. The evidence cell MUST be a positive verification, not an absence; a row without it is `F` (flag), not `S`. In chat: counts per class + ID list + path. Clean → say so, skip the artifact.

## `apply`

1. **Load the latest manifest.** None → run `audit` first. **The manifest is a proposal, not a license:** re-confirm each `S`-row's evidence against CURRENT live state immediately before executing it (ground truth may have changed since the audit). A row whose evidence is now absent, "not found", "ambiguous", or only an absence → demote to FLAG, do not auto-apply.
2. **Auto path (safe — standing authorization): only `S` rows with re-confirmed positive evidence.**
   - **Engram:** `mem_update(id, content, [topic_key])` — rewrite the stale observation to *restate* the current truth (don't blank it). This is what actually stops it resurfacing; note mem_update overwrites irreversibly, which is why it is auto only under positive completion proof. Set a deterministic `topic_key` (`status/<area>`) so future saves UPSERT. Optionally also `mem_compare(memory_id_a=<new/correct id>, memory_id_b=<stale id>, relation:"supersedes", confidence, reasoning)` to *record the relation* — but it does NOT hide the stale memory on its own, so it never replaces the mem_update.
   - **Native:** EDIT `MEMORY.md` — MOVE the entry out of "Remaining Work" (to Done) or rewrite in place, and update its index line. Move/rewrite is auto; see gate below for removal.
   - **ORPHAN:** fix only the dangling reference; leave the claim.
3. **Gated path (destructive — typed confirmation `eliminar` + IDs):** `mem_delete` of an observation; deleting a native memory file, **removing** a native entry outright, or **emptying** a section (edit-as-delete counts as delete). `F`/`D` rows that the user reviews and approves also run here. Hesitation → keep/supersede, never delete.
4. **Tracker sync (when a tracker is declared).** A declared tracker is validated/read automatically — no per-run request. Propagate writes: a confirmed-stale "pending" mapped to an open ticket means the ticket is out of date — move/close it so tracker, ledger, and memory agree; the write is outward-facing, so confirm unless the project's config/flow authorizes tracker writes. Undeclared tracker → don't sync; propose declaring it first (see Resolve). Use the tracker MCP/CLI (Linear MCP, `acli`) via tool-search.
5. **Report** executed vs skipped, delete the consumed manifest, and update the ledger row if a finding revealed phase/decision drift.

## Notes

- **Does NOT configure Engram** (the plugin self-manages). It *uses* the lifecycle tools — `mem_update` / `mem_delete` (invalidate/remove), `mem_review action=list` (harvest observations whose decay window passed, an extra staleness signal). `mem_compare` is NOT a lifecycle tool — it only records a supersedes/conflict relation and does not by itself remove or hide a memory.
- **Why audit runs in the main thread** (unlike flow-hygiene's custodian): reconciliation needs the live Engram session the main thread holds; a fresh subagent would lack it. Only the wide disk sweep is delegable.
- Background and the upstream protocol: `memory-routing.md > Invalidation` (via the `memory-policy` skill) and `quality/debugging.md > Reporting state from ground truth`.
