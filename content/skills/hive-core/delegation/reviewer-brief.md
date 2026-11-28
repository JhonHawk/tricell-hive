# Reviewer brief

Read before briefing a reviewing or verifying child. It also applies when a reviewer's finding is reread after a correction.

Give a reviewer what it needs to judge the work and nothing that argues for it:

- the artifact: the diff, the plan files or the captures; a code review's diff leaves out the files `git diff -M --name-status` reports as `R100`, named in one line; a native review mechanism that accepts paths gets every changed path except those `R100` files, so added files and files renamed with changes are included;
- the requirements it must meet: the request, the ticket's criteria and the acceptance criteria, the diff base's `## Skills` lines [child brief](child-brief.md) copies, or the current ones without a diff, and for a UX review the image of the approved example when the work has one;
- the output of the checks already run, as their commands and results;
- for a reviewer that signs in, the account rows and saved browser state paths of [child brief](child-brief.md), never a value.

Never include the producer's reasoning, its account of why the work is right, a claim such as "verified" or "tests pass" without its output, or a list of facts or conclusions the reviewer should not re-check, such as "already settled" or "facts already gathered", since a framing that the work is correct lowers what a reviewer finds.

A UX reviewer's brief names the regions the change alters and the layouts and shared components that contain or reuse them, and, for each screen it captures, how to reach each state its role checks (loading, empty, error, the longest real values, such as the fixture route, the account and how to force the error) or which of them cannot be reached. It does the same for the states the measured checks of [criteria](../../create-design/references/criteria.md) need: for a list, search or filters, no results apart from an empty catalog, a list of more than 1000 items and returning from a detail view; for a form or any submit, how to hold the submit pending, such as a delayed stub or a throttled network. Frames drawn in a design file have no states to reach: their brief is the one [frame review brief](../../create-design/references/frame-review-brief.md) sets, and the brief for a design another person drew is the one [review](../../create-design/references/review.md) sets.

Reread a revised plan, or a correction of a UX reviewer's findings, with one follow-up to that same child; when its ID is gone, the host has no tool to message it, or the message returns an error, brief a new child of its role with the earlier findings. Of several plan reviewers, those [plan review](../../flow-plan/references/plan-review.md) names reread. A UX reread names the corrected regions and states, which it rereads as its role's step 3 sets, with the means to reach them; never run a new full UX review.
