---
name: unattended-delegation
description: Coordinate delegated unattended work with a bounded objective and expiry. Use after the user explicitly delegates unattended authority, says they are leaving (for example, going to sleep) while authorized work is running, falls silent after answering the route, delivery and review questions of the running work, or a declared scheduled job matches; never activate from task duration alone or before those questions are answered.
---

# Unattended delegation

Activate this mode from one of four sources: an explicit delegation; a user's message that they are leaving while authorized work is running; a declared scheduled job whose scope and expiry match the current work; or the user's silence after they answered the route, delivery and review questions of the current work, which the user decided on 2026-10-18 counts as delegating that work. Each source delegates the objective of the work already authorized, up to the delivery point the answers chose, and expires when that objective completes or the user sends any other message than the ones below, such as one that changes the objective, the scope or an answer in force, asks anything other than how the work is going, gives a new instruction, reports something new, or tells you to stop. A message in any language that only tells you to continue, acknowledges progress, or asks how the work is going keeps the delegation for the same objective and expiry: answer a status question in one line, say that the delegation continues, and keep deciding. Task duration alone, silence before those questions are answered, and a background or scheduled turn without a matching declared job activate nothing.

At activation, state the source and quote the message or answer it rests on, then the objective, root or repository, references, destination, evidence, and an expiry or deadline. This skill grants no broader tools or host permissions.

Before delegating or starting an external-work wait, read [delegation and waits](../flow-research/references/delegation.md). Before resolving needed settings, read [project settings](../flow-plan/references/project-settings.md); before writing or cleaning support artifacts, read [artifact layout and retention](../workspace-conventions/references/artifact-layout.md); before reporting, read [reporting](../flow-close/references/reporting.md). Keep consequential decisions and evidence in the existing task record when one exists. Do not require a ledger, branch, hook, or new tracking system solely for this mode.

## Decisions

Decide every question that comes up yourself instead of asking or waiting: take the option you would recommend, record the decision and its reason in the task record, and list them in the final report as [reporting](../flow-close/references/reporting.md#report-a-running-or-completed-item) sets.

## Effects

Within the objective, you may take any effect the work needs, including deployments and writes to shared or production systems, when it is reversible or made reversible first and it disrupts no other component beyond a brief planned interruption. Before each such effect, record in the task record the exact target, the change, and how to revert it; this replaces asking for confirmation. For example:

- Enabling or creating a cloud resource that only this work uses is allowed.
- A database migration is allowed after a verified backup, with the service suspended only while it runs.
- A force-push that the objective needs, such as realigning a pull-request branch, is allowed with `--force-with-lease` on the exact remote commit and the previous commit recorded as the way back; this is the one exception to the explicit-authorization rule of [git-workflow](../git-workflow/SKILL.md), and a lease that fails because someone else pushed stops the push, as that skill sets.
- Editing the title or body of a pull request, or replying on its review threads, is allowed with the previous text saved in the task record.

Irreversible or destructive actions stay forbidden: deleting or overwriting data, resources, credentials, history, or branches without a verified way back, and any change that breaks or interrupts other components or their users beyond that planned interruption. Skip such an action, continue the work that does not depend on it, and report it with the decision it needs. The secret rules apply unchanged.

## End

At expiry or when the objective is complete, stop unattended effects, report completed and unresolved work, every effect applied with how to revert it, and the decisions taken, listed as the reporting reference sets for a completion report, and return control to the user. Do not imply activation or success when none of the four sources applies, and never record a delegation the user did not give: an inferred activation is reported as inferred, with the answer it rests on.
