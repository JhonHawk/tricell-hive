
## Memory Routing — Engram vs. Native File-Memory

> Two memory systems run at once: Claude Code's **native file-memory** (harness-injected, always in context) and **Engram** (MCP server, SQLite, retrieved on demand). Their save-triggers overlap; this rule draws the boundary so the same fact never lands in both and drifts. It does NOT configure Engram (its plugin self-manages); it governs which system receives a given save.

### The boundary

- **Engram is the default sink** for all work-record memory (`mem_save`): decisions, bugfixes, discoveries, conventions, session summaries, project state. It is also the only cross-tool substrate — one store serves Claude Code, Codex, and opencode under a shared project name (below).
- **Native file-memory is the always-hot exception:** only facts that must be in context every session without a search — who the user is (`user`), standing behavioral corrections (`feedback`).
- **Litmus:** "must this be in context every session without my searching for it?" Yes → native. No, retrieved when the topic comes up → Engram.
- **Never write the same fact to both systems.** Both protocols say "save proactively" on overlapping triggers; this rule is the tiebreaker and wins by specificity.
- **Compaction is not total amnesia:** the thread continues with a summary plus the unsummarized recent context, and native memory, the ledger, and git persist on their own — Engram is the searchable cross-session work record, not the only survivor.
- **Flow-phase artifacts and status facts use a deterministic `topic_key`** (`flow/{epic-or-project-slug}/{artifact}`, `status/<area>`) so saves upsert, never duplicate — the ledger stays the source of truth, Engram is the resume-mirror; convention defined in `flow-core`. Observations about decisions/discoveries/executions record the producing session's slug (`session: sessions/YYYY-MM-DD-<slug>`) so recall points at the durable documents (`project-structure.md > Session capture layer`).

### Workspace project identity (multi-repo)

A multi-repo workspace declares ONE unified Engram project — `.engram/config.json` with `project_name: <group>-<project>`, at the workspace root AND in each child repo. The `engram-init-workspace` skill owns the mechanics and writes both placements idempotently. When detection returns `ambiguous` under a canonical `projects/<group>/<project>/` path: derive the name from the path, pass it as explicit `project`, proceed — and offer the skill once to fix it durably. Non-canonical path with no config → ask; never guess a name.

### Invalidation — supersede, don't append

Memory drifts when a fact becomes false and the old record survives — "corrected by addition" leaves both "X pending" and "X done" alive, and the next session resurfaces the stale one as current.

- **Status/pending facts upsert via `topic_key`.** A project's current state or an evolving decision is ONE living observation per topic — never a fresh observation each time.
- **When a fact becomes false, invalidate the old record.** Engram: `mem_update` the stale observation to restate the current truth (it overwrites — invalidate only once the new truth is verified); `mem_compare(supersedes)` merely records a relation and does NOT hide the stale one — never a substitute for the update. Native: MOVE the resolved item out of "Remaining Work" — a struck-through entry left in a pending section still reads as pending. **Existence ≠ completion:** a related file merely existing doesn't prove a pending task done; that needs a positive signal (closing commit, passed phase/tests). `/memory-sync` owns the verification.
- **Fix the layer the fact lives in.** A status deliberately mirrored across both layers (ledger-mirrored flow state) is invalidated in both, or one goes current while the other stays stale.

### Reporting state from ground truth

"What's pending / what's next / where are we?" is answered from authority, not memory:

- **Implementation state ("X is done/exists") is authoritative only in the live system** (git/disk, running app, DB). The ledger (`PROJECT.md`) is authoritative only for coordination state nothing else records — current flow phase, pointers, declared tracker; for implementation claims it is a record to verify against the live system, never a substitute. Memory is a CLAIM to verify against both.
- **Verify a remembered claim against the live source for its type before reporting it.** Implementation → git first (closing commit/PR, merged branch, the file/test on disk). Tool/library behavior or defaults → the docs for the INSTALLED version (context7 anchored to the lockfile/manifest, never latest), never a memory or a misread inspection command. **Absence of evidence ≠ proof of absence:** an empty `config get` or a grep miss doesn't prove "off". A claim contradicting ground truth is stale — report the real state and invalidate it (above).
- **Recurring drift, or "what's pending?" surfacing finished work → run `/memory-sync audit`.** At session close, IF this session's completions map to a still-"pending" memory, supersede it before saving the summary — conditional on this-session completions, not a per-session ritual.

### Tracker sync (governed by project declaration)

Reconciliation uses an issue tracker (Linear, Jira, Monday, GitHub Issues) only when the project DECLARES one — in its `AGENTS.md`/`CLAUDE.md` or the ledger's `Tracker` field; the declaration IS the standing read authorization, no per-run request. A merely *detectable* tracker (ticket keys like `TRI-229`, a connected MCP) is not consulted on sight — `/memory-sync audit` flags it and proposes adding the declaration. No declared tracker → local sources only (git/disk, ledger, Engram, native).

- A declared tracker ranks alongside the ledger and above memory — authoritative for tracking state, but a status *record*, not the live system: if it says done while git/disk disagrees, live state wins and the ticket is what to correct.
- **Writes are outward-facing — batch them.** Collect every proposed ticket update/close into the session-close confirmation: one approval covers the batch, never per-ticket asks mid-run. Standing write authorization only when the project's config/flow declares it. Use the tracker's MCP/CLI (Linear MCP, `acli` for Jira) via tool-search.

### Promotion

An Engram observation that proves durably always-relevant (recurs, referenced often) earns a one-line pointer in the native `MEMORY.md` index. Don't pre-emptively promote — let relevance prove itself.
