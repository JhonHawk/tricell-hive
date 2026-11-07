# Pass the native session identity explicitly

Read only when a `hive view work` command rejects the identity as missing or ambiguous (its error names `--host` and `--native-session-id`). The normal path passes neither flag: the binary reads the host's session variable or the session map that Hive's session hook wrote. The explicit path is the fallback for a host whose hook is not installed, not approved, or not yet running in this session.

Source of the facts below: `_support/docs/architecture/host-session-events.md` in the Hive repository, checked 2026-11-04 against the host versions it lists, and 2026-11-06 for Codex and Grok. A fact marked unverified stays unverified until the installed host's current official documentation confirms it.

1. Name the host with the value the viewer uses: `claude-code`, `codex`, `grok`, `pi`, `opencode`, `cursor` or `antigravity`.
2. Read the one variable named for that host in the table, with `printenv NAME` or the shell's equivalent. Never dump the environment, never guess another variable name, never derive the ID from a file name, an Engram ID or a timestamp.
3. Pass both `--host HOST` and `--native-session-id VALUE` on every `work` command of this work, including `close`. The ID names the real session, so passing it for a work another session opened requires membership the lead accepted; do not substitute the lead's identity.
4. When the variable is empty, unverified, or the row says none, publication is unavailable for this session: report that, tell the user the manager's hook may need an update (`hive update`) or approval (Codex `/hooks`) and a session restart, and continue independent work. Do not synthesize an ID.

| Host | Variable in the agent's own shell tool | Status |
|---|---|---|
| `claude-code` | `CLAUDE_CODE_SESSION_ID` | documented, observed |
| `pi` | `PI_SESSION_ID` | documented |
| `opencode` | `OPENCODE_SESSION_ID` | read from installed code, not run |
| `codex` | `CODEX_SESSION_ID` | read from Codex 0.167 source and observed equal to the hook's session ID; not `CODEX_THREAD_ID`, which names a subagent's own thread |
| `grok` | none found | unavailable |
| `antigravity` | none found | unavailable as a variable: `agy` is resolved from the session map its hook writes (the Antigravity App's variable was not checked) |
| `cursor` | `CURSOR_CONVERSATION_ID` | read from installed code, unverified equality with the hook's conversation ID: use only after Cursor's current documentation confirms it |

A variable reaches only the agent's own shell tool, not commands the user types with `!`. A host started from another host's shell inherits that host's variable; the binary already prefers the variable of the nearest host and ignores inherited ones. When it still rejected the identity as ambiguous because another host's variable or session was also found, pass the flags only when the table's variable for the host you are certain you run in is set.
