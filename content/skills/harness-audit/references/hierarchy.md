# Instruction hierarchy

Read this before judging any instruction file. A file's value depends on which hosts load it, from which entry points, and alongside what.

## Compute the effective set (HA-HI-01)

For each host in use and each declared or observed entry point, list:

1. **Startup set**: files loaded when the session starts, in order, with sizes.
2. **On-demand set**: files a host adds later when the agent opens files in a subdirectory.
3. **Configuration inputs** that change either set.

Judge a level only against these sets. When a conclusion depends on version-sensitive behavior, confirm it with the host's own view or report it as a hypothesis (HA-ME-02; see [native tools](native-tools.md)).

## Observe entry points

Count sessions per start directory for the targets and their ancestors, recording the period covered:

- Claude Code: one `*.jsonl` per session under `<Claude home>/projects/<start path with separators replaced by ->/`.
- Codex: the `cwd` of the first `session_meta` record in each file under `~/.codex/sessions/`.
- Grok Build: one directory per session under `~/.grok/sessions/<URL-encoded start path>/`. The first user record of its `chat_history.jsonl` holds the `<rules>` actually loaded, which is direct evidence of the startup set.

Histories show where sessions started, not what the user intends. Report both when they differ, and read only metadata, never conversation content beyond the loaded rules.

## Loading behavior

Checked on Claude Code 2.1.280 and Codex 0.155.1. Reconfirm when the installed version differs.

| Aspect | Claude Code | Codex |
| --- | --- | --- |
| Startup discovery | Managed-policy `CLAUDE.md`, the user file `<Claude home>/CLAUDE.md` and user rules, then `CLAUDE.md`, `.claude/CLAUDE.md`, and `CLAUDE.local.md` in the entry directory and every ancestor, plus `.claude/rules/` files without `paths`; `@path` imports resolve, including in ancestors; `AGENTS.md` joins only as described below | Global file, then each directory from the git root (or the entry directory when there is no git root) down to the entry directory: `AGENTS.override.md`, else `AGENTS.md`, else configured fallback names |
| On demand | `CLAUDE.md` and `CLAUDE.local.md` in a subdirectory, and what they import, when the agent opens a file there; a subdirectory `AGENTS.md` only when `AGENTS.md` is read natively; rules whose `paths` match the opened file | None |
| `AGENTS.md` | Read natively only when no `CLAUDE.md`, `.claude/CLAUDE.md`, or `CLAUDE.local.md` exists at or above the entry point (default setting) | Native file |
| Order | Root to entry directory; closer files read last | Root to entry directory; closer files read last |
| Size | Target under 200 lines per file; files over 4 MiB are skipped | Project chain stops silently at `project_doc_max_bytes` (32 KiB default); the global file does not count |

Grok Build 1.0.38, observed in two sessions on 2026-09-22 and not documented: the startup set held `<Claude home>/CLAUDE.md` plus the entry directory's `AGENTS.md` and `CLAUDE.md`; a workspace `AGENTS.md` one level above a repository entry point was not loaded. Confirm against the session's recorded `<rules>` before relying on it.

Configuration inputs to read when present:

- Claude Code: the *Project instructions* setting (`claude-md-or-agents-md` default, `claude-md-and-agents-md`, `claude-md`, `managed-only`), `claudeMdExcludes`, `.claude/CLAUDE.md`, `.claude/rules/` (rules with `paths` load when matching files are opened), user rules under the Claude home, and sessions where `AGENTS.md` support is unavailable.
- Codex: `project_doc_max_bytes`, `project_doc_fallback_filenames`, and any `AGENTS.override.md` in the global or project chain.

## Rules

- **HA-HI-02 — Loading claims.** A file's statement about what loads (for example, "the workspace file is not auto-loaded from here") must match every host in use. A mismatch is `stale`; when hosts differ, rewrite the statement per host.
- **HA-HI-03 — `AGENTS.md` suppression.** When any `CLAUDE.md`, `.claude/CLAUDE.md`, or `CLAUDE.local.md` exists at or above an entry point, Claude Code skips `AGENTS.md` by default. Every `AGENTS.md` that must reach Claude Code needs an `@AGENTS.md` import or a symlink from its own level's `CLAUDE.md`. Evaluate ancestors too: a workspace `CLAUDE.md` suppresses a repository `AGENTS.md` that has no sibling `CLAUDE.md`.
- **HA-HI-04 — `CLAUDE.md` content.** `CLAUDE.md` imports `@AGENTS.md` (or is a symlink to it) and may add a block that applies only to Claude Code. Host-neutral content below the import is `relocate` to `AGENTS.md`; a note with no function for the agent, such as advice to maintainers about where to put rules, is `delete`; an import-only `CLAUDE.md` is `keep`, since it duplicates nothing and covers sessions without native `AGENTS.md` support. A lone `AGENTS.md` is valid only when no `CLAUDE.md`, `.claude/CLAUDE.md`, or `CLAUDE.local.md` exists at or above the entry point; a private `CLAUDE.local.md` alone suppresses it, so propose a `CLAUDE.md` that imports it.
- **HA-HI-05 — Placement by level.** What a repository session needs to operate lives in that repository: Codex never loads a workspace file from a child repository, even when Claude Code does. Only content that spans repositories stays above. Move content with `relocate`, never by copying, and give each repository a one-line upward pointer when its sessions need the level above.
- **HA-HI-06 — Package files.** A package file inside a monorepo covers only its own scope and never copies the repository file. Codex reads it only when the session starts inside the package. Claude Code reads it on demand only through that package's `CLAUDE.md` import, or natively when no `CLAUDE.md`-family file exists at or above (HA-HI-03); without either path it never reaches Claude Code. A package gotcha needed by sessions started at the repository root requires a conditional pointer in the root file or `relocate`. Never propose deleting a package file that has content of its own as redundant.
- **HA-HI-07 — Pure pointers.** A file that only points to a level every target host already loads from every declared entry point is `low` redundancy. Keep it when sessions may start where the pointed file does not load.

## Reading layered findings

- Content above a repository that reaches it through the repository's upward pointer is conditionally reached, not missing.
- Report the same rule once, at the level that should own it, citing the other copies.
- Measure cost per entry point: the global layer and ancestor files are paid in every session that loads them, so a small file in a heavily used chain can matter more than a large one used rarely.
