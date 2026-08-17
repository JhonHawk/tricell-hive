---
name: engram-init-workspace
description: >
  Init unified Engram .engram/config.json for multi-repo workspaces (shared memory
  bucket). Use on ambiguous project detection, new projects/<group>/<project>/, or
  explicit Engram init/register.
argument-hint: "[project-name]"
allowed-tools: Read, Glob, Bash(bash *bootstrap-workspace.sh *)
---

# Engram: initialize a unified workspace project

Engram keys memory by a detected project name. Single-repo projects need no setup — git-remote detection already gives every tool the same key. In a multi-repo workspace (`projects/<group>/<project>/` with N git repos inside), detection breaks two ways: at the workspace root (not a git repo) it returns `ambiguous`, and `mem_context`/`mem_search` without an explicit `project` fail outright; inside a child repo the git remote resolves a per-repo name (`chat-hub-backend`, `chat-hub-frontend`…), fragmenting memory across buckets that never see each other.

The fix is ONE unified project for the whole workspace: `.engram/config.json` = `{ "project_name": "<group>-<project>" }` (e.g. `acme-chat-hub`). Engram reads the config as case 0 — before any git logic — but only within the enclosing git boundary; it is **never inherited downward**, so placement matters: the config at the **workspace root** covers tools launched there (Claude Code), and **each child repo** needs its own copy with the same `project_name` for tools launched inside one repo (Codex/opencode). This also aligns Engram's bucket with Claude Code's native per-workspace memory (one bucket per workspace). This skill writes both placements idempotently; the routing boundary between Engram and native memory is `rules-situational/memory-routing.md`.

## Steps

1. **Detect — write nothing yet.** The script decides what this directory is; don't pre-judge.
   ```bash
   bash "${CLAUDE_SKILL_DIR}/bootstrap-workspace.sh" --detect
   ```
   (Append a directory argument to target somewhere other than the current directory.) Read `verdict=` and `location=`:
   - **`workspace`** — canonical layout; `project_name`/`workspace_root` resolved. `location` sets the default write scope (see the Scope table). Go to step 2.
   - **`needs-name`** — multi-repo dir outside `projects/<group>/<project>`; the name can't be derived. Use `$ARGUMENTS` if given, else ask for `<group>-<project>`. Go to step 2 with `--name <name>`.
   - **`not-workspace`** — **stop.** Report `reason=` (e.g. a standalone repo needs no config — git-remote already unifies). Only proceed if the user explicitly wants a custom name for *this one repo*: re-run with `--name <name>` (writes that repo only — verdict becomes `register-repo`, never touching siblings or the parent).

2. **Write — no confirmation round.** The script is idempotent, never clobbers, and the config is an untracked, reversible file — proceed directly; the report carries the outcome. The only ask is `needs-name` with no `$ARGUMENTS` (genuine input — step 1). Re-run without `--detect`, adding scope flags as needed:
   ```bash
   bash "${CLAUDE_SKILL_DIR}/bootstrap-workspace.sh" [--name N] [--all] [--only a,b] [--exclude c] [dir]
   ```
   Refuses unless detection is `workspace` or `register-repo`.

3. **Report** `project_name`, `workspace_root`, and created vs. skipped configs; have the user verify from the workspace root that `mem_current_project` returns the name with `source: config`. Fold in the gitignore tip (once): suggest adding `.engram/` to the global excludes file (`git config --global core.excludesfile`, typically `~/.config/git/ignore`) — unless the team also uses Engram and wants the identity committed.

## Scope — what gets written

The default depends on where detection ran (`location`); flags override it.

| Situation (`location`) | Default scope | Override flags |
|---|---|---|
| At the workspace root (`workspace_root`) | root + **all** child repos | `--only a,b` / `--exclude c` to filter |
| Inside a child repo (`child_repo`) | root + **only that repo** | `--all` for the whole workspace; `--only`/`--exclude` to pick |
| Standalone repo, non-canonical parent (`standalone_repo`) | nothing — refuses | `--name N` registers **only this repo**, never the parent or siblings |

The root config is always (re)ensured for canonical workspaces — it's what a tool launched at the workspace root reads.

## Notes

- **Does not migrate existing memory.** It declares identity going forward. Observations already saved under fragmented per-repo names stay there — they are not moved into the unified bucket. Moot for a new or empty workspace.
- **Single-repo projects need no config** — git-remote detection already gives every tool the same key. This skill targets multi-repo workspaces.
