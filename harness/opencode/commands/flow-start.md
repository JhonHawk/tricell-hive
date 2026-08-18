---
description: greenfield wizard — intake, workspace bootstrap and technical foundation (flow pack)
---
Execute the flow-pack skill `flow-start` now, with these arguments: $ARGUMENTS

1. Read !`echo ~/.agents/skills/flow-core/references/harness-mechanics.md` first — it maps
   Claude Code mechanic names (Agent tool, AskUserQuestion, ToolSearch) to opencode
   equivalents (task tool, inline questions, direct MCP calls).
2. Read !`echo ~/.agents/skills/flow-start/SKILL.md` and follow its phases exactly, applying the
   mechanics translation. Argument hint: `[<group> <project>]`.
3. Honor the flow contract: orchestrate via subagents, never implement in the main
   thread when the skill says to dispatch.
