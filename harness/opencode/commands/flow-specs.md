---
description: specs repo lifecycle: init (product map + structure), epic/revise drafting as deltas with tracker sync, business review gate (flow pack)
---
Execute the flow-pack skill `flow-specs` now, with these arguments: $ARGUMENTS

1. Read !`echo ~/.agents/skills/flow-core/references/harness-mechanics.md` first — it maps
   Claude Code mechanic names (Agent tool, AskUserQuestion, ToolSearch) to opencode
   equivalents (task tool, inline questions, direct MCP calls).
2. Read !`echo ~/.agents/skills/flow-specs/SKILL.md` and follow its phases exactly, applying the
   mechanics translation. Argument hint: `[init | epic <name> | revise <epic-ref> | review <spec-ref>]`.
3. Honor the flow contract: orchestrate via subagents, never implement in the main
   thread when the skill says to dispatch.
