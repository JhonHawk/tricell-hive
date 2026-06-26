---
description: Reconcile project memory (Engram + native) against ground truth; invalidate stale pending/status memories. audit | apply
---
Execute the skill `memory-sync` now, with these arguments: $ARGUMENTS

1. Read `~/.agents/skills/memory-sync/SKILL.md` and follow its phases exactly.
2. Translate Claude Code mechanics to opencode equivalents: the Skill tool → this
   command; AskUserQuestion → an inline question; the Explore subagent → the task tool;
   Engram tools (`mem_context`, `mem_update`, `mem_compare`, `mem_delete`, `mem_review`)
   → direct MCP calls.
3. Native file-memory (`~/.claude/projects/...`) is Claude-Code-only — in opencode it does
   not exist, so reconcile the Engram layer (and any opencode-side memory) and skip the
   Claude-native layer. Ground-truth precedence (live state, then ledger) is unchanged.
