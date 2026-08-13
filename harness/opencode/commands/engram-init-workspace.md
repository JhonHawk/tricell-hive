---
description: Init unified Engram .engram/config.json for a multi-repo workspace (shared memory bucket)
---
Execute the skill `engram-init-workspace` now, with these arguments: $ARGUMENTS

1. Read `~/.agents/skills/engram-init-workspace/SKILL.md` and follow it exactly.
2. It is idempotent: existing conforming configs are reported, never rewritten; the
   bootstrap script writes the workspace root AND each child repo placement.
