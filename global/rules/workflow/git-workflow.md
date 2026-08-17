---
alwaysApply: true
---

## Git Workflow

> Universal conventions. Project `CLAUDE.md`/`AGENTS.md` overrides (protected branches, branching model, commit semantics) win. **Precedence:** explicit user verb > invoked flow skill's declared git scope > session defaults — never re-ask at a lower level what a higher level settled. Safety gates apply at every level.

### Authorization
- **An explicit git verb IS the authorization** for that operation on the changeset at hand. Zero redundant confirmations: consolidate anything genuinely ambiguous (file scope, destination, strategy, a safety gate's confirmation) into ONE question block up front, execute the whole chain without interruption, then summarize — SHAs, refs updated, tree state, next step as a suggestion, never an action.
- **Infer before asking:** draft the Conventional Commit message from the diff; stage the session's files (all, when the whole tree is the session's); use the single remote/upstream. Ask only what the diff, session, or repo state cannot answer.
- **Strict scope:** execute only the requested verbs plus what the question block authorized — never chain unrequested operations or "complete the flow". Authorization does not extend to later changesets unless commit autonomy applies (below). Session-close hygiene (below) is standing-authorized.
- **Flow skills:** a user-invoked flow skill's declared git semantics (per-task commits, push, PR, merge to a non-protected integration branch) are authorized by its invocation and approved plan. Safety gates never relax.
- **No pre-commit gates inside git commands** — lint/build/tests run at task level, before the git task.
### Safety gates (all modes — fold confirmations into the task's single question block)
- **Protected branches** (project-declared; default: the production-deploying branch — `production`, or `main`/`master` where the trunk IS production — plus `qa` where it exists): no direct commits unless the project allows it. Exceptions: a single-branch repo (bare `master` IS the workflow), or a repo whose history shows direct-to-default as the norm (no PR gate, no CI on branches) — confirm once per session and treat as standing; a project declaration removes even that first ask.
- **Force-push and published-history rewrites** (`rebase` on shared branches, `reset --hard`, `amend` on pushed commits): always confirm, presenting what gets overwritten.
- **Merge/promotion into the production-deploying branch** (`production`, or `main`/`master` where the trunk IS production): always confirm — it is a production deploy (`CLAUDE.md > Destructive Operations`), and no rollback path substitutes for the confirmation. Promotion into `qa` follows that rule's non-prod carve-out — declared, not asked.
- **Push:** an explicit push verb is the confirmation; otherwise ask. Session-close remote pruning and declared-workflow repos (exceptions above) are standing-authorized.
- **Branch deletion:** confirm — except confirmed-merged branches at session close (below). Unmerged deletion, force-delete, or unverifiable merge status always asks.

### Everything else lives in `git-mechanics.md`

Branching model, commit semantics and the session mode, PRs and promotion, session-close
hygiene, and recovery are situational: they apply once you are already branching, committing,
or opening a PR. They load through the `git-mechanics` skill — **invoke it before the first
git verb of a session**, since choosing a branch name or a session mode happens before any
of it is on screen. The gates above stay here because their trigger is an action whose cost
is irreversible, and a gate that depends on a model invoking a skill is not a gate.
