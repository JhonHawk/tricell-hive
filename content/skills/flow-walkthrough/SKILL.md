---
name: flow-walkthrough
description: Publish substantive work stages, documents and UI evidence to the local viewer, and present a built interface for human validation. Use from research, plan, build, verify and close, or when the user requests a walkthrough or demo.
---

# Publish work and walk an interface

A flow gets a viewer session when it was explicitly invoked (its command typed or its route accepted) and did not end as `flow-research`'s one bounded lookup. For such a flow, the primary producer reads and follows [viewer publication](references/viewer-publication.md) at the start, at each stage change, new capture or document snapshot, and at close. Children return evidence to that producer. A simple factual reply does not create a session. If publication is unavailable, continue independent work and report what was not published.

For a UI walkthrough, serve the candidate locally with stub or seeded data of realistic volume and text length, as [flow-plan](../flow-plan/SKILL.md#ui-decisions) defines, and an account that reaches the changed screens. Read [UI verification](../flow-verify/references/ui-verification.md) for the proportional capture scope and reuse existing evidence. Publish current captures and instructions using the publication procedure, then give the user the actual target, steps, expected result, observations and available evidence paths. The calling flow owns its existing acceptance question and handles corrections; this skill returns to it without adding a question. On a correction, publish only affected evidence and clearly distinguish it from earlier evidence.

Only when the `hive` binary is absent and a UI walkthrough is needed, read [the legacy manifest](references/walkthrough-manifest.md) and produce its standalone HTML fallback. A stopped or incompatible service is not a missing binary and does not select that fallback.
