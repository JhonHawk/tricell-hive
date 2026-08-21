---
alwaysApply: true
---

## Memory Routing — Engram vs. Native File-Memory

> Not always-on anywhere: a memory operation has no file behind it, so every harness reaches this through the `memory-policy` skill. The companion half — how to ANSWER a state question — lives in `quality/reporting-integrity.md > Reporting state from ground truth`; this file owns the store mechanics.

> Two memory systems run at once: Claude Code's **native file-memory** (harness-injected, always in context) and **Engram** (MCP server, SQLite, retrieved on demand). Their save-triggers overlap; this rule draws the boundary so the same fact never lands in both and drifts. It does NOT configure Engram (its plugin self-manages); it governs which system receives a given save.

### The boundary

- **Engram is the default sink** for all work-record memory (`mem_save`): decisions, bugfixes, discoveries, conventions, session summaries, project state. It is also the only cross-tool substrate — one store serves Claude Code, Codex, and opencode under a shared project name (below).
- **Native file-memory is the always-hot exception:** only facts that must be in context every session without a search — who the user is (`user`), standing behavioral corrections (`feedback`).
- **Litmus:** "must this be in context every session without my searching for it?" Yes → native. No, retrieved when the topic comes up → Engram.
- **Never write the same fact to both systems.** Both protocols say "save proactively" on overlapping triggers; this rule is the tiebreaker and wins by specificity.
- **Compaction is not total amnesia:** the thread continues with a summary plus the unsummarized recent context, and native memory, the ledger, and git persist on their own — Engram is the searchable cross-session work record, not the only survivor.
- **Flow-phase artifacts and status facts use a deterministic `topic_key`** (`flow/{epic-or-project-slug}/{artifact}`, `status/<area>`) so saves upsert, never duplicate — the ledger stays the source of truth, Engram is the resume-mirror; convention defined in `flow-core`. Observations about decisions/discoveries/executions record the producing session's slug (`session: sessions/YYYY-MM-DD-<slug>`) so recall points at the durable documents (`project-structure.md > Session capture layer`).
- **Git session modes are never memory:** the git mode and branch strategy resolve per session or from the repo's `AGENTS.md` declaration (`git-mechanics.md > Commits`) — never saved to or read from Engram, whatever a save-preference trigger suggests. Safety gates never become preferences: push, merge, production, and data deletion confirm per operation, always.

### Save cadence — batch to close; save now only what would hurt to lose

The plugin-injected protocol wins on *mechanics* (tools, envelopes, judgment flow); this layer governs *when* to save. Its per-task "save immediately" triggers name what is save-worthy — not a command to write mid-task.

- **Mid-session `mem_save` is reserved for facts that would hurt to lose if the session died now:** an architecture/scope decision, a root cause, a new convention, a user correction. Micro-decisions and incremental progress consolidate into `mem_session_summary` and the close-time upserts.
- **One living fact = one upsert.** An evolving fact revisited during the session updates once at close via its `topic_key` — never N observations tracking each intermediate state.
- **Close is ONE consolidated pass.** Close here is the user declaring the session over — not a finished plan, which triggers git hygiene instead (`git-mechanics.md > End-of-work hygiene`). At most one upsert per living `topic_key` plus one `mem_session_summary` per session close; a mid-session save or compaction-forced summary already covering the fact is updated, never re-saved.
- **A save-nudge is satisfied by the next consolidated save** when something durable exists to record; it is not an instruction to invent an observation.
- The close-time invalidation pass (supersede below) is not "too many calls" — it is the designed cost of not accumulating stale memory.

### Workspace project identity (multi-repo)

A multi-repo workspace declares ONE unified Engram project — `.engram/config.json` with `project_name: <group>-<project>`, at the workspace root AND in each child repo. The `engram-init-workspace` skill owns the mechanics and writes both placements idempotently. When detection returns `ambiguous` under a canonical `projects/<group>/<project>/` path: derive the name from the path, pass it as explicit `project`, proceed — and offer the skill once to fix it durably. Non-canonical path with no config → ask; never guess a name.

### Invalidation — supersede, don't append

Memory drifts when a fact becomes false and the old record survives — "corrected by addition" leaves both "X pending" and "X done" alive, and the next session resurfaces the stale one as current.

- **Status/pending facts upsert via `topic_key`** — ONE living observation per topic, never a fresh one each time. State that `/status-fetch` re-derives live is not saved at all: it ages into the drift `/memory-sync` then has to clean.
- **When a fact becomes false, invalidate the old record.** Engram: `mem_update` the stale observation to restate the current truth (it overwrites — invalidate only once the new truth is verified); `mem_compare(supersedes)` merely records a relation and does NOT hide the stale one — never a substitute for the update. Native: MOVE the resolved item out of "Remaining Work" — a struck-through entry left in a pending section still reads as pending. Verify completion against ground truth before invalidating; `/memory-sync` owns the sweep.
- **Fix the layer the fact lives in.** A status deliberately mirrored across both layers (ledger-mirrored flow state) is invalidated in both, or one goes current while the other stays stale.
- **At session close, IF this session's completions map to a still-"pending" memory, supersede it before saving the summary** — conditional on this-session completions, not a per-session ritual.

### Tracker sync (governed by project declaration)

Reconciliation uses an issue tracker (Linear, Jira, Monday, GitHub Issues) only when the project DECLARES one; the declaration IS the standing read authorization, no per-run request. **One home per project shape:** a flow project declares in its ledger (the three fields — `Tracker`, `Tracker access`, `Deferred marker`); a repo without a ledger declares in `AGENTS.md`/`CLAUDE.md`. The declaration's write is always confirm-gated — it grants standing authorization to an external system (unlike the base-branch fact, which rides the session's commits). A merely *detectable* tracker (ticket keys like `TRI-229`, a connected MCP) is not consulted on sight — `/memory-sync audit` flags it and proposes adding the declaration. No declared tracker → local sources only (git/disk, ledger, Engram, native).

- A declared tracker ranks alongside the ledger and above memory — authoritative for tracking state, but a status *record*, not the live system: if it says done while git/disk disagrees, live state wins and the ticket is what to correct.
- **Writes are outward-facing — batch them.** Collect every proposed ticket update/close into the finished plan's close confirmation: one approval covers the batch, never per-ticket asks mid-run. Standing write authorization only when the project's config/flow declares it. Use the tracker's MCP/CLI (Linear MCP, `acli` for Jira) via tool-search.
- **Opening a ticket to capture a pending item is exempt from the batch — reporting it is not.** Work that would otherwise be lost is captured when it is found, not at close: it records state, it does not change any. It surfaces in the plan's close table (`quality/reporting-integrity.md`), never only in the tracker — a ticket the agent invented is one the user has no way to know exists. Closes and status changes stay in the confirmed batch.

### Promotion

An Engram observation that proves durably always-relevant (recurs, referenced often) earns a one-line pointer in the native `MEMORY.md` index. Don't pre-emptively promote — let relevance prove itself.
