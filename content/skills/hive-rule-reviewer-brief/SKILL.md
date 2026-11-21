---
name: hive-rule-reviewer-brief
description: "Hive rule leaf: read when briefing a reviewing or verifying child, or rereading a revised plan or a UX finding. Give it the artifact, the requirements and the check output, never the producer's case."
disable-model-invocation: true
---

# Reviewer brief

Read when preparing the brief of a child that reviews or verifies, and when a reviewer's finding is reread after a correction.

Give a reviewer what it needs to judge the work and nothing that argues for it:

- the artifact: the diff, the plan files or the captures;
- the requirements it must meet: the request, the ticket's criteria and the acceptance criteria, and for a UX review the image of the approved example when the work has one;
- the output of the checks already run, as their commands and results.

Never include the producer's reasoning, its account of why the work is right, a claim such as "verified" or "tests pass" without its output, or a list of facts or conclusions the reviewer should not re-check, such as "already settled" or "facts already gathered", since a framing that the work is correct lowers what a reviewer finds. A UX reviewer covers the regions the change alters and their possible impact on the layouts and other components that contain or reuse them, such as containers, lists, dialogs and shared components, measured and exercised on the captures in its brief. For frames in a design file, which have no states to reach, the brief instead gives the file, the node and image of each frame drawn, the nodes of the sibling screens it follows with what they fix, as step 3 of [create-design](../create-design/SKILL.md) lists it, the decisions recorded for the app with the corrections read from the siblings' example records, each frame's width as [draw](../create-design/references/draw.md#width) sets, the design library the stage bound, declared or not, or that it bound none, with the record's inventory and `candidate:` lines and whether the file could use it, the request and research's decisions as requirements, but not the drawer's list of proposed business rules, so the reviewer finds the rules nothing settles on its own; only the brief of the one follow-up review also gives each rule card's node with its frame's and the record's `rule:` lines of those screens, as [rule cards](../create-design/references/rule-cards.md) sets; outside that case, its brief also gives how to reach each state its role always checks, loading, empty, error and the longest real values, such as the fixture route, the account and how to force the error, or says which of them cannot be reached. The brief does the same for the states the measured checks of [criteria](../create-design/references/criteria.md) need: for a list, search or filters, no results apart from an empty catalog, a list of more than 1000 items, and returning from a detail view; for a form or any submit, how to hold the submit pending, such as a delayed stub or a throttled network.

For a design another person drew, to review with nothing to draw, the brief gives the file and nodes or the image, the request when there is one, the app's existing screens and recorded decisions; no sibling list, library inventory, width or how the design was reached.

Reread a revised plan, or a correction of a UX reviewer's findings, with one follow-up to that same child; when its ID is gone, the host has no tool to message it, or the message returns an error, brief a new child of its role with the earlier findings. Of several plan reviewers, those [plan review](../flow-plan/references/plan-review.md) names reread. A UX reread names the corrected regions and states, which it rereads as its role's step 3 sets, with the means to reach them; never run a new full UX review.
