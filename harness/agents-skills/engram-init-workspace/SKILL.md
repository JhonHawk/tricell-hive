---
name: engram-init-workspace
description: Initialize a unified Engram project for a multi-repo workspace by creating .engram/config.json at the workspace root and each child git repo, so every repo shares one memory bucket. Use when Engram returns ambiguous project detection, when setting up a new workspace under projects/group/project/, or when the user asks to register or init the Engram project for a workspace.
---

# Engram: initialize a unified workspace project

Engram keys memory by a detected project name. In a multi-repo workspace (`projects/<group>/<project>/` with N git repos inside), auto-detection fragments per repo (each repo's git remote → a different name) and returns `ambiguous` at the workspace root. This skill declares ONE unified project for the whole workspace via `.engram/config.json`, which Engram reads as case 0 — before any git logic. Background: `rules/workflow/memory-routing.md > Workspace project identity`.

## Steps

1. **Detect — write nothing yet.** The script decides what this directory is; don't pre-judge.
   ```bash
   bash "~/.agents/skills/engram-init-workspace/bootstrap-workspace.sh" --detect
   ```
   (Append a directory argument to target somewhere other than the current directory.) Read `verdict=` and `location=`:
   - **`workspace`** — canonical layout; `project_name`/`workspace_root` resolved. `location` sets the default write scope (see the Scope table). Go to step 2.
   - **`needs-name`** — multi-repo dir outside `projects/<group>/<project>`; the name can't be derived. Use `$ARGUMENTS` if given, else ask for `<group>-<project>`. Go to step 2 with `--name <name>`.
   - **`not-workspace`** — **stop.** Report `reason=` (e.g. a standalone repo needs no config — git-remote already unifies). Only proceed if the user explicitly wants a custom name for *this one repo*: re-run with `--name <name>` (writes that repo only — verdict becomes `register-repo`, never touching siblings or the parent).

2. **Confirm** (skip only when `$ARGUMENTS` gave an explicit name): show `project_name`, `workspace_root`, `current_repo` (if any), and which repos the scope will write. Proceed on confirmation.

3. **Write** — re-run without `--detect`, adding scope flags as needed:
   ```bash
   bash "~/.agents/skills/engram-init-workspace/bootstrap-workspace.sh" [--name N] [--all] [--only a,b] [--exclude c] [dir]
   ```
   Idempotent: skips existing configs, never clobbers; refuses unless detection is `workspace` or `register-repo`.

4. **Report** created vs. skipped, then have the user verify from the workspace root that `mem_current_project` returns the name with `source: config`.

## Scope — what gets written

The default depends on where detection ran (`location`); flags override it.

| Situation (`location`) | Default scope | Override flags |
|---|---|---|
| At the workspace root (`workspace_root`) | root + **all** child repos | `--only a,b` / `--exclude c` to filter |
| Inside a child repo (`child_repo`) | root + **only that repo** | `--all` for the whole workspace; `--only`/`--exclude` to pick |
| Standalone repo, non-canonical parent (`standalone_repo`) | nothing — refuses | `--name N` registers **only this repo**, never the parent or siblings |

The root config is always (re)ensured for canonical workspaces — it's what a tool launched at the workspace root reads.

5. **Offer the gitignore tip (once).** The per-repo `.engram/config.json` is untracked in each repo. Suggest adding `.engram/` to the global git ignore file (the one referenced by `git config --global core.excludesfile`, typically `~/.config/git/ignore`) so it never pollutes any repo — unless the team also uses Engram and wants the identity committed.

## Notes

- **Does not migrate existing memory.** It declares identity going forward. Observations already saved under fragmented per-repo names stay there — they are not moved into the unified bucket. Moot for a new or empty workspace.
- **Single-repo projects need no config** — git-remote detection already gives every tool the same key. This skill targets multi-repo workspaces.
