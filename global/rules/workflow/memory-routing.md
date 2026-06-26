## Memory Routing — Engram vs. Native File-Memory

> Two memory systems run at once: Claude Code's **native file-memory** (harness-injected from `~/.claude/projects/<proj>/memory/`, governed by the built-in `# Memory` protocol) and **Engram** (MCP server, SQLite, injected via its own plugin skill + hooks). Their save-triggers overlap — without a boundary, the same fact lands in both and they drift. This rule draws the boundary. It does NOT configure Engram (its plugin self-manages, as its author intends); it governs which system receives a given save.

### The boundary

- **Engram is the default sink.** All work-record memory goes to Engram via `mem_save`: decisions, bugfixes, discoveries, conventions established, session summaries, project state. Engram is also the **only cross-tool substrate** — one SQLite store serves Claude Code, Codex, and opencode, keyed by a shared project name (see **Workspace project identity** below). Native file-memory is Claude-Code-only and cannot be shared across tools.
- **Native file-memory is the always-hot exception.** Only facts that must be in context *every session without a search* live there: who the user is (`user`), standing behavioral corrections (`feedback`). The harness loads these automatically; Engram requires a retrieval call.
- **Litmus:** "Must this be in context every session without my searching for it?" Yes → native file-memory. No, retrieved when the topic comes up → Engram.
- **Flow-phase artifacts use a deterministic `topic_key`.** When a `flow-*` skill mirrors a phase artifact to Engram for resume-after-compaction, it keys it `flow/{epic-or-project-slug}/{artifact}` so the save upserts, never duplicates — the ledger (`PROJECT.md`) stays the source of truth, Engram is the resume-mirror. Convention defined in `flow-core`.
- **Memories record the session slug.** Any Engram observation about a decision, discovery, or execution records the **slug of the session** that produced the documents — `session: sessions/YYYY-MM-DD-<slug>` in the body, and/or via the `flow/{epic}/{slug}` key. When the memory is recalled, the slug tells you *where the decision/finding documents live* (the versioned session is the durable record; the memory points at it). This is the bridge between Engram recall and the session capture layer (`project-structure.md > Session capture layer`).

### Workspace project identity (multi-repo)

Engram keys memory by a detected project name. The user's canonical layout is a workspace folder with N git repos inside (`project-structure.md`). Single-repo projects need no setup — git-remote detection already gives every tool the same key. **Multi-repo workspaces break that:**

- **At the workspace root** (not a git repo) → multiple git children → `ambiguous`; `mem_context`/`mem_search` without an explicit `project` fail outright.
- **Inside a child repo** → git-remote resolves a per-repo name (`chat-hub-backend`, `chat-hub-frontend`…) → memory fragments across buckets that never see each other.

Fix: declare ONE unified project per workspace via `.engram/config.json` = `{ "project_name": "<group>-<project>" }` (e.g. `acme-chat-hub`, matching `project-structure.md`'s `<group>/<project>`). Detection reads this config *before* any git logic, but only within the enclosing git boundary — it is **never inherited downward**, so placement matters:

- **Workspace root** — covers tools launched there (Claude Code in these workspaces). Required.
- **Each child repo root** — covers tools launched inside one repo (Codex/opencode). Same `project_name` in every repo. Add only for repos you open tools in directly.

Bootstrap both placements with the `engram-init-workspace` skill — it derives `<group>-<project>` from the canonical path and writes them idempotently (never clobbering an existing config). That skill is the standard mechanism; the manual one-liner is only a fallback.

This aligns Engram's bucket with Claude Code's native per-workspace memory (one bucket per workspace), keeping both layers consistent. If a workspace has no config and detection returns `ambiguous`, pass `project` explicitly — never guess a name.

### Anti-double-write

- **Never write the same fact to both systems.** Apply the litmus and pick one: a `user`/`feedback` fact → native only; everything else → Engram only.
- The native `# Memory` protocol and Engram's injected protocol both say "save proactively" on overlapping triggers. This rule is the tiebreaker and wins by specificity: default to Engram, reserve native for the always-hot exception above.

### Invalidation — supersede, don't append

Memory drifts when a fact becomes false and the old version is never invalidated — it gets *corrected by addition* (a new "actually X is done" note) instead of *replaced*. Both layers then hold "X pending" and "X done" at once, and the next session resurfaces the stale one as current — re-reporting finished work as pending, sometimes re-implementing it.

- **Status/pending facts use a deterministic `topic_key` so saves UPSERT, not append.** A project's "current state", "what's pending", or an evolving decision is ONE living observation per topic — `mem_save` with `topic_key` (`status/<area>`, or `flow/{slug}/{artifact}` for flow phases), never a fresh observation each time.
- **When a fact becomes false, invalidate the old record — don't leave stale and corrected both alive.** Engram: `mem_update` the stale observation to restate the current truth — that is what stops it resurfacing (it overwrites, so invalidate only once the new truth is verified). `mem_compare(supersedes)` merely *records* a relation; it does not hide the stale one, so it never substitutes for the update. Native: MOVE a resolved item OUT of "Remaining Work" (to Done) — a struck-through/"DESACTUALIZADO" entry left in a pending section still reads as pending, the anti-pattern. **Existence ≠ completion:** a related file merely existing doesn't prove a *pending task* done — that needs a positive signal (closing commit, passed phase/tests). `/memory-sync` owns the verification.
- **Fix the layer the fact lives in, not a copy in both.** The boundary above keeps each fact in one layer; a status entry deliberately mirrored across both (ledger-mirrored flow state) must be invalidated in both, or one goes current while the other stays stale.

### Reporting state from ground truth

"What's pending / what's next / where are we?" is answered from authority, not memory:

- **The only authoritative source for *implementation* state ("X is done/exists") is the live system** (git/disk, running app, DB). The **ledger** (`PROJECT.md`) is authoritative only for *coordination* state that nothing else records — current flow phase, pointers to the specs repo, the declared tracker; for implementation claims it is a record to verify against the live system, not a substitute (consistent with the next bullet and `debugging.md`). This split is also why the ledger need not be versioned — its authoritative half is regenerable from the versioned layers it points to, or promoted into the specs repo. Memory is a CLAIM to verify against both, never the source.
- **Verify a remembered claim against the live source for its *type* before reporting it.** Implementation ("X is done/exists") → git first: a closing commit/PR (`git log`/`git show`), a merged branch, the file/test on disk. Tool/library behavior or default ("the cooldown is active", version-sensitive) → the tool's authoritative docs **for the installed version** (context7; reconciliation is always the already-installed case, so anchor to the lockfile/manifest version, never latest), never a memory or a misread inspection command. A ledger phase, a closed ticket, or an Engram note is a *record* that corroborates — never a substitute for the authoritative source. **Absence-of-evidence ≠ proof of absence:** an empty/`undefined` inspection result (`config get`, a `grep` miss) doesn't prove "off". A claim that contradicts ground truth is stale: report the *real* state and invalidate it (above), don't echo it.
- **Recurring drift, or "what's pending?" surfacing finished work → run `/memory-sync audit`** to reconcile both layers. At session close, IF work completed this session maps to a still-"pending" memory, supersede it before saving the summary — don't re-save finished work as pending. This is conditional on this-session completions, not an unconditional per-session audit; `/memory-sync` owns the full procedure.

### Tracker sync (governed by project declaration)

Reconciliation uses an issue tracker (Linear, Jira, Monday, GitHub Issues) when the project **declares** one — in its `AGENTS.md`/`CLAUDE.md` or the ledger's `Tracker` field. A declared tracker is consulted without asking each run; the declaration *is* the standing authorization, so no per-pass request is needed. A tracker that is merely *detectable* but undeclared (ticket keys like `PROJ-229`/`BILL-48`, a connected MCP) is NOT consulted on sight: `/memory-sync audit` flags it and proposes adding the declaration (confirm before writing the project config), and subsequent runs then use it automatically. No declared tracker → reconciliation uses local sources only: git/disk, ledger, Engram, native.

- A declared tracker ranks alongside the ledger and above memory — authoritative for tracking state, but like the ledger a status *record*, not the live system: if it says done while git/disk disagrees, live state wins and the ticket is what to correct.
- Reconciliation propagates to a declared tracker: a confirmed-stale "pending" mapped to an open ticket means update/close it. The read is automatic once declared; the write is an outward-facing action — confirm unless the project's config/flow authorizes tracker writes. Use the tracker's MCP/CLI (Linear MCP, `acli` for Jira) via tool-search.

### Promotion

- An Engram observation that proves durably always-relevant (recurs, referenced often) earns a one-line pointer in the native `MEMORY.md` index. This keeps the always-hot layer small and high-signal. Don't pre-emptively promote — let relevance prove itself.
