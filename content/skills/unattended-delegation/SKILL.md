---
name: unattended-delegation
description: Coordinate explicitly delegated unattended work with a bounded objective and expiry. Use after the user explicitly delegates unattended authority, says they are leaving (for example, going to sleep) while authorized work is running, or a declared scheduled job matches; never activate from silence or task duration.
---

# Explicit unattended delegation

Activate this mode only from an explicit delegation, a user's message that they are leaving while authorized work is running, or a declared scheduled job whose scope and expiry match the current work. A leaving message delegates the objective of the work already running and expires when that objective completes or the user returns.

At activation, state the objective, root or repository, references, destination, evidence, and an expiry or deadline. This skill grants no broader tools or host permissions.

Delegate under the shared delegation rules. Keep consequential decisions and evidence in the existing task record when one exists. Do not require a ledger, branch, hook, or new tracking system solely for this mode.

## Decisions

Decide every question that comes up yourself instead of asking or waiting: take the option you would recommend, record the decision and its reason in the task record, and list each one in the final report with how to revert it.

## Effects

Within the objective, you may take any effect the work needs, including deployments and writes to shared or production systems, when it is reversible or made reversible first and it disrupts no other component beyond a brief planned interruption. Before each such effect, record in the task record the exact target, the change, and how to revert it; this replaces asking for confirmation. For example:

- Enabling or creating a cloud resource that only this work uses is allowed.
- A database migration is allowed after a verified backup, with the service suspended only while it runs.

Irreversible or destructive actions stay forbidden: deleting or overwriting data, resources, credentials, history, or branches without a verified way back, and any change that breaks or interrupts other components or their users beyond that planned interruption. Skip such an action, continue the work that does not depend on it, and report it with the decision it needs. The secret rules apply unchanged.

## End

At expiry or when the objective is complete, stop unattended effects, report completed and unresolved work, every decision taken and every effect applied with how to revert it, and return control to the user. Do not imply activation or success when no matching explicit delegation exists.
