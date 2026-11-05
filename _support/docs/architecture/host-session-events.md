# Host session events and identity

Verified 2026-11-04 for the `viewer-passive-capture` change (task T0). Each fact is marked **doc** (official documentation, URL or local doc path), **code** (strings or source of the installed version, not run), **obs** (observed on the maintainer's machine) or **inf** (inferred). Nothing here was verified by running an agent session. Recheck before relying on a fact for a newer host version.

| Host and version | Hook or plugin location | Session ID in the hook | Session ID in the agent's own shell tool | Native task events |
|---|---|---|---|---|
| Claude Code 2.1.318 | `~/.claude/settings.json` `hooks.<Event>[]` (doc: code.claude.com/docs/en/hooks) | stdin `session_id`, `cwd`, `transcript_path` (doc) | `CLAUDE_CODE_SESSION_ID`, updated on `/clear` (doc: /env-vars; obs) | `TaskCreated` (on `TaskCreate` only) and `TaskCompleted`, with `task_id`, `task_subject` (doc); none with legacy `TodoWrite` |
| Codex CLI 0.166.0 | `~/.codex/hooks.json`, same shape (doc: developers.openai.com/codex/hooks) | stdin `session_id` (the parent's in subagent hooks; subagents carry `agent_id`), `cwd`, `transcript_path` (doc) | `CODEX_THREAD_ID` (code, unverified) | none (doc); `PostToolUse` on `update_plan` is the nearest |
| Grok Build 1.0.74 | any file in `~/.grok/hooks/`, merged (doc: `~/.grok/docs/user-guide/10-hooks.md:64`) | stdin camelCase `sessionId`, `cwd`, `workspaceRoot`; env `GROK_SESSION_ID`, `GROK_HOOK_EVENT` (doc) | none found | none (doc) |
| OpenCode 2.2.5 | plugin in `~/.config/opencode/plugins/`, auto-loaded and hot-reloaded, run by the shared background service (obs: opencode.log) | `event.data.sessionID`, `event.location.directory` from `ctx.event.subscribe` (code) | `OPENCODE_SESSION_ID` in the LLM bash tool (code) | no `todo.updated` in 2.2.5 (code); todos from `ctx.tool.hook("execute.after")` when `tool === "todowrite"`, `input.todos[]` with `content`, `status`, no ID |
| Pi 1.5.0 | extension in `~/.pi/agent/extensions/*.ts`, auto-loaded (doc: pi docs `extensions.md:127-131`) | `ctx.sessionManager.getSessionId()`, `ctx.cwd` (doc) | `PI_SESSION_ID`, `PI_SESSION_FILE`, not in user `!` commands (doc: `environment-variables.md:26-27`) | none (doc) |
| Cursor CLI 2026.10.22 | `~/.cursor/hooks.json` `{"version":1,"hooks":{…}}` (code; CLI use not documented) | `conversation_id`, `workspace_roots`, `transcript_path`; `session_id` on `sessionStart`/`sessionEnd` (doc, code) | `CURSOR_CONVERSATION_ID` (code, unverified equality with the hook field) | none (doc) |

## Constraints the design must respect

- **Cross-host loading.** Grok (`compat.claude.hooks`, default on) and Cursor CLI also run the hooks in `~/.claude/settings.json`, with their own payload shapes (code, obs for Grok in `~/.grok/logs/hooks.log`). A hook command registered for Claude Code must detect the host that actually runs it (Grok: `GROK_HOOK_EVENT` or a `sessionId` key; Cursor: `cursor_version`, `generation_id`, `workspace_roots`, or a `conversation_id` without `session_id`; never the inherited variables `CURSOR_VERSION`/`CURSOR_PROJECT_DIR`, which a Claude Code or Codex session started from a Cursor terminal also carries, so a `sessionStart` or `sessionEnd` with both IDs and no `cursor_version` follows the hook's host argument) and act for that host or exit silently.
- **Silence.** Plain stdout enters the model context on Claude Code `SessionStart`, `UserPromptSubmit`, `UserPromptExpansion` and `PostModelSwitch` (doc), on Codex `SessionStart` and `UserPromptSubmit` (doc), and on Grok `PostToolUse` (doc). Codex `Stop` and `SubagentStop` expect JSON on stdout; do not register them.
- **Time budgets.** Claude Code `SessionEnd` shares a 1.5 s budget unless an entry sets a larger `timeout` (doc). Grok's default hook timeout is 5 s (doc). Codex `SessionEnd` fires only for the main thread, on close, archive or 30 minutes idle (doc).
- **Trust.** Claude Code holds back settings hooks in an untrusted folder until the workspace trust dialog is accepted; `-p` counts the folder as trusted (doc). Codex skips new or changed hooks until the user approves them in `/hooks`; `codex exec` has `--dangerously-bypass-hook-trust` (doc, code). Grok holds back user hooks in untrusted folders; the home directory counts as trusted (doc).
- **Identity by process.** OpenCode plugins run in one shared service, so a process-ancestry map cannot tell its sessions apart: the plugin passes the session ID explicitly. Cursor `sessionStart` does not fire when a chat resumes (code), so the map also needs `beforeSubmitPrompt`. Grok and Codex subagents share the parent's process.
- **Filtering.** Claude Code `SubagentStart`/`SubagentStop` also fire for internal agents with an empty `agent_type` (doc).
- **Shell-tool variables** reach only the agent's own shell tool, not commands the user types with `!` (doc for Pi, code for the others).

## Session map

Contract between the hook that writes the process-to-session map (T6) and the resolver that reads it (`tooling/cli/view_identity*.go`, T2). A change to one side changes this section.

**Location.** `<state>/hooks/sessions/<pid>.json`, where `<pid>` is the host CLI's own process, never the hook's `sh -c` wrapper or its short-lived children: name the file after the nearest ancestor of the hook whose process is the host (for Claude Code the `claude` process, for Codex `codex`, and so on). A hook that cannot tell which ancestor is the host writes nothing, and the walk stops at the nearest ancestor that is any host CLI: a hook registered for Claude Code and run by Grok finds `grok` first, so it writes no map for `claude-code` and never one on a process further up. The resolver walks up from its own process and uses the nearest ancestor that has a map, so the file must sit on a process that is an ancestor of the agent's shell.

**Schema.**

```json
{"host": "claude-code", "pid": 4242, "process_started_at": "2026-11-04T08:00:00Z", "changed_at": "2026-11-04T08:00:00Z",
 "sessions": [{"session_id": "...", "started_at": "2026-11-04T08:00:00Z", "ended_at": "2026-11-04T09:00:00Z"}]}
```

- `host` is the canonical host name the viewer uses: `claude-code`, `codex`, `grok`, `opencode`, `pi` or `cursor`. `pid` must equal the file name. All times are RFC 3339.
- `process_started_at` is the start time of the host process, read once by the writer (`ps -o lstart=` or equivalent). The reader compares it with the live process's start time and treats a difference of more than 2 s as no map, so a pid reused after a crash (a crashed host never writes `ended_at`) does not resolve a dead session. A map without this field is rejected.
- A session is live when it has no `ended_at` and `started_at >= changed_at`. Exactly one live session resolves; none or several is ambiguous and the command is rejected.
- `changed_at` is set to the first session's `started_at` when the file is created and moves only when a session **replaces** another in the same process: `/clear`, resume into another session, or the end of the last live session. Ending one session while another live session shares the process leaves it. The replacing session's `started_at` equals the new `changed_at`. It does not move when a second session starts alongside the first (a Codex subagent thread, several OpenCode sessions), so two live sessions stay detectable as ambiguous.
- `ended_at` is written on end; the writer never infers it from the process dying.

**Caps and modes.** At most 64 entries in `sessions` (the writer drops the oldest ended entries first) and at most 1 MiB per file. Files are written atomically (temporary file in the same folder, then rename) at mode `0600`; `hooks/` and `hooks/sessions/` are `0700`. The reader rejects, as ambiguous, a map file or either folder that is a symlink, is not owned by the current user or is group- or world-writable, and opens the file with no-follow semantics. The same check applies to `publisher/v1/current/` and `current/<host>/` before any current-work file is read or written.

**Variables versus the map.** A host variable (`CLAUDE_CODE_SESSION_ID`, `PI_SESSION_ID`, `OPENCODE_SESSION_ID`) is accepted only if the nearest ancestor that has a valid map, or that is a known host CLI process (`claude`, `codex`, `grok`, `opencode`, `pi`, `cursor-agent`), belongs to the same host; otherwise a nested host that inherited the outer variable would publish as the outer session. Cursor has no variable source: `CURSOR_CONVERSATION_ID` equality with the hook's `conversation_id` is unverified, so Cursor resolves through the map.

## Reporting rules (`hive hook`)

- **Budget.** The whole run is bounded at 2 s; whatever is unfinished stays in the queues.
- **`occurred_at`.** A task event takes a time that is the same on every run and never the hook's clock (the host's timestamp, the transcript's modification time, the host process start, then the epoch). An activity-only event (prompt, stop, subagent, session start or end) takes the host's timestamp, then the transcript's modification time, then the hook's clock, because the viewer stamps last activity with the event's time (never later than its own clock, never backwards) and applying activity twice only moves it forward.
- **`data.turn_id`.** An activity-only event carries the host's prompt or turn ID when the payload has one (Grok `promptId`, Cursor `generation_id`, Claude Code `prompt_id`/`turn_id`, Codex `turn_id`/`prompt_id`), so two events of one session are two operations. OpenCode and Pi envelopes carry none and rely on the time. A task event carries none.
- **Binding.** Every event names the session's one current work and its active stage when they resolve; a rejected binding is retried once without it.
- **Task events and the work's age.** The viewer ignores a native task event whose `occurred_at` is before the work was created, and `reconcile` skips transcript tasks older than the current work: they belong to earlier work in the same session. A task event whose time is only a stand-in (the host process start, because the payload and transcript give none) can therefore be older than the work; Claude Code and OpenCode always give a better time.
- **Reconcile reads.** A larger session file is read from its last part (16 MiB from the command, 4 MiB from the session-start hook), from the first whole line; the result says `partial` when it did. Activity is sent whatever its age.
