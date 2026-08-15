---
name: agents-md-primary
description: >
  Convert a project to the AGENTS.md-primary pattern: AGENTS.md becomes the single
  canonical harness-instructions file and CLAUDE.md becomes its `@AGENTS.md` import
  (plus genuinely Claude-specific content below the import). Use on projects where
  AGENTS.md and CLAUDE.md duplicate content, where only one of the pair exists, or to
  find conversion candidates across many projects (scan). Also audits project
  AGENTS.md/CLAUDE.md content against the global hive canon — per-rule harness-coverage
  matrix, dedup, stale-fork detection, promotion/injection routing (audit | apply).
  Idempotent.
argument-hint: "[path | scan <root> | audit [path] | apply]"
disable-model-invocation: true
---

# /agents-md-primary — one canonical file, every harness ambient

Why this pattern: Claude Code reads CLAUDE.md; Codex, opencode, and the rest of the
AGENTS-compatible ecosystem read AGENTS.md. Duplicating content across both guarantees
drift. The inversion makes AGENTS.md canonical and CLAUDE.md a one-line importer
(`@AGENTS.md` — official Claude Code import, loaded at session start), with
Claude-specific instructions below the import when they genuinely exist.

## `scan <root>` — find candidates (read-only)

Walk `<root>` (default `~/Development/projects`, depth ≤ 4 to cover
`projects/<group>/<project>/<repo>`) and classify every directory that has AGENTS.md
and/or CLAUDE.md at its root:

- `converted` — CLAUDE.md is the `@AGENTS.md` import (or a symlink to AGENTS.md)
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

**Per rule, compute the harness-coverage matrix — never a boolean.** Which DEPLOYED
always-on layer already carries it: `~/.claude/rules` (Claude ✓, Grok ✓ via symlink) ·
the condensed deployed core (Codex ✓, opencode ✓) · a router-skill injection or the
opencode rules plugin (situational reach). A project rule duplicating a global-rules-only
item is still LOAD-BEARING for Codex/opencode — deletion requires coverage in every
harness the project uses.

**Classify each rule into exactly one outcome:**

| Outcome | When | Proposed action |
|---|---|---|
| **delete** | Full coverage — pure noise (and reclaims Codex's per-repo `project_doc_max_bytes` budget) | Remove from the project file |
| **promote-to-core** | Universal (gate, every-session procedure), partial coverage | Condensed line into `harness/AGENTS.md` (27 KiB budget is the gate) → then delete from EVERY project |
| **inject-to-router** | Situational (language, workspace, memory policy), partial coverage | Add the owning global rule to `SKILL_REFERENCE_INJECTIONS` → delete from the project |
| **keep — project canon** | Genuinely project-specific, needed cross-harness | Stays: the project AGENTS.md IS its canonical channel |
| **re-anchor / explicit override** | Stale fork (paraphrased copy of an evolved global rule — the drift that produces contradictory instructions) or a deliberate contradiction | Rewrite quoting the canonical text, or as `overrides global <rule> because <reason>` — overrides legitimately WIN (`git-workflow.md` precedence); they must read as intentional |

Delegate the per-project reading to subagents (context hygiene; the semantic comparison
is judgment work — paraphrases count as duplicates). Output: a per-file table
(rule · current home · coverage CC/Grok/Codex/opencode · outcome · proposed diff) plus
the promotion/injection candidates for the hive. Audit changes nothing.

## `apply` — execute the confirmed audit manifest

One approval covers the batch; contradictions and overrides are listed individually
inside it. Project edits ride each repo's session git mode; hive changes
(promotions to the core, injection-map edits) are hive commits with `build.py` re-run.
Never apply without an audit manifest from this session.

## Out of scope

- Nested `CLAUDE.md`/`AGENTS.md` in subdirectories — convert one root per invocation
  (point the skill at the subdirectory if needed).
- `CLAUDE.local.md` — personal file, never touched.
- Content quality in `convert`: it relocates, it does not rewrite or prune — pruning and
  dedup are `audit`/`apply`'s job.
