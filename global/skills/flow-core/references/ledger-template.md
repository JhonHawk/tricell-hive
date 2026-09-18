# PROJECT.md — ledger template

The ledger is the **operational index** of a client workspace: it answers "where is this
project and where is everything" for any session that opens it. It is NOT an archive — it
points, the specs repo preserves. It lives at `<project>/_support/PROJECT.md` (non-versioned
layer) precisely because its value is current state, not history.

Two consumers read it: flow skills (contract step OPEN) and conversational sessions (via the
pointer in the workspace CLAUDE.md). Write for both: terse, scannable, paths always concrete.

## Maintenance rules

- Every flow skill updates the ledger at CLOSE. Conversational sessions should update it when
  they produce artifacts worth finding later — if they don't, applying the
  workspace-hygiene-playbook repairs it.
- A decision recorded here is a **pointer**: once the specs repo exists, the decision's full
  text gets promoted to `<project>-specs/` and the ledger row links to it. A decision that
  lives only in the ledger is at risk — the ledger is not versioned.
- Keep "Artifacts" pruned: when a scratch file expires or gets promoted, update the row.
  Stale pointers are worse than no pointers — they teach readers to distrust the index.
- **The ledger does NOT index sessions.** A session's index is versioned and co-located at
  `<project>-specs/sessions/README.md` (or `<repo>/_support/sessions/README.md` for a
  standalone repo). The ledger is non-versioned and lives outside the specs repo, so it cannot
  be the index of versioned sessions — it points to the specs repo and does not list sessions.
  A promoted session decision still gets a Decisions row (`Promoted to` + the originating
  session slug). See `project-structure.md > Session capture layer`.
- **Handoff poda is mandatory.** `## Current handoff` holds only the last closed phase's
  handoff. A phase that consumes it at OPEN (DoR) collapses it to one line in
  `## Handoff history` before writing its own. Two expanded handoffs at once means a phase
  forgot to poda — the workspace-hygiene-playbook flags it. History grows one line per phase
  forever; the cost is negligible and the trail is worth keeping.
- Dates in absolute form (`2026-06-11`), never "yesterday" or "last week".
- **Tracker fields are how flow skills resolve "the tracker"** (shared access/batching mechanics: `references/tracker-access.md`) — they never hardcode
  Linear/Jira. `Tracker: none` means the project is untracked: the specs repo's
  `tasks.md` files are the only task source, and close phases skip tracker updates.
  `Tracker access` decides the mechanics: `mcp` (load via ToolSearch), `cli` (`acli`
  for Jira — often the most effective route), `api` (REST/GraphQL with the named env
  token), `manual` (no programmatic access: read tasks from the specs repo and report
  "update the tracker manually: …" items at close, never fail or silently skip).
- **Deferred ≠ pending.** A task carrying the project's `Deferred marker` was consciously
  paused — "worth doing, not now" (a postponed 2FA, a feature waiting on a client
  decision). Flow skills treat it as out of scope: never a scope candidate, never listed
  among pendientes, never in next-session suggestions — at most a one-line count with IDs.
  It re-enters scope only when the user explicitly un-defers it; removing the marker IS
  that decision. Untracked projects (`Tracker: none`) mark deferral directly in the
  epic's `tasks.md` with the declared marker.

## Template

```markdown
# <project> — Project Ledger

| | |
|---|---|
| Group / client | <group> |
| Project token | <project-token>        <!-- seed for all infra naming --> |
| Current stage | arranque / specs / desarrollo / operación |
| Specs repo | <project>-specs/ (or "not yet created — specs stage pending") |
| Tracker | linear (team FAC) / jira (project ATSCL) / none |
| Tracker access | mcp / cli (acli) / api (env <TOKEN_VAR>) / manual — ordered by preference when several work |
| Deferred marker | linear label "deferred" / jira label / tasks.md `[deferred]` / none — how paused-by-decision work is marked |
| Executor dispatch | free (default) / plan-required — `plan-required` makes the `executor-dispatch-gate` hook DENY dispatching an executor agent or editing project code inline while no plan holds implementation authority (docs, the ledger, plans and task records are never gated); `free` only advises |
| Release notes | OPTIONAL — overrides for `release-notes-template.md` (canal: correo / whatsapp / slack · idioma · tono); absent = template defaults |
| QA walk | OPTIONAL — who runs the functional walk after a QA promotion: `sdd-verify` (dispatched automatically after liveness, report versioned) / `offer` (default — stated pending at close, `sdd-verify` offered) / `none` (client QA owns it; never run or offer, close is liveness only) |
| Last updated | YYYY-MM-DD (<what changed>) |

## Stage status

| Stage | Status | Closed on | Key outputs |
|---|---|---|---|
| arranque | done / in progress / pending | YYYY-MM-DD | <workspace + foundation paths> |
| specs | … | | <specs repo, epics> |
| desarrollo | in progress (epic <id>) | — | <branch / PRs; per-unit chain: brainstorm → spec → plan → build> |
| operación | … | | <promoted envs, release notes> |

## Current handoff

<!-- DoD of the last closed phase; DoR of the next. Only the latest lives here expanded.
     The next phase consumes this at OPEN, then collapses it into Handoff history (poda). -->

**Closed phase:** F<N> <name> · **scope:** <epic-id / project> · **date:** YYYY-MM-DD
- Produced: <paths of what this phase built>
- Spec-changes triggered: <what changed backward, or "none">
- Decisions left: <one line, or pointer to Decisions table>
- Input for the next phase: <repos/paths the next phase treats as bounded context>
- Next suggested: <next action — a live command like /flow-build, or "apply the <name> playbook"> <scope>  (<why — which phases are done for this scope>)

## Handoff history

<!-- One line per consumed handoff. A phase appends here when it consumes Current handoff. -->

- F<N> <name> → F<N+1> (consumed YYYY-MM-DD): <one-line what was handed off>

## Artifacts

| Path | What it is | Status | Date |
|---|---|---|---|
| _support/docs/requirements-v2.md | Improved requirements doc | current | YYYY-MM-DD |
| _support/workspace/<report>.html | <one line> | scratch — expires after <event> | YYYY-MM-DD |
| <project>-specs/epics/<id>/ | <epic name> | source of truth | YYYY-MM-DD |

## Decisions

| Decision | Date | Promoted to |
|---|---|---|
| <one-line decision> | YYYY-MM-DD | <project>-specs/decisions/<file>.md (or "pending promotion") |

## Open questions

- [ ] <question> — blocks <what>, owner: <client / user / team>

## Promotion prerequisites

Configs/credentials a dev session found are required BEFORE promoting to an environment —
populated by flow-build close, read at promotion via git (`references/promotion-playbook.md`).
Drop the row once satisfied.

| Environment | Requirement | Status | Discovered |
|---|---|---|---|
| qa / production | <env var / credential / webhook to register> | pending review / configured | YYYY-MM-DD |

## Standing jobs

Recurring, signal-driven work delegated without a per-run handover
(`unattended-autonomy-mode.md > Standing jobs`). A row is the activation key: a scheduled or
background turn runs delegated only under a row whose trigger matches it, else it fails closed.
Max action is always propose-never-land. Edit or drop the row to revoke.

| Job | Trigger | Scope | May do | Expires | Decision log |
|---|---|---|---|---|---|
| <name> | schedule <cron> / PR opened on <repo> / CI red on <branch> / <channel> | <repo(s), paths> | push own branch · open/update PR · comment · open ticket — never merge or promote | YYYY-MM-DD or <condition> | <project>-specs/sessions/standing-<job>/decision-log.md |

## Next session

<1-3 lines: where work should resume, suggested scope.>
```
