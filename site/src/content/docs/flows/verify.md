---
title: Verify
description: Check a built candidate at a level you can see and change, then integrate it as your delivery answer allows.
---

Verify is the step between a built change and a finished one. [Build](/flows/build/) hands over a candidate; Verify gathers the evidence that it meets what you asked for, and says what was not checked. The [Verify skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md) owns the steps and their order.

## When it starts

Verify starts when a candidate exists: commits, or the working tree when your delivery answer is to hold. You can also start it on a change that is already built. Without a candidate, it says what is missing and returns the work to Build.

```text wrap
Verify the change on this branch and integrate it if everything passes.
Skip verify for this one; I will check it myself.
```

## What it does

- Establishes passing evidence for the full local suite, reusing Build's results when they still cover the candidate and running missing or invalidated checks. It serves the candidate the way production would and exercises the changed behavior through its real path, checking each numbered acceptance criterion. A screen that shows a sign-in page, an error or an empty list counts as a setup failure, not as evidence. Before it signs in, it reuses an account from your project's account index when one fits, as the [test accounts rule](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md) sets.
- For an interface change, captures the changed screens and states on desktop and, when the layout is responsive, mobile, as the [UI verification reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/references/ui-verification.md) sets. An independent UX reviewer works from those captures and runs an automated accessibility audit, plus any UI linter your project declares, on each screen and state it reviews.
- Runs the code review you chose, alongside the walk. Every reviewer works in a separate helper that gets the change, its requirements and the check outputs, never the builder's account of why it is right; for an interface change, the UX reviewer also checks how the changed regions affect the layouts and components around them. Findings are checked against the candidate, with a local reproduction when available and authorized or cited source evidence for contradictions established by reading. An unconfirmed concern stays identified as such. A security review runs only if you ask for it.
- Fixes, in one round, your corrections, every interface defect the change introduced, and each confirmed blocking finding. A second review covers only that round's fixes. Other findings go to the close instead of another round.

## What it asks you

When your delivery answer stops for validation, Verify shows the result in the [viewer](/viewer/) while the reviews still run, saying that approval is subject to their findings, and asks you to approve or request corrections. After a round that changed something visible, it shows only those screens and asks again. A visible defect that was already there, in a region the change affects, comes to you as a finding with a recommendation. Decisions follow the [delivery and question rules](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).

## How it ends

Before the push, it ensures the locally runnable CI jobs have current passing evidence, repeating checks affected by changed or uncertain inputs and naming the rest as CI-only. The [Verify skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md#check-reuse) defines when evidence stays valid; the real-path walk and required independent verification remain. Then it integrates as your delivery says: push, pull request, a single wait for CI to end, one read of the reviews, and merge. Review bots count only for the latest commit, their findings are checked against the code and get one round of fixes, and a bot that has not reviewed yet does not hold the merge. With the `ready` delivery it stops before the merge instead, with the pull request open, checked against the current base, for you to merge. A failing CI job is diagnosed from its log and fixed at the cause. [Close](/flows/close/) follows.
