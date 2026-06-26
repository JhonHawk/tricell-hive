---
description: F3 — specs repo lifecycle: init, epic drafting with Linear sync, pre-implementation review gate (flow pack)
---
Execute the flow-pack skill `flow-specs` now, with these arguments: $ARGUMENTS

1. Read `~/.agents/skills/flow-core/references/harness-mechanics.md` first — it maps
   Claude Code mechanic names (Agent tool, AskUserQuestion, ToolSearch) to opencode
   equivalents (task tool, inline questions, direct MCP calls).
2. Read `~/.agents/skills/flow-specs/SKILL.md` and follow its phases exactly, applying the
   mechanics translation. Argument hint: `[init | epic <name> | review <path>]`.
3. Honor the flow contract: orchestrate via subagents, never implement in the main
   thread when the skill says to dispatch.
