---
name: task-routing
description: >
  Load before the FIRST `Write`/`Edit` on project code OR the first git verb of the session —
  whichever comes first: a commit-only session fires on the git verb without ever reaching
  edit-intent. Covers HOW work gets done — which specialist takes it, whether to delegate at all,
  how a multi-domain task is chained and verified, the gap analysis that settles prerequisites
  before a plan's tasks are written, the direct route's task record before the first write — AND the git mechanics: branching, session git mode, commits,
  PR/promotion, close. Any request to review, audit, investigate, diagnose, refactor across files,
  or "how would you approach X" is this decision, even when the user never says "delegate" — they
  ask for a result, not a routing choice. NOT for: a short question, work outside a software
  project, a SMALL read-only investigation (≤3 files, ≤3 queries), or the trivial carve-out (typo,
  rename, one-line config). Triggers: review, audit, investigate, diagnose, refactor, plan,
  branch, commit, PR, merge, promote; revisar, auditar, investigar, planear.
---

# task-routing — who does the work, what must be settled first, and how git carries it

Three rule sets that govern the moment before execution: **who** a task goes to, **what has to
be resolved** before its tasks are written, and **how the session branches, commits, reviews,
promotes and closes**. None is always-on, because none has a file that announces it — you
delegate, plan, and pick a branch name by intent, not by opening a `.tsx`.

That is also the risk. Nothing loads these for you: **not invoking this skill is identical to
not having the rules.** Delegating without them means guessing the specialist and skipping the
verification layer; planning without them means gaps that surface mid-execution instead of at
the plan gate; committing without them means a session mode and branch chosen by accident.

## Routing table — read the row that matches

| Situation | Read |
|---|---|
| Choosing which agent gets a task; a task spanning 2+ domains; whether to delegate at all; how verification and review are staged | `agent-routing.md` |
| Writing the tasks of a plan; a prerequisite that blocks the work; two authoritative sources that disagree | `gap-resolution.md` |
| Judging whether a change is TRIVIAL (the carve-out that scales review, tests and risk-surfacing down); deciding who OWNS an open decision; whether an alternative is worth surfacing | `critical-thinking.md` |
| Cutting a branch, picking its name, or deciding the repo's branching model and base branch | `git-mechanics.md > Branching` |
| First edit-intent of the session (the git mode question), or writing a commit message | `git-mechanics.md > Commits` |
| Starting implementation with no `/flow-plan` contract (the direct route), before the first write — whether a task record is created | `agent-routing.md > Direct Route` |
| Opening a PR, choosing the review route, waiting on CI, merging, promoting to an environment | `git-mechanics.md > PRs & promotion` |
| The plan or change-group finished: pruning merged branches, handling unmerged ones | `git-mechanics.md > End-of-work hygiene` |
| A push 404s on a private repo | `git-mechanics.md > Recovery` |

References are injected at build time from `global/rules-situational/` into `references/`.
Stable path after deploy: `~/.claude/skills/task-routing/references/<file>.md` (Claude Code
and Grok) or `~/.agents/skills/task-routing/references/<file>.md` (Codex and opencode).

## Rules of use

- **Read it before the first delegation of a session, not after.** The gates in
  `agent-routing.md` are counted in tool calls and files touched; by the time delegating
  feels overdue, the threshold is already crossed.
- **Read the git rows before the first git verb, not at commit time.** The session git mode
  is decided at first edit-intent and the branch name is decided before the branch exists —
  both are already behind you by the time `git commit` is typed.
- **Reading the `git-mechanics.md` reference does not discharge the session-mode ask.** At
  first edit-intent the next ACT is emitting that structured question — all four options.
- **Sizing or listing a reference (`wc`, `ls`, head-of-file) does not discharge reading it** —
  when a routing-table row's moment fires, the row's file gets READ before acting.
- **The git gates are not here.** What authorizes an operation, protected branches,
  force-push, and promotion into the production-deploying branch stay always-on in
  `git-workflow.md` — a gate that depends on a model invoking a skill is not a gate — along
  with the precedence that governs all of it (explicit user verb > invoked flow skill's
  declared git scope > session defaults). Read that one for *whether you may*; the
  `git-mechanics.md` reference for *how*.
- A repo's own `AGENTS.md`/`CLAUDE.md` declaration (`Base branch`, `Git mode`, `PR review`)
  wins over anything here; read the repo before reading this.
- A task in one clear domain routes to that specialist without reading anything: the table
  exists for the ambiguous cases and for the chains.
- **Trivial mechanical work is never delegated** — that answer needs no reference.
- A reference already loaded this session is not reloaded.
- Only these four files live here. Rules cited by name from inside them belong elsewhere:
  `testing.md` and the quality depth rules → `language-rules/references/`;
  `git-workflow.md`, `security.md` and the other gates are always-on and already in context.
