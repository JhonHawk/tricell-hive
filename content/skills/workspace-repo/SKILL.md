---
name: workspace-repo
description: Turn a declared workspace of several repositories into a coordinator repository that versions the workspace's guidance, its repository table and its knowledge, while each child repository stays an independent clone inside it, or clone the repositories a coordinator lists. Use only when the user asks to make a workspace a repository, to create a workspace or coordinator repository, or to set one up after cloning it; not to check or organize a project's layout, which project-layout does, nor to merge repositories into a monorepo.
---

# Turn a workspace into a repository

Git is initialized at a workspace root only through this skill and only on the user's request, as the global guidance's «Support folder» sets. The coordinator versions what the whole workspace shares: its `AGENTS.md` with the `## Hive` section and a `## Repositories` table, the knowledge home `_support/docs` with its index, and the change records in `_support/openspec`. Each child repository keeps its own history, branches and complete `## Hive` section.

When the user asks to set up a coordinator they already cloned, run `python3 -I <workspace-repo-dir>/scripts/workspace_repo.py setup --root <root>` ([workspace repo](scripts/workspace_repo.py)), which clones each repository of the table that is missing, report what it cloned, and stop.

1. **Assess, changing nothing.** Confirm with the user the workspace root and its repositories when no `AGENTS.md` declares them, never inferring them from a parent folder. For each child repository record its path, its `origin` remote and its current branch, and which one holds `Specs`. Record what the root already holds: `AGENTS.md`, `CLAUDE.md`, `HIVE.local.md`, `.engram/`, `_support/`. Read [migrate](../project-layout/references/migrate.md) and apply its viability test: when the repositories pass it, recommend a monorepo instead. Run the check of [project-layout](../project-layout/SKILL.md) on the root and on each child, and run `workspace_repo.py plan` with each `--repo <path>=<remote>`; it prints what apply would write and writes nothing.
2. **Ask one round** with the assessment, as [decision format](../hive-core/references/questions.md) sets:
   - whether to create the coordinator, or the monorepo when the viability test passes;
   - its tracker, proposing the one its children share, since `hive doctor` requires it in the coordinator's `## Hive`;
   - whether its `## Hive` declares `Delivery: direct-base`, recommended so that knowledge and change records are committed straight to the coordinator's base; say it is the only value that authorizes those commits without asking;
   - what happens to a specs repository: kept as a child, retired once project-layout's migration moves its current knowledge into the coordinator, or promoted to be the coordinator while keeping its history, which is manual, as **Promote a specs repository** below sets;
   - which findings of the check project-layout organizes in this work, recommending those that move current knowledge into `_support/docs` and `_support/openspec`.
3. **Apply.** Run `workspace_repo.py apply` with the repositories, `--tracker` and, only when the user chose it, `--direct-base`. It writes a `.gitignore` that ignores each child, `HIVE.local.md`, `.engram/` and `_support/{workspace,data,scripts,secrets}`, a `.ignore` that re-includes each child so a search from the root finds their files, the `## Hive` and `## Repositories` sections of `AGENTS.md`, preserving its other text, `_support/openspec/` and the index `_support/docs/README.md`, then runs `git init`. It never creates a remote, pushes or commits. Commit its files through [git_state.py](../git-workflow/scripts/git_state.py) `commit` with explicit paths, then run project-layout's organize steps for the findings the user chose, with the coordinator as the destination of the knowledge they move.
4. **Verify.** The check of project-layout on the root exits 0, or reports only failures of findings the user left out of this work in step 2; a word that appears only in a child is found by a search from the root; `git -C <root> status --short` lists no child content and no `HIVE.local.md`; `hive doctor` at the root reports no blocking finding.
5. **Report** what was created, the commit, and the next steps that are the user's: creating the remote and pushing. Before the user adds `origin`, say that the viewer then identifies the workspace by its remote instead of its path, so a viewer work already open for it is refused as a changed project until it is closed: close or finish those works first.

## Promote a specs repository

When the user chooses to make an existing specs repository the coordinator, keeping its history: tag its current commit, move its current content under the coordinator's paths (`_support/openspec`, `_support/docs`) with `git mv` and commit it inside the specs repository, then move that repository's `.git` and tree to the workspace root, outside which the children already sit, and run step 3 with the files apply would write merged into the moved ones by hand. Propose each command and run it only after the user's answer; never do this on a repository with uncommitted work or on a shared branch.

## Limits

- Codex, Grok and Cursor do not load the coordinator's `AGENTS.md` in a session started inside a child, so each child keeps its complete `## Hive` section; sessions that work across repositories start at the coordinator root.
- Most hosts do not discover a child's skills from the coordinator root, so the coordinator's `## Skills` repeats each child's registry lines with the child's path as prefix, as HA-IF-19 of [instruction files](../harness-audit/references/instruction-files.md) sets.
- A host's native worktree made from the coordinator holds none of the children: create worktrees inside the child that holds the change.
- A change across repositories is still one pull request per repository; the coordinator makes no change atomic.
