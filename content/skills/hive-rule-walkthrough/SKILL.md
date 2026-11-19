---
name: hive-rule-walkthrough
description: "Hive rule leaf: read when a build or verify captures an interface, when a flow or role publishes captures, or when the user asks for a walkthrough or demo of a built interface."
disable-model-invocation: true
---

# Publish captures and walk an interface

Read when a build or verify captures an interface, when a flow or role publishes captures, and when the user asks for a walkthrough or demo of a built interface, which follows this leaf alone. The close of a work with captures follows [viewer publication](../hive-rule-viewer-close/SKILL.md).

The viewer holds only UI captures and their iteration. A build or verify flow that was explicitly invoked (its command typed, its route accepted, research continuing into build at the user's request, plan continuing into build as its handoff sets, or build continuing into verify), and a walkthrough or demo the user asks for, gets a viewer session at its first UI capture; research and plan publish nothing, and a flow without captures creates no viewer work. At that first capture, the primary producer reads [viewer publication](../hive-rule-viewer/SKILL.md) once per session, and again after a compaction, and follows it for each batch of captures and at close. Children return captures to that producer. If publication is unavailable, continue independent work and report what was not published.

For a UI walkthrough, serve the candidate locally with seeded data of realistic volume and text length, or with stub data only for the first view of a build, as [flow-plan](../hive-rule-ui-plan/SKILL.md) defines, and an account that reaches the changed screens, chosen after reading [test accounts](../hive-rule-test-accounts/SKILL.md). Read [UI verification](../hive-rule-ui-verification/SKILL.md) for the proportional capture scope and reuse existing evidence. Publish the current captures using the publication procedure, then give the user the actual target, steps, expected result, observations and available evidence paths. The calling flow owns its existing acceptance question and handles corrections; this reference adds only the open-viewer option that [viewer publication](../hive-rule-viewer-open/SKILL.md) sets. On a correction, capture again only the affected screens, under their same titles: the viewer numbers the new iteration and keeps the earlier one.
