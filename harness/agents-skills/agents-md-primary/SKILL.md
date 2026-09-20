---
name: agents-md-primary
description: >
  Convert a project to the AGENTS.md-primary pattern: AGENTS.md becomes the single canonical
  harness-instructions file and CLAUDE.md becomes its `@AGENTS.md` import (plus genuinely
  Claude-specific content below the import). Use on projects where AGENTS.md and CLAUDE.md
  duplicate content, where only one of the pair exists, or to find conversion candidates across
  many projects (scan). Also audits a project's AGENTS.md/CLAUDE.md against the deployed global
  canon — harness coverage, duplication, stale forks, content quality, and instruction budget —
  and checks every declared claim against git and disk rather than taking the document's word for
  it. Subcommands: audit | apply. Idempotent.
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
- `blind` — the half that is missing makes the root invisible to a harness in use:
  `AGENTS.md` with no sibling `CLAUDE.md` (Claude Code does NOT fall back), or `CLAUDE.md`
  with no `AGENTS.md` (Codex/opencode/Grok). Count only REPO and WORKSPACE roots — a
  subtree file, `_support/backup/**`, and a deployable source tree (`global/`, `harness/`)
  are not findings.
- `unmanaged` — an active root (any child repo with a commit < 90 days) carrying NO
  instruction file at all. `convert` has nothing to trigger on here, so this is the only
  subcommand that ever finds it; the conventions usually exist in `docs/CONTRIBUTING.md`
  or `README.md`, which no harness loads. Propose promoting that file, never a stub.

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

Audits a project's `AGENTS.md`/`CLAUDE.md` against the DEPLOYED global canon and writes a
manifest to disk for `apply` to execute. Read-only; it never touches the hub it was
deployed from.

**Read [references/audit.md](references/audit.md) before running it** — it carries the
scope (child repos plus every nested `AGENTS.md`), the harness-coverage matrix and its
LEVEL axis, the admission test, the five outcomes and their evidence, the backtick-path
resolver, the completeness and ground-truth checks, and the manifest's shape, which is
`apply`'s input.

## `apply` — execute the confirmed audit manifest

One approval covers the batch; contradictions and overrides are listed individually
inside it. Project edits ride each repo's session git mode. A hook-config edit proposed by
**verification parity** rides that same approval — it is the one non-instruction file this
skill writes; verify it with the manager's own runner (`lefthook run pre-push`) before
reporting it done. Hub-destined proposals
(promote-to-core, injection-map edits) are OUT of apply's scope — the manifest carries
them for the user to execute in a session opened in the hub, where `build.py` and its
validation run. **Read the manifest from disk** (`> audit` writes it), so a run days later needs no
re-audit. **Revalidate before writing, entry by entry:** every path still resolves and every
quoted line still matches. A tree that moved underneath — a migration, an earlier partial
apply, someone else's commit — invalidates the entries it touched; those are re-audited, never
applied blind, and the report says which ones went stale. No manifest on disk and none in this
session → run `audit` first; nothing left to apply → say so rather than re-deriving one.

## Out of scope

- Nested `CLAUDE.md`/`AGENTS.md` in subdirectories — convert one root per invocation
  (point the skill at the subdirectory if needed).
- **Writing `CLAUDE.local.md`** — personal file, never written by any subcommand. `audit`
  READS it (above): excluding it from reading is what let a false git mode, a false merge
  strategy and a dead stack survive in three separate workspaces.
- Content quality in `convert`: it relocates, it does not rewrite or prune — pruning and
  dedup are `audit`/`apply`'s job.
- **Rewriting the user's prose, in any subcommand.** The skill proposes deleting, re-anchoring,
  or adding declarative blocks; it never recomposes, compresses, or restyles wording that stays.
- **Rationale as an inline comment.** Instruction files are tokenized verbatim by Codex,
  opencode and Grok (Claude Code strips HTML comments from `CLAUDE.md`; nothing strips them from
  `AGENTS.md`, which is where the canonical content lives). A rule's reason belongs in the
  ticket it cites or the commit body — never in the loaded file.
