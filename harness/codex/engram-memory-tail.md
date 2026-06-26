<!--
Codex-only tail appended to ~/.codex/AGENTS.md by /deploy-global (step 13b).
NOT part of the shared harness/AGENTS.md — opencode and Claude Code inject the
Engram protocol through their own plugins, so appending this there would duplicate.
Codex has no such plugin: this carries the memory protocol into Codex as a
project-doc instruction (additive, invisible in the TUI, survives `engram setup`).
Keep it lean — it loads into every Codex session's context.
-->

## Engram Persistent Memory

Persistent memory via Engram MCP tools (survives sessions and compaction). Save proactively — don't wait to be asked.

Save (`mem_save`) right after: a bug fix, an architecture/design decision, a non-obvious discovery, a config/env change, an established convention, or a learned user preference.
- title: verb + what (searchable) · type: bugfix|decision|architecture|discovery|pattern|config|preference · scope: project (default)|personal
- topic_key for evolving topics — reuse the same key to update instead of duplicating; `mem_suggest_topic_key` if unsure
- content: What (one line) · Why · Where (files) · Learned (gotchas, omit if none)

Search before acting when the user says recall/remember/"qué hicimos", or you start work that may have been done before: `mem_context` first, then `mem_search` (keywords), then `mem_get_observation` for full content.

Session close: before saying done/listo, call `mem_session_summary` with Goal · Instructions · Discoveries · Accomplished · Next Steps · Relevant Files.
