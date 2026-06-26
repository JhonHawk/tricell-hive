---
description: F2 — bootstrap a client project workspace: structure, ledger, ambient AGENTS.md/CLAUDE.md pair (flow pack)
---
Execute the flow-pack skill `flow-kickoff` now, with these arguments: $ARGUMENTS

1. Read `~/.agents/skills/flow-core/references/harness-mechanics.md` first — it maps
   Claude Code mechanic names (Agent tool, AskUserQuestion, ToolSearch) to opencode
   equivalents (task tool, inline questions, direct MCP calls).
2. Read `~/.agents/skills/flow-kickoff/SKILL.md` and follow its phases exactly, applying the
   mechanics translation. Argument hint: `[group] [project]`.
3. Honor the flow contract: orchestrate via subagents, never implement in the main
   thread when the skill says to dispatch.
