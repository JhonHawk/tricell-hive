---
name: status-fetch
description: >
  Fetch the project's live external state — git across repos, the declared tracker's
  board, open PRs, deploy jobs — in an isolated subagent, returning compact facts.
  Use for "what's pending", "what's next", "where are we", or before a status report.
  Returns facts to interpret, never a decision about what to work on.
---

# status-fetch

Gather live state in a forked subagent so the sweep — often tens of thousands of tokens
of tracker pages, git output and PR listings — never enters the main session. Only the
compact report returns.

**This runs in a fresh subagent with no conversation history** — on Claude Code by
frontmatter, elsewhere because the caller dispatches it. Everything comes from `$ARGUMENTS`
and disk; report facts, and the main thread decides what they mean.

## 1 — Resolve scope from disk

- `$ARGUMENTS` names a repo → scope to it. `all` or empty → every git repo at or below cwd
  (workspace root: each child repo; single repo: itself).
- **Ledger** — `_support/PROJECT.md` at or above cwd. Read it if present: it declares the
  `Tracker`, `Tracker access`, and `Deferred marker` fields. Absent → local sources only.
- **Tracker authorization is the declaration.** No declared tracker → do NOT query Linear,
  Jira, or any tracker MCP; report `tracker: no declarado` and use git and disk alone.
  A ticket key in a branch name is not a declaration.

## 2 — Fetch, in parallel where the calls are independent

Per repo: current branch, ahead/behind vs upstream, working-tree cleanliness, last commits,
other local branches, stashes, worktrees.

Beyond one repo, or when the tracker is declared, also: open PRs and their check state
(`gh pr list`, `gh pr checks`), the tracker's live tickets, and deploy jobs for branches
that deploy an environment.

**Environment branches resolve against `origin/<branch>`, never the local checkout** — an
absent LOCAL branch is not an absent environment. Report each pair's divergence as counts
(`development↔qa`, `qa↔production`), plus each environment's head SHA; prose without numbers
is not a divergence report.

**Batch the reads.** One command per repo, or one loop over all of them — never a round-trip
per fact.

**Query the tracker by non-terminal state — one call per state, all dispatched in a single
parallel batch.** One unfiltered read comes back ordered by recency and truncates with the
open backlog outside the page. The failure mode to avoid is the SEQUENTIAL refilter — board
call after board call until something looks right — never the one parallel fan-out. A
truncated response (page cap, `hasNextPage`) is declared as truncated: a page's count is
never reported as a total.

## 3 — Report facts, not conclusions

A compact table per dimension (repos · tracker · PRs · deploys), then at most three lines
of anomalies — a ledger claim contradicted by live state, a branch diverged from its
upstream, a ticket marked done whose code is absent.

**Cross-check every checkable claim the ledger makes, not only its ticket states** — the
deployed SHAs, versions, and environment heads it records are the ones that go stale first.
Compare them against the live heads fetched above and name the mismatch.

- **Empty is a finding.** Write `ninguno` explicitly; an omitted row reads as unchecked.
- **Say what you could not reach.** An unauthorized tracker, a failed `gh` call, an
  unreadable repo: name it. Silence about a gap reports it as absence.
- **No recommendations, no priorities, no next steps.** Deciding what work exists and what
  to do about it belongs to the main thread. Ranking tickets here is out of scope.
- Distinguish a **record** (ledger, tracker) from **live state** (git, deploys); on
  conflict, live state wins and the record is what needs correcting.

Keep the whole report under ~80 lines. It is the only thing that survives this fork.
