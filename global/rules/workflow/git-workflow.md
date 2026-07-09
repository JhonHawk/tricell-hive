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

### Branching
- **Decide at first commit intent, not before** — editing needs no branch decision (`git switch -c` carries uncommitted changes losslessly). Ask once per session, folding into the verb's question block when both fire: stay on the current branch, or a dedicated branch per change-group (**Isolated Mode** — also auto-activated by a pre-existing worktree or an explicit user instruction). Default: stay. An approved plan that declares git semantics settles this at the plan gate instead.
- **A branch's unit is a cohesive change-group** — one open thematic branch at a time; a new theme waits until the previous one integrates. A worktree is the explicit exception for genuine overlap.
- **Trunk-shaped, not git-flow:** no permanent `develop`/`release`/`hotfix` branches unless the repo already uses them. Names: `<type>/<kebab-description>` mirroring commit prefixes (`feat/user-auth`).

### Commits
- **Granularity is inferred, never asked up-front.** Dedicated branch/worktree → autonomous, granular commits at natural seams (an auditable, bisectable N-commit history — the standing default in every repo; a project may declare squash-only or other semantics). Shared/protected branch → per-changeset confirmation; at the second explicit commit request, offer session autonomy once.
- **Conventional commits** (`feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`, `perf:`, `ci:`; optional scope), small and focused — one concern per commit, split mixed ones. Never any AI attribution in commits, PRs, messages, or file outputs.
- **Checkpoints** (Isolated Mode, plans with 4+ file-modifying tasks): a progress pulse roughly every 3 such tasks — a brief note by default; a blocking question only when something is decision-relevant (deviation, concern, red tests).

### Safety gates (all modes — fold confirmations into the task's single question block)
- **Protected branches** (project-declared; default `main`/`master`): no direct commits unless the project allows it. Exceptions: a single-branch repo (bare `master` IS the workflow), or a repo whose history shows direct-to-default as the norm (no PR gate, no CI on branches) — confirm once per session and treat as standing; a project declaration removes even that first ask.
- **Force-push and published-history rewrites** (`rebase` on shared branches, `reset --hard`, `amend` on pushed commits): always confirm, presenting what gets overwritten.
- **Merge to `main`/`master`:** always confirm.
- **Push:** an explicit push verb is the confirmation; otherwise ask. Session-close remote pruning and declared-workflow repos (exceptions above) are standing-authorized.
- **Branch deletion:** confirm — except confirmed-merged branches at session close (below). Unmerged deletion, force-delete, or unverifiable merge status always asks.

### PRs & promotion
- PR title in conventional commit format; description covers what/why plus verification. One logical change per PR.
- **Shared integration branches receive changes via PR**, never a direct local merge, wherever the project uses a PR gate; a solo repo without one may merge locally.
- **You own the CI of any PR you merge until the landed branch is green:** wait for PR checks before merging (never on red or pending), watch the post-merge run, and fix failures you triggered instead of handing them back. CI is authoritative over local runs.
- **Promotion moves the integration branch's whole current state**, not the session diff, unless the user explicitly scopes a partial promotion.

### Session close (standing-authorized, automatic)
- Prune branches confirmed 100% merged (`git branch --merged <target>`): delete the local and its merged remote counterpart; return the checkout to the integration branch. Then `git fetch --prune` and fast-forward local integration branches behind their upstream.
- **Unmerged branches are decisions, not noise:** check for real work (`git log <target>..<branch>`), recommend integrating before close, or ask if it looks abandoned. Never silently leave, delete, or auto-reconcile a diverged branch.
- Report: deletions, fast-forwards, divergences, unmerged branches with their recommendation, and the branch now checked out.

### Recovery
- `push` returns 404 on a private repo: check `gh auth status` / `gh auth switch` before assuming the repo is missing; cache the active account for the session.
