---
name: agents-md-primary
description: >
  Convert a project to the AGENTS.md-primary pattern: AGENTS.md becomes the single
  canonical harness-instructions file and CLAUDE.md becomes its `@AGENTS.md` import
  (plus genuinely Claude-specific content below the import). Use on projects where
  AGENTS.md and CLAUDE.md duplicate content, where only one of the pair exists, or to
  find conversion candidates across many projects (scan). Also audits project
  AGENTS.md/CLAUDE.md content against the global hive canon — per-rule harness-coverage
  matrix, dedup, stale-fork detection, promotion/injection routing — plus content quality:
  agent-discoverable rules, stale file paths, missing git-workflow declarations and ledger
  pointers, instruction budget (audit | apply). Idempotent.
argument-hint: "[path | scan <root> | audit [path] | apply]"
disable-model-invocation: true
---

# /agents-md-primary — one canonical file, every harness ambient

Why this pattern: Claude Code reads CLAUDE.md; Codex, opencode, and the rest of the
AGENTS-compatible ecosystem read AGENTS.md. Duplicating content across both guarantees
drift. The inversion makes AGENTS.md canonical and CLAUDE.md a one-line importer
(`@AGENTS.md` — official Claude Code import, loaded at session start), with
Claude-specific instructions below the import when they genuinely exist.

**The import law (empirical, Claude Code 2.1.233 — undocumented upstream):** `@AGENTS.md`
resolves in the cwd's own CLAUDE.md and in lazily-loaded subtree CLAUDE.md files; it does
NOT resolve in eagerly-loaded ANCESTOR CLAUDE.md files (their literal text loads, the
import is skipped). Consequence for `convert`: at a repo root the pattern is complete for
sessions opened there and for parent-workspace sessions that touch the repo (lazy path
resolves); at a WORKSPACE root the pattern's content operates only in sessions opened at
that root — warn this trade in the convert report when the target is a workspace root.

## `scan <root>` — find candidates (read-only)

Walk `<root>` (default `~/Development/projects`, depth ≤ 4 to cover
`projects/<group>/<project>/<repo>`) and classify every directory that has AGENTS.md
and/or CLAUDE.md at its root:

- `converted` — CLAUDE.md carries the `@AGENTS.md` import (or is a symlink to AGENTS.md).
  Detect the import ANYWHERE in the file, not just on line 1: a heading or preamble above it
  resolves identically. Matching only the first line misreports conformant repos as `both`.
- `duplicated` — both exist with substantially overlapping content ← the targets
- `divergent` — both exist with conflicting or deliberately different content
- `claude-only` / `agents-only` — half the pair missing

Report as a table with a suggested action per row. Change nothing.

## Convert (default; `$1` = project path, default cwd)

1. **Classify** the pair at the target (same buckets as scan). `converted` → say so and
   stop (idempotency). A CLAUDE.md symlink to AGENTS.md is already a valid form of the
   pattern — report it, don't rewrite it.
2. **Single-file cases (proceed directly — fully reversible):**
   - `claude-only` → move the content to AGENTS.md verbatim; CLAUDE.md becomes the
     import. Fix self-references ("this CLAUDE.md" → "this AGENTS.md").
   - `agents-only` → create CLAUDE.md containing exactly `@AGENTS.md`.
3. **Both-files case — merge analysis before touching anything.** Diff
   section-by-section into three buckets:
   - **Shared/duplicated** → lives once, in AGENTS.md.
   - **Harness-neutral but unique to either file** → AGENTS.md.
   - **Genuinely Claude-specific** (plan-mode behavior, Claude skills/hooks/subagent
     names, AskUserQuestion, model/effort settings) → CLAUDE.md, below the import.
     Test: would this line mislead a Codex/opencode session? Yes → it stays Claude-side.
   - **Contradictions** (both files rule differently on the same topic) → never pick
     silently; list each with both versions (divergence protocol) for the user to
     resolve in the confirmation step.
4. **Gate by signal, not by case.** A clean both-files merge (no contradictions, no
   divergent-by-design pair) → proceed-and-report like the single-file cases: fully
   reversible (uncommitted diff in a git repo, `.bak` files otherwise — step 6), and the
   summary diff is the review surface. Confirm only on signal: **contradictions** from
   step 3 (present both versions — divergence protocol), or a pair `divergent` by design
   (mostly disjoint scopes — e.g. AGENTS.md as a local compatibility guide, not a
   duplicate) → the pattern may not apply; stop unless the user overrides.
5. **Write:**
   - `AGENTS.md` — the canonical content. Preserve the original language and wording;
     translating or rewriting prose is out of scope for this skill.
   - `CLAUDE.md` — first line `@AGENTS.md`, then a blank line, then the Claude-specific
     block if any (under a `## Claude Code` heading).
6. **Verify, no silent loss:** every section of both originals must be accounted for —
   moved, kept Claude-side, or dropped with the reason stated in the report. Show a
   summary diff. If the project is a git repo, leave the changes uncommitted and suggest
   the commit; if it is NOT a git repo, write `CLAUDE.md.bak`/`AGENTS.md.bak` first.

## `audit [path]` — dedup project rules against the global canon (read-only)

Target: one project/workspace root (default cwd) and its child repos' `AGENTS.md`/
`CLAUDE.md`. **Coverage canon = the DEPLOYED layers — what sessions actually load:**
`~/.claude/CLAUDE.md` + always-on `~/.claude/rules/`, the condensed core at
`~/.codex/AGENTS.md` / `~/.config/opencode/AGENTS.md`, Grok's `~/.grok/rules/` symlinks,
the deployed router-skill references, and opencode's rules plugin. The hive repo — WHEN
locatable — serves two other roles only: flagging deployed-vs-source drift, and receiving
the promote-to-core / inject-to-router outcomes (those are hive edits). Not locatable →
those two outcomes are reported as destination-less proposals and the rest of the audit
runs unchanged.

**Per rule, compute the harness-coverage matrix — never a boolean, and with a LEVEL
axis.** Which DEPLOYED always-on layer already carries it: `~/.claude/rules` (Claude ✓,
Grok ✓ via symlink) · the condensed deployed core (Codex ✓, opencode ✓) · a router-skill
injection or the opencode rules plugin (situational reach). A project rule duplicating a
global-rules-only item is still LOAD-BEARING for Codex/opencode — deletion requires
coverage in every harness the project uses. The LEVEL of the audited file changes what
"coverage" means: a REPO-level file loads in child-repo sessions (all harnesses); a
WORKSPACE-ROOT file operates ONLY in sessions opened at the workspace root — from a
child-repo session no harness auto-loads it, and Claude Code's ancestor walk loads the
workspace CLAUDE.md but does NOT resolve its `@AGENTS.md` import (the import law below).

**Classify each rule into exactly one outcome:**

| Outcome | When | Proposed action |
|---|---|---|
| **delete** | Full coverage — pure noise (and reclaims Codex's per-repo `project_doc_max_bytes` budget) | Remove from the project file |
| **promote-to-core** | Universal (gate, every-session procedure), partial coverage | Condensed line into `harness/AGENTS.md` (27 KiB budget is the gate) → then delete from EVERY project |
| **inject-to-router** | Situational (language, workspace, memory policy), partial coverage | Add the owning global rule to `SKILL_REFERENCE_INJECTIONS` → delete from the project |
| **discoverable** | The repo's own files already state it — package manager (lockfile), scripts (`package.json`), framework (its config), directory inventory | Remove. Discriminator is the no-op test: delete the line and name what the agent would do differently. Nothing → it is a no-op. Verify against disk before proposing, never from the rule's wording |
| **stale** | An implementation detail that no longer matches the repo — distinct from `re-anchor`, which is drift against a GLOBAL rule | Resolve every backtick path against disk (below). Resolves elsewhere → **moved**: rewrite the path. Nowhere → **absent**: delete the claim |
| **keep — project canon** | Genuinely project-specific, needed cross-harness | Split by LEVEL: a REPO file stays — it is the canonical channel for all harnesses; a WORKSPACE-ROOT file operates only in workspace-root sessions — its cross-harness content is proposed DOWN to the governing repo (plus a per-repo pointer to ledger/workspace root). **Record why it is canon** — a ticket reference (`TRI-514`) is the preferred form; it holds the full reason outside the context budget. That note is what lets the NEXT audit delete the rule instead of re-deriving it |
| **re-anchor / explicit override** | Stale fork (paraphrased copy of an evolved global rule — the drift that produces contradictory instructions) or a deliberate contradiction | Rewrite quoting the canonical text, or as `overrides global <rule> because <reason>` — overrides legitimately WIN (`git-workflow.md` precedence); they must read as intentional |

**Resolving backtick paths (feeds the `stale` outcome).** A naive resolver is unusable — ~90% of
its hits are false. Keep only tokens that contain `/` AND end in a `.ext`, discarding anything
starting with `@ * / http ~` or containing `:// < > *` or spaces — that drops TS aliases, URL
routes, slash-commands, regexes and hostnames. Unresolved → retry by basename (`find -name`)
before reporting: most broken paths are MOVED, not absent, and the two take opposite fixes.

**Completeness checks — same audit pass, per child repo of a workspace:**

- **Project pointers**: the repo `AGENTS.md` carries `Ledger: <relative path>` and the
  workspace-root path, with the instruction to open them for project-scope tasks
  (child-repo sessions never auto-load the workspace files). Missing → propose the
  2-4 line block (~200 B).
- **`## Git Workflow` declarations**: `Base branch:` / `Git mode:` / `PR review:` per
  `git-workflow.md`'s declaration block. Each declared value removes a per-session question.
  `Base branch` is inferable from branch topology; `PR review` from the review apps wired to
  the repo. **`Git mode` is never inferable — it is the user's preference, not a repo fact:**
  propose one from the repo's class (PR-gated with CI → `pr-merge`; specs/mocks/docs
  single-branch → `direct-base`) and confirm it in the apply, never infer it silently.
- **Instruction budget**: report each file's size, and the repo's Codex chain — Codex
  concatenates git-root → cwd, so nested `AGENTS.md` files are charged together against
  `project_doc_max_bytes` (32 KiB default; raising it is per-machine and does not travel with
  the repo). Informational; the fleet's median file is ~2.3 KB, so flag only real outliers.
- **Tracker declaration**: a flow project declares in its LEDGER (the three fields —
  `memory-routing.md` owns the home and the confirm gate); the audit only FLAGS absence
  and proposes the ledger edit — never writes it, and the proposal rides the same
  confirm-gated batch.

Delegate the per-project reading to subagents (context hygiene; the semantic comparison
is judgment work — paraphrases count as duplicates). Output: a per-file table
(rule · current home · coverage CC/Grok/Codex/opencode · outcome · proposed diff), the
completeness proposals per repo, plus the promotion/injection candidates for the hive.
Audit changes nothing.

## `apply` — execute the confirmed audit manifest

One approval covers the batch; contradictions and overrides are listed individually
inside it. Project edits ride each repo's session git mode; hive changes
(promotions to the core, injection-map edits) are hive commits with `build.py` re-run.
Never apply without an audit manifest from this session.

**Shared surface:** `inject-to-router` WRITES `SKILL_REFERENCE_INJECTIONS`; `/manage-rules
validate` check 9 VALIDATES that same map's harness reachability. Two owners, one file —
after an injection edit, that check is the verification step, not an optional follow-up.

## Out of scope

- Nested `CLAUDE.md`/`AGENTS.md` in subdirectories — convert one root per invocation
  (point the skill at the subdirectory if needed).
- `CLAUDE.local.md` — personal file, never touched.
- Content quality in `convert`: it relocates, it does not rewrite or prune — pruning and
  dedup are `audit`/`apply`'s job.
- **Rewriting the user's prose, in any subcommand.** The skill proposes deleting, re-anchoring,
  or adding declarative blocks; it never recomposes, compresses, or restyles wording that stays.
- **Rationale as an inline comment.** Instruction files are tokenized verbatim by Codex,
  opencode and Grok (Claude Code strips HTML comments from `CLAUDE.md`; nothing strips them from
  `AGENTS.md`, which is where the canonical content lives). A rule's reason belongs in the
  ticket it cites, the commit body, or the hive's bibliography — never in the loaded file.
