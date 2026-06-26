---
name: agents-md-primary
description: >
  Convert a project to the AGENTS.md-primary pattern: AGENTS.md becomes the single
  canonical harness-instructions file and CLAUDE.md becomes its `@AGENTS.md` import
  (plus genuinely Claude-specific content below the import). Use on projects where
  AGENTS.md and CLAUDE.md duplicate content, where only one of the pair exists, or to
  find conversion candidates across many projects (scan). Idempotent.
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
4. **Gate (both-files case only):** present the conversion plan — what moves where,
   what stays Claude-side, contradictions to resolve — and get confirmation. If the pair
   is `divergent` by design (mostly disjoint scopes — e.g. a repo where AGENTS.md is a
   local compatibility guide, not a duplicate), say the pattern may not apply and stop
   unless the user overrides.
5. **Write:**
   - `AGENTS.md` — the canonical content. Preserve the original language and wording;
     translating or rewriting prose is out of scope for this skill.
   - `CLAUDE.md` — first line `@AGENTS.md`, then a blank line, then the Claude-specific
     block if any (under a `## Claude Code` heading).
6. **Verify, no silent loss:** every section of both originals must be accounted for —
   moved, kept Claude-side, or dropped with the reason stated in the report. Show a
   summary diff. If the project is a git repo, leave the changes uncommitted and suggest
   the commit; if it is NOT a git repo, write `CLAUDE.md.bak`/`AGENTS.md.bak` first.

## Out of scope

- Nested `CLAUDE.md`/`AGENTS.md` in subdirectories — convert one root per invocation
  (point the skill at the subdirectory if needed).
- `CLAUDE.local.md` — personal file, never touched.
- Content quality: this skill relocates, it does not rewrite or prune.
