---
title: Verify
description: Check a built candidate at a level you can see and change, then integrate it as your delivery settings allow.
---

Verify is the step between a built change and a finished one. Build hands over a candidate; Verify gathers the evidence that it meets what you asked for and says what was not checked.

## When to use

Use Verify when a candidate exists and needs evidence before it is accepted. It follows Build in the usual order, and you can also start here with a change that is already built.

Example requests:

```text wrap
Verify the change on this branch and integrate it if everything passes.
Full verify, including the UI review.
Skip verify for this one; I will check it myself.
```

## What it produces

At the start the agent says in one line whether verify runs and at which level. The level combines settings that already exist: the project's `UI verification` and `Review` values, your answers to the delivery questions, and whether the change is mechanical. Verify adds no setting and no question of its own. The levels, and what each check covers, are in the [skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md).

- Affected tests and evidence for the changed behavior, always when verify runs.
- A full local suite for a change that is not mechanical.
- A local run of the changed behavior when it calls for one.
- For a visible interface change, independent reviewers and a functional walk, as the `UI verification` value selects, followed by the human validation of the [walkthrough](/flows/build/#what-happens-to-a-user-interface-change).
- The dedicated code review, at the point your delivery mode sets.
- The [security review](/flows/security/) only when you ask for it.

When it finishes, the report names the level that ran and each check that failed, could not run or was accepted without verification. If the delivery mode lets it, the accepted candidate is integrated, and [Close](/flows/close/) follows.

## Raising, lowering or skipping the level

You can raise or lower the level for the work with a request such as "full verify", "no UI review" or "skip verify". Verify is skipped on its own only for a mechanical change, and the report then names the checks that already cover it. When you ask to skip a check, it counts as accepted without verification: the agent names each one at the start and in the report, and does not treat it as a pending gate. That acceptance never waives CI, branch protection or a review the project requires. Unattended delegation cannot skip verify or lower its level.

## What it does not do

- It does not write the change; that is [Build](/flows/build/).
- It does not mark a check as passed when it could not run it.
- It does not run a security review you did not ask for. It offers one when the change crosses a security boundary.
- It does not push or merge beyond the delivery you authorized, and it does not deploy; that is [Deploy](/flows/deploy/).

## Related flows

- [Build](/flows/build/) comes first and hands over the candidate.
- [Security](/flows/security/) runs inside Verify, or on a merged commit, on request.
- [Close](/flows/close/) follows once the candidate is integrated.
- [Deploy](/flows/deploy/) is a separate step after integration.

## FAQ

### Do I need a plan first?
No. Verify needs a candidate and the criteria it must meet. When it is entered directly, it asks only for the delivery and review values that are still missing.

### Is the code review part of Verify?
Yes. A declared `Review` value is the default mechanism, not a demand: if you skip verify, the dedicated review counts as accepted without verification, unless the repository requires it.

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-verify/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md)
