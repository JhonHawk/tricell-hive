---
name: git-mechanics
description: >
  Load BEFORE the first git verb of a session — branching, committing, opening or merging a
  PR, promoting between environments, or pruning branches once the work is done. Covers the
  branching model per repo class, the session git mode and commit semantics, PR and promotion
  gates with the two review phases, end-of-work pruning, and recovery.
  Triggers: branch, commit, push, PR, merge, promote, release, plan finished.
---

# git-mechanics — how to branch, commit, review, promote and close

The situational half of the git conventions. **The gates are not here**: what authorizes an
operation, protected branches, force-push, and promotion into the production-deploying branch
stay always-on in `git-workflow.md`, because a gate that depends on a model invoking a skill
is not a gate. Read that one for *whether you may*; this one for *how*.

## Routing table

| Situation | Read |
|---|---|
| Cutting a branch, picking its name, or deciding the repo's branching model and base branch | `git-mechanics.md > Branching` |
| First edit-intent of the session (the git mode question), or writing a commit message | `git-mechanics.md > Commits` |
| Opening a PR, choosing the review route, waiting on CI, merging, promoting to an environment | `git-mechanics.md > PRs & promotion` |
| The plan or change-group finished: pruning merged branches, handling unmerged ones | `git-mechanics.md > End-of-work hygiene` |
| A push 404s on a private repo | `git-mechanics.md > Recovery` |

Reference injected at build time from `global/rules-situational/`. Stable path after deploy:
`~/.agents/skills/git-mechanics/references/git-mechanics.md` (Claude Code and Grok) or
`~/.agents/skills/git-mechanics/references/git-mechanics.md` (Codex and opencode).

## Rules of use

- **Read it before the first git verb, not at commit time.** The session git mode is decided
  at first edit-intent and the branch name is decided before the branch exists — both are
  already behind you by the time `git commit` is typed.
- The precedence that governs all of it — explicit user verb > invoked flow skill's declared
  git scope > session defaults — is in the always-on half, along with the gates.
- A repo's own `AGENTS.md`/`CLAUDE.md` declaration (`Base branch`, `Git mode`, `PR review`)
  wins over anything here; read the repo before reading this.
- A reference already loaded this session is not reloaded.
