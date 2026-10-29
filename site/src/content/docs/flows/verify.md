---
title: Verify
description: Check a built candidate at a level you can see and change, then integrate it as your delivery answer allows.
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

The level combines answers that already exist: your answers to the delivery, review and `UI verification` questions, and whether the change is mechanical. Verify adds no setting and no question of its own. The levels, and what each check covers, are in the [skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md).

- Affected tests and evidence for the changed behavior, always when verify runs.
- The full local suite of each affected application, run before a validation stop, or once on the final candidate when no CI gate runs it. A mechanical change does not need it.
- A local run of the changed behavior when it calls for one. When it does not, the report says why in one line; there is no question about it.
- For a visible interface change, independent reviewers and a functional walk, as your `UI verification` answer selects, followed, in a delivery mode that stops for your validation, by the human validation of the [walkthrough](/flows/build/#what-happens-to-a-user-interface-change).
- The dedicated code review at the point your delivery mode sets, as [review and gates](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/references/review-and-gates.md#dedicated-review-timing) sets. A local review runs after the full suite and before validation under `interactive`, before the push under `automatic` and `direct-base`, or on the working tree under `hold`; corrections you ask for during validation get one more review, limited to what changed. A mechanism that works only on a pull request runs after the push instead, and the validation message says it is still due. Under `hold` or `direct-base`, which open no pull request, the agent asks for a local mechanism or reports the review as not run. A mechanical change gets none.
- The [security review](/flows/security/) only when you ask for it. It runs in parallel with the code review of the same candidate, and its findings on a candidate not yet integrated are handled by the same severity, refutation and re-review rules.

When it finishes, you see what changed, each check that failed, could not run or was accepted without verification, and the level only when it ran below the one agreed; checks that passed are not listed. When your validation or [Close](/flows/close/) comes next, that summary arrives inside the validation request or the completion report rather than in a message of its own, so the work does not stop in between; when Verify stops at a blocker, or at the authorized result with nothing left for Close, the summary comes on its own. If the delivery mode lets it, the accepted candidate is integrated, and Close follows. The rule is in the [Verify skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md#integrate-and-close).

When the delivery mode stops for your validation, the agent first opens every link or service it is about to give you, as you would, and fixes a broken local setup before asking. It asks as that validation: which mode stopped, what to check and where, and what follows once you accept, which under `interactive` is the push, pull request and merge your delivery answer already covers, and under `hold` nothing is published. It never words the stop as a request to publish. The target check is in [review and gates](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/references/review-and-gates.md#runtime-and-ui-gates) and the wording in [questions](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/questions.md).

## Corrections during the validation pause

Each correction you ask for is classified first and then run as its own round; a repeated correction on the same aspect can lead the agent to ask whether to go back to the plan. [Changes and loops](/flows/overview/) shows where each kind of request returns the work, and the [scope changes reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/scope-changes.md) holds the classification.

## Raising, lowering or skipping the level

You can raise or lower the level for the work with a request such as "full verify", "no UI review" or "skip verify". Verify is skipped on its own only for a mechanical change, and the report then names the checks that already cover it. When you ask to skip a check, it counts as accepted without verification: the agent names each one in the report, and does not treat it as a pending gate. That acceptance never waives CI, branch protection or a review the project requires. Unattended delegation cannot skip verify or lower its level.

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
Yes, except for a mechanical change. The review mechanism you chose is the default, not a demand: if you skip verify, the dedicated review counts as accepted without verification, unless the repository requires it.

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-verify/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md)
