---
name: task-routing
description: >
  Load before the FIRST `Write`/`Edit` on project code — the same moment `git-mechanics`
  fires — and before deciding HOW work gets done: which specialist takes
  it, whether to delegate at all, how a multi-domain task is chained and verified, and what
  must be settled before a plan's tasks are written. It carries the gap analysis whose
  prerequisites (accounts and roles, seed rows, running services) are cheapest resolved
  before the code, not at the verification gate. Any request to review, audit, investigate,
  diagnose, refactor across files, or "how would you approach X" is this decision, even when
  the user never says the word delegate — they ask for a result, not for a routing choice.
  NOT for: a read-only investigation, a short question, or work outside a software project —
  none reach edit-intent, and the trivial carve-out (typo, rename, one-line config) is out too.
  Triggers: implement, build, feature, fix, review, audit, investigate, diagnose, plan,
  approach, refactor, verify; implementa, construye, arregla, revisar, auditar, investigar,
  diagnosticar, planear, "cómo lo abordarías".
---

# task-routing — who does the work, and what must be settled before it starts

Two rules that govern the moment before execution: **who** a task goes to, and **what has to
be resolved** before its tasks are written. Neither is always-on, because neither has a file
that announces it — you delegate and you plan by intent, not by opening a `.tsx`.

That is also the risk. Nothing loads these for you: **not invoking this skill is identical to
not having the rules.** Delegating without them means guessing the specialist and skipping the
verification layer; planning without them means gaps that surface mid-execution instead of at
the plan gate.

## Routing table — read the row that matches

| Situation | Read |
|---|---|
| Choosing which agent gets a task; a task spanning 2+ domains; whether to delegate at all; how verification and review are staged | `agent-routing.md` |
| Writing the tasks of a plan; a prerequisite that blocks the work; two authoritative sources that disagree | `gap-resolution.md` |
| Judging whether a change is TRIVIAL (the carve-out that scales review, tests and risk-surfacing down); deciding who OWNS an open decision; whether an alternative is worth surfacing | `critical-thinking.md` |

References are injected at build time from `global/rules-situational/` into `references/`.
Stable path after deploy: `~/.agents/skills/task-routing/references/<file>.md` (Claude Code
and Grok) or `~/.agents/skills/task-routing/references/<file>.md` (Codex and opencode).

## Rules of use

- **Read it before the first delegation of a session, not after.** The gates in
  `agent-routing.md` are counted in tool calls and files touched; by the time delegating
  feels overdue, the threshold is already crossed.
- A task in one clear domain routes to that specialist without reading anything: the table
  exists for the ambiguous cases and for the chains.
- **Trivial mechanical work is never delegated** — that answer needs no reference.
- A reference already loaded this session is not reloaded.
- Only these two files live here. Rules cited by name from inside them belong elsewhere:
  `testing.md` and the quality depth rules → `language-rules/references/`;
  `git-workflow.md`, `security.md` and the other gates are always-on and already in context.
