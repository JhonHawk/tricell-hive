---
alwaysApply: true
---

## Git Workflow

> Universal git conventions with two operating modes. Project-specific overrides (branching model, protected branches) live in each project's CLAUDE.md.

### Explicit Git Tasks

> Applies whenever the user issues a direct git verb on **any** operation — `commit`, `push`, `pull`, `merge`, `rebase`, `cherry-pick`, `tag`, `stash`, `branch` (create/switch/delete), `worktree` (including via the `EnterWorktree`/`ExitWorktree` tools), `reset`, etc. The verb itself is the authorization. Goal: zero redundant confirmations during the run, zero scope creep.

**Three-step pattern:**

1. **Single contextual question block (`AskUserQuestion`)** — at most one round-trip at the start, consolidating every decision needed for the whole chain.
   - Only ask what cannot be inferred from the diff, the current session, the user's order, or repo state.
   - Max 4 questions per call (tool limit).
   - Questions are contextual to the task — never boilerplate. The catalog below is orientative, not a checklist.
   - If everything is inferable, **skip the block** and emit a brief start summary instead.

2. **Fluid execution** — apply the answers across the entire requested chain. No further interruptions.

3. **End summary** — SHAs, refs/branches updated, working tree state, and (when relevant) the next logical step **as a suggestion, not an action**.

**What to consider asking, by task (orientative — depends on context):**

| Task | Decisions to evaluate for the block |
|---|---|
| `commit` | File scope (if working tree mixes session and unrelated changes), proposed message (only if diff is ambiguous), whether to chain a push |
| `push` | Remote/branch destination (if ambiguous), what to do on upstream divergence (pull, rebase, force) |
| `merge` | Source branch (if not stated), strategy (merge commit, squash, rebase), conflict policy |
| `rebase` | Onto which ref, interactive or not, conflict policy |
| `branch` | Name (if not given), base, checkout after creation |
| `tag` | Annotated or lightweight, message (if annotated), push the tag |
| `stash` | Include untracked, stash message |
| `reset` / `cherry-pick` | Explicit risk confirmation, mode (`--soft`/`--mixed`/`--hard`) |

**Inference rules — try first, ask only when ambiguous:**
- **Commit message:** if the diff cleanly suggests a Conventional Commit message, draft it and show it in the start summary. Don't ask.
- **Files to stage:** if the session touched identifiable files, stage only those. If the entire working tree belongs to the session, stage all. Only ask when ownership is ambiguous.
- **Remote:** if the repo has a single remote and the branch has an upstream configured, use it without asking.

**Strict scope — the golden rule.** Execute **only** the requested task plus what the initial block authorizes. If the user asked for `commit` and the block did not add `push`, do not push. Never chain merge/rebase/pull/tag without explicit authorization. Do not "complete the flow" on your own initiative. Authorization is bounded to the changeset and the verbs the user issued — it does NOT extend to later changesets unless the session is in autonomous commit mode (see Direct Mode), and it does NOT extend to verbs the user did not issue, regardless of session mode.

**No pre-commit gates inside the git command.** Lint, typecheck, build, and tests are **not** executed as part of the `commit` command. Build validation at task level still applies (see `global/CLAUDE.md > Build & Lint`) and runs separately, before the git task — not folded into it.

**Safety gates remain intact.** The gates in `### Safety (both modes)` still require explicit confirmation. When a requested task triggers one, fold the confirmation into the **same** initial question block — a single round-trip, not a second interruption.

### Flow-skill scope

User-invoked flow skills (`/flow-build`, `/flow-deploy`, …) may declare their own git
semantics in their body — e.g. per-task commits, autonomous push and PR on the session
branch. Within that skill's scope, **invocation IS the authorization**:

- The once-per-session questions below (branching strategy, commit mode) do NOT re-fire
  inside the skill — its approved plan settles both for its scope.
- Isolated Mode checkpoints are satisfied by the skill's own per-task verification and
  close report.
- The permanent safety gates never relax, regardless of skill scope: commits, pushes,
  and merges **to protected branches**, force-push, and published-history rewrites still
  require explicit user confirmation. Two carve-outs:
  - **User authorization** — per invocation, or standing in the project's `CLAUDE.md`.
  - **Single-branch repos** — when the protected default branch is the repo's ONLY
    branch (specs or mocks repos living on bare `master`), committing and pushing to it
    IS the declared workflow; no per-operation confirmation. Force-push and history
    rewrites stay gated.
- Merges to a **non-protected integration branch** (e.g. `development`) are authorized
  by the flow skill's approved plan when it declared the merge semantics (target branch,
  merge mechanics, CI gate). A flow that verified its work merges it — leaving approved
  work unmerged on N sibling branches manufactures conflicts for the user.

### Once-per-session decisions

Two decisions are settled once per session and persist for its duration: branching strategy and commit mode. Both follow the same pattern.

**Pattern:**
- Asked once via `AskUserQuestion` when the decision first becomes relevant.
- Persists across the session and across every repo touched.
- User may switch mid-session with an explicit phrase; takes effect immediately.
- Ambiguous answers default to the least-disruptive option (specified per instance).

#### Branching strategy
- **Trigger.** Before the first file edit in any git repository.
- **Per-repo override is the default tiebreaker.** When the session-wide strategy clashes with a specific repo's state (checked out on a protected branch with no working-branch equivalent, or the repo's `CLAUDE.md` declares different branching defaults), ask narrowly for that repo without re-asking the global strategy. Strategy persists across repos only when each repo's state is compatible — never override a protected-branch safety gate to honour a global answer.
- **Options:**
  1. **Stay on current branch (multi-commit).** Work and commit on the checked-out branch. Safety gates for protected branches still apply.
  2. **Create a dedicated branch per change-group.** Claude proposes branch name + base derived from the project's declared branching model (its `CLAUDE.md`); user confirms. Enters Isolated Mode.
- **Default (ambiguous):** Stay on current branch.
- **A branch's unit is a cohesive change-group, not a single commit.** Commits that share one theme stay on the same branch — group them. A genuinely different theme earns a new branch. **Themes don't run in parallel:** do not open a branch for a new theme until the previous one is fully integrated (merged/PR landed). One open thematic branch at a time keeps history linear and review boundaries clean; the cost is no concurrent thematic work — when you genuinely need overlap, use a worktree and say so. Worktree isolation is the explicit exception, not the silent default.
- **Not git-flow.** This model is deliberately trunk/GitHub-flow-shaped — no permanent `develop`/`release`/`hotfix` branches; don't reintroduce them.
- **Skip entirely when:** not inside a git repo, session is 100% read/analysis, or a pre-existing worktree is detected (Isolated Mode auto-activates).

#### Commit mode
- **Trigger.** First commit context (user verb or post-implementation offer) in Direct Mode.
- **Options:**
  1. **Per-changeset confirmation.** Claude waits or offers; user confirms each commit. Explicit verbs (`commit`, `commit y push`) authorize only the changeset at hand.
  2. **Autonomous until further notice.** Claude commits during the session without per-commit confirmation.
- **Default (ambiguous):** Per-changeset confirmation.
- **Scope:** covers only `commit`. `push`, `merge`, `tag`, and any safety-gated operation remain governed by `### Safety (both modes)` regardless of mode.

### Modes of Operation

#### Direct Mode (default)
Active when the branching strategy resolves to "Stay on current branch".

- **Protected branches** are declared in the project's `CLAUDE.md` (default when undeclared: `main`, `master`). Safety gates in `### Safety (both modes)` always apply.
- Commit behavior follows `### Once-per-session decisions > Commit mode`.
- **Next steps after commit:** *suggest* (PR, push, deploy); never execute without explicit request.
- Non-commit verbs (`push`, `pull`, `merge`, `rebase`, `tag`) require explicit user instruction per `### Explicit Git Tasks`.

#### Isolated Mode (worktree / dedicated branch)
Activated when **any** of:
1. Working inside a **git worktree** (`git worktree list`).
2. The user chose "Create a dedicated branch per change-group" in the branching strategy.
3. The user explicitly instructed Claude to work on a dedicated branch.

- **Commits are autonomous** within the dedicated branch — safety checkpoints during execution.
- **Branch creation is NOT autonomous** — requires the branching strategy answer (option 2) or an explicit user verb.
- **Control gate:** merge/PR/keep/discard decision at the end (pattern: `finishing-a-development-branch`).
- **Explicit git tasks still apply** for direct user verbs (`### Explicit Git Tasks`).

##### Checkpoints
For plans with **4+ tasks** that modify files, intermediate checkpoints are mandatory:
- **Frequency:** pause at natural seams — roughly every 3 *file-modifying* tasks (read/research tasks don't count). If a single task touches many files or produces a large diff, pause sooner. If a checkpoint arrives with nothing decision-relevant (no deviations, no concerns, tests green), emit a brief progress note and continue instead of a blocking `AskUserQuestion`.
- **Content:** brief summary of completed tasks, any concerns or deviations, test status.
- **Options:** continue / adjust / pause.
- **Not a code review** — spec and quality reviewers already handle that. This is a progress pulse.

Plans with 1-3 file-modifying tasks, or research-only plans: final gate only (no intermediate checkpoints).

### Conventions (both modes)
- **Conventional commits required.** Prefixes: `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`, `perf:`, `ci:`. Scope optional: `feat(auth):`.
- **Commits must be small and focused.** One concern per commit. If a task touches multiple concerns, split into separate commits.
- **Never include attribution, "Generated with Claude Code", "Co-Authored-By", or any AI credit** in commits, PRs, messages, or file outputs.
- **A documentary `sessions/` folder ≠ a git work-session.** The session capture layer (`project-structure.md > Session capture layer`) versions the session journal in the specs repo (or `<repo>/_support/sessions/` standalone) and keeps its raw gitignored in `_support/` — commit the journal, never the raw. This is orthogonal to the branching/commit/cleanup semantics here, which use "session" in the work-session sense.

### Branches
- **Branch naming:** `<type>/<short-description>` in kebab-case. Examples: `feat/user-auth`, `fix/invoice-rounding`, `chore/upgrade-deps`.

### Pull Requests
- **PR title follows conventional commit format.** Same prefixes as commits.
- **PR description:** summary of changes (what and why) + test plan or verification steps. No boilerplate filler.
- **One PR per logical change.** Don't bundle unrelated work. If a refactor enables a feature, consider splitting into two PRs.
- **You own the CI of any PR you merged — until the branch it landed on is green.** (Inert where the repo runs no CI on its branches.) Watch depends on where CI runs: a **PR check** → wait for it to conclude before merging (never merge on red or pending); CI on the **post-merge push** → watch that run after merging. Use the host's own tooling — `gh pr checks <n> --watch` / `gh run watch` (the equivalent for GitLab/Bitbucket/etc). A merge that leaves the target branch red is a blocker *you* triggered — surface the failing output and fix it, never hand "the CI is failing" back to the user. Local lint/build is necessary but not sufficient (`testing.md > Execution Scope`): CI is authoritative over the local subset. **Trigger-independent** — explicit `merge` verb, autonomous flow merge, or a drifted flow session alike.
- **Shared integration branches receive changes via PR, never a direct local merge.** When the project integrates through PRs — any repo with a CI/review gate on PRs, the common case — a feature or task branch lands on the integration branch (`development`, `main`, …) by opening and merging a PR, not by a local `git merge`/`--no-ff` + push. A direct local merge bypasses the review and CI the PR exists to enforce, and skips the CI-ownership watch above. *Applies where the project uses PRs; a solo repo with no PR/CI flow may merge locally.* Holds under flow autonomy and when a flow session drifts to direct interaction alike — "merge to `development`" at a session's end still means open the PR, not `git merge` it.

### Environment Promotion
- **Promote the integration branch's WHOLE current state, not the session's changes.** "Promote/deploy to QA (or prod)" — whether suggesting it or executing it — defaults to everything currently on the integration branch, which accumulates work across sessions; promoting only the recent diff leaves prior work unpromoted and desyncs environments. A partial or cherry-picked promotion happens ONLY when the user explicitly scopes it.

### Branch Cleanup & Sync (session close)

A work branch that outlives its merge — or a local integration branch left behind its remote — is what makes the next session start on the wrong or stale branch. Resolve at session close, not necessarily per merge.

- **At session close, prune merged work branches and return to the integration branch.** For each local branch confirmed merged into its target (`git branch --merged <target>` proves it): delete the local, and delete its remote counterpart if it too is merged. If the checkout sits on a merged work branch, switch back to the integration branch (`development`/`main`) first — that switch is what fixes "next session starts on the wrong branch".
- **In the same pass, sync the local integration branches with the remote.** `git fetch --prune` first (it also drops remote-tracking refs for branches deleted upstream — part of the cleanup). Then fast-forward each local integration branch behind its upstream (`development`/`main`/`qa`): `git merge --ff-only` for the checked-out one, `git fetch origin <branch>:<branch>` for the others. This leaves the next session on an up-to-date base.
- **This is automatic — no per-branch confirmation.** A confirmed-merged branch has a known rollback path (its commits live on the target and remote), and a fast-forward and `fetch` change no working-tree content — both meet the destructive-op carve-out in `global/CLAUDE.md > Destructive Operations`.
- **Unmerged branches are reviewed and integrated, not silently left or deleted.** For each local branch NOT merged into its target, inspect whether it carries real un-integrated work (`git log <target>..<branch>`): if it does, surface it with a recommendation to integrate (open a PR / merge per the project's flow) before closing; if it looks like abandoned leftover, ask what to do. Never delete it, never auto-merge/rebase it without confirmation, and never close the session leaving it unreported — an unmerged branch is a decision to take, not noise to skip.
- **Still gated — ask first (destructive):** deleting an unmerged branch (loses work), force-delete (`git branch -D` / `push --delete` of unmerged), any branch whose merge status you could not verify, or reconciling a non-fast-forward integration branch (local commits the remote lacks) — never auto-merge or rebase to reconcile divergence.
- **Report in the close summary:** which locals/remotes were deleted, which integration branches were fast-forwarded (and any that diverged), any unmerged branches found with their integrate/keep/discard recommendation, and the branch now checked out.

### Safety (both modes)
- **Never commit directly to protected branches** (declared in the project's `CLAUDE.md`, see `#### Direct Mode`) unless the project explicitly allows it. The default protected list is `main`/`master`; a project declaration may extend or modify it. **Single-branch exception:** when the protected default is the repo's only branch (specs/mocks repos on bare `master`), commits and pushes to it are the normal workflow — this gate and the push gate below don't apply there; force-push and history rewrites still do.
- **Never force-push** without explicit user confirmation. Present what will be overwritten and the risk.
- **Never merge to main/master** without user confirmation.
- **Never push to remote** without user confirmation.
- **Never rewrite published history** (`rebase` on shared branches, `reset --hard`, `amend` on pushed commits) without confirmation.
- **Never delete branches** (local or remote) without confirmation. **Carve-out (session-close cleanup):** local and remote branches confirmed 100% merged into their integration target (`git branch --merged`) are pruned automatically at session close — see `### Branch Cleanup & Sync (session close)`. Unmerged branches, force-delete, and any unverifiable merge status stay gated.

### Recovery
- **`push` returns 404 "Repository not found"**: reactive check, not preventive. When the 404 surfaces (or the active `gh` account is unknown for this session), run `gh auth status` and switch with `gh auth switch --user <name>` if needed. Cache the active account for the rest of the session — do not re-check on every push. GitHub returns 404 (not 403) for unauthorized private repos by design; don't read 404 as "missing repo" until auth is verified.
