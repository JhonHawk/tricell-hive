---
name: task-routing
description: >
  Load before substantive research, BI analysis, architecture, infrastructure, agent
  configuration or development work; also before the first project edit-intent, git verb,
  delegation or plan's tasks. Covers deliverable and effects, specialist routing,
  prerequisites, verification, task records and applicable git mechanics. Short self-contained
  questions and trivial mechanical work skip the intake; project edit/git gates still apply.
  Triggers: review, audit, investigate, diagnose, analyze, design, refactor, plan, branch,
  commit, PR, merge, promote; revisar, auditar, investigar, analizar, diseñar, planear.
---

# task-routing — deliverable, effects, execution and evidence

Use the common intake in `agent-routing.md` for substantive work, including work outside
software projects. Infer what the conversation settles; ask only for missing information
that changes the work. Read git mechanics when project edits or git operations apply.

## Routing table — read the row that matches

| Situation | Read |
|---|---|
| Starting substantive work: requested deliverable, scope, authorized effects, completion evidence and stopping condition | `agent-routing.md > Common intake` |
| Choosing which agent gets a task; a task spanning 2+ domains; whether to delegate at all; how verification and review are staged | `agent-routing.md` |
| Writing the tasks of a plan; a prerequisite that blocks the work; two authoritative sources that disagree | `gap-resolution.md` |
| Judging whether a change is TRIVIAL (the carve-out that scales review, tests and risk-surfacing down); deciding who OWNS an open decision; whether an alternative is worth surfacing | `critical-thinking.md` |
| Cutting a branch, picking its name, or deciding the repo's branching model and base branch | `git-mechanics.md > Branching` |
| First edit-intent of the session (the git mode question), or writing a commit message | `git-mechanics.md > Commits` |
| Work without a `/flow-plan` contract (the direct route): whether it needs a task record | `agent-routing.md > Direct Route` |
| Opening a PR, choosing the review route, waiting on CI, merging, promoting to an environment | `git-mechanics.md > PRs & promotion` |
| Repository change-group finished: pruning merged branches, handling unmerged ones | `git-mechanics.md > End-of-work hygiene` |
| A push 404s on a private repo | `git-mechanics.md > Recovery` |

References are injected at build time from `global/rules-situational/` into `references/`.
Stable path after deploy: `~/.agents/skills/task-routing/references/<file>.md` (Claude Code
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
- A repo's own `AGENTS.md`/`CLAUDE.md` declaration (`Base branch`, `Issue tracker`) wins
  over anything here; `PR review` is the RECOMMENDED option of a question still asked every
  session that opens a PR (`git-mechanics.md > PRs & promotion`). Read the repo before this.
- A task in one clear domain uses its known specialist; ambiguity uses the table. This does
  not skip the common intake, applicable prerequisites or authorization gates.
- **Trivial mechanical work is never delegated** — that answer needs no reference.
- A reference already loaded this session is not reloaded.
- Only these four files live here. Rules cited by name from inside them belong elsewhere:
  `testing.md` and the quality depth rules → `language-rules/references/`;
  `git-workflow.md`, `security-floor.md` and the other gates are always-on and already in context.
