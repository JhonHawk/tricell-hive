---
description: Multi-lens preventive audit of runtime repos — parallel readers, refuted inventory, HTML report (flow pack)
---
Execute the flow-pack skill `flow-audit` now, with these arguments: $ARGUMENTS

1. Read !`echo ~/.agents/skills/flow-core/references/harness-mechanics.md` first — it maps
   Claude Code mechanic names (Agent tool, AskUserQuestion, ToolSearch) to opencode
   equivalents (task tool, inline questions, direct MCP calls).
2. Read !`echo ~/.agents/skills/flow-audit/SKILL.md` and follow its phases exactly, applying the
   mechanics translation. Argument hint: `[full | smells | security | db | perf | arch | secrets] [<repos…>]`.
3. Honor the flow contract: orchestrate via subagents, never implement in the main
   thread when the skill says to dispatch.
