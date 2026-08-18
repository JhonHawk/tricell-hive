---
description: N independent generators + adversarial cross-exam + canon synthesis for a consequential question, claim, or design decision
---
Execute the skill `adversarial-research` now, with these arguments: $ARGUMENTS

1. Read !`echo ~/.agents/skills/adversarial-research/SKILL.md` and follow its phases exactly.
2. Translate Claude Code mechanics to opencode equivalents: the Skill tool → this
   command; generator and refuter subagents (Agent tool) → the task tool, dispatching
   all generators in one parallel batch; AskUserQuestion → an inline question.
3. `codex:codex-rescue` is a Claude-Code-plugin-only agent — in opencode it does not
   exist, so fill every generator slot with local agents seeded with distinct
   perspectives per the skill's degradation rule, and note the substitution in the
   synthesis. Refuter budget, independence rules, and ground-truth precedence are
   unchanged.
