# Offer the Figma design stage only with a known Figma file

## Why

In a real session (Claude Code, Opus 5.5, session `db6a272b`, 2026-11-15, project project-f), research recommended the Figma design stage for 4–5 new screens. Its only basis was that the Figma tools appeared in the host's deferred tool list. The app had no `DESIGN.md`, and the session had no record of any Figma file for the project. The user then had to supply the file. The user decided that the stage is not offered when no memory or other source identifies the project's Figma file.

This adds a condition to an existing rule at the user's decision. It needs no recurrence.

## What changes

- `flow-research` «A screen to design»: the offer also requires a known Figma file. The file is known when the user named it in the conversation, the app's `DESIGN.md` declares it, or the project's persistent memory records it. The block names that file and puts its link in the command's argument. Without a known file, research does not mention the stage, the same as without a connection.
- `flow-figma-design` step 3: the file the offering block named counts like a file the user named.
- Site pages Research and Figma design, and the `workflow-lifecycle` spec, gain the same condition and a scenario for a connection with no known file.

A typed `/flow-figma-design`, and a request whose deliverable is the design, are unchanged: the stage still asks for the file.

## Places checked

`content/guidance/global.md` (flow transitions only, unchanged), `content/skills/flow-research/SKILL.md`, `content/skills/flow-figma-design/SKILL.md`, `content/skills/flow-plan/SKILL.md`, `content/skills/unattended-delegation/SKILL.md`, `site/src/content/docs/flows/{research,figma-design,overview,plan}.md`, `_support/openspec/specs/workflow-lifecycle/spec.md`, `tests/` (no detector names the offer). `global.md` size unchanged.
