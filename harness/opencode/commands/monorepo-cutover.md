---
description: Migrate a multi-repo product to a monorepo (or bootstrap one greenfield) — intake decisions, staged execution to QA, done-checklist
---
Execute the skill `monorepo-cutover` now, with these arguments: $ARGUMENTS

1. Read `~/.agents/skills/monorepo-cutover/SKILL.md` and follow it exactly, starting at
   Step 0 (the stack fork). Argument hint: `[intake | execute | verify] [<workspace>]`.
2. Read `~/.agents/skills/monorepo-cutover/references/cutover-playbook.md` in full before
   the first phase — every rule in it was paid for by a real failure.
3. Candidate tracking is an Engram upsert (`topic_key: migration/<project>-monorepo-cutover`),
   one observation per project, updated at every phase close.
