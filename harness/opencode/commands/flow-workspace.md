---
description: Workspace hygiene: custodian audit and approved apply — git state, stray files, broken pointers (flow pack)
---
Execute the flow-pack skill `flow-workspace` now, with these arguments: $ARGUMENTS

1. Read !`echo ~/.agents/skills/flow-core/references/harness-mechanics.md` first — it maps
   Claude Code mechanic names (Agent tool, AskUserQuestion, ToolSearch) to opencode
   equivalents (task tool, inline questions, direct MCP calls).
2. Read !`echo ~/.agents/skills/flow-workspace/SKILL.md` and follow its phases exactly, applying the
   mechanics translation. Argument hint: `[audit | apply]`.
3. Honor the flow contract: orchestrate via subagents, never implement in the main
   thread when the skill says to dispatch.
