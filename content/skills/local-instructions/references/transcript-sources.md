# Transcript sources

Read when `extract_candidates.py` reports zero sessions for a host you used, or when the user asks which hosts the skill covers. The script finds the files itself; this page records where they are and how each host names a project, so a missing host can be explained instead of guessed.

| Host | Files | How the project shows | Covered by the script |
| --- | --- | --- | --- |
| Claude Code | `<Claude home>/projects/<folder>/*.jsonl`; subagent transcripts sit in `<folder>/<session>/subagents/` | The folder is the working directory with `/` and `.` replaced by `-`, such as `-Users-me-work-shop`; a worktree under `.claude/worktrees/<name>` gets its own folder, such as `-Users-me-work-shop--claude-worktrees-w1`. Each record carries `cwd`, which the script matches. | Yes, top-level sessions only |
| Codex | `<Codex home>/sessions/YYYY/MM/DD/rollout-*.jsonl` | The first record, `session_meta`, carries `payload.cwd`. | Yes |
| Grok | `<Grok home>/sessions/<folder>/<session>/updates.jsonl` | The folder is the working directory URL-encoded, such as `%2FUsers%2Fme%2Fwork%2Fshop`. | No: not covered |
| Pi | `<Pi home>/sessions/<folder>/*_<session>.jsonl` | The folder is the working directory with `/` replaced by `-` and wrapped in `--`, such as `--Users-me-work-shop--`. | No: not covered |
| OpenCode, Cursor, Antigravity | Not located | — | No: not covered |

The homes default to `~/.claude`, `~/.codex`, `~/.grok` and `~/.pi/agent`, unless `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `GROK_HOME` or `PI_CODING_AGENT_DIR` names another folder; the script reads the first two. The file layouts come from the Hive viewer's session reconciliation, and the folder encodings from folders observed on one machine (2026-11-07).

Folder names are only a hint, since the encoding is lossy (`/a/shop.app` and `/a/shop-app` share a folder): the script decides by each session's recorded working directory, which must equal a root or start with the root followed by `/`. A Codex session whose `session_meta` source is a subagent is left out. The script reads a session started in a root passed with `--root`, below it or in a folder above it, and counts each command where it ran: the directory recorded for that command, moved by any `cd` earlier in the same command line. A command run outside every root is left out. Pass the repository root, and also the declared workspace root when facts of several repositories matter.
