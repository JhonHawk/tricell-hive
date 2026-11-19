---
name: hive-rule-close-report
description: "Hive rule leaf: read when you write a final report or classify a finding. Sets the report's parts, the finding priorities P0 to P3 and the candidate records."
disable-model-invocation: true
---

# Final report, findings and candidates

Read when you write the final report of a close or classify a finding.

## Final report

Report the whole work item in one message, each part only when it has content:

- **Integrated:** what changed and where: the pull request, the commits, the ticket's state; for a pull request left open, its head commit, the base commit its current-base check passed on, and the CI result.
- **Resolved:** each acceptance criterion with how it was shown, the corrections the user asked for, and the defect's proof and the try-it block the [verification floor](../hive-rule-verification-floor/SKILL.md) sets.
- **Omitted:** each check that was skipped or could not run, with its reason, including the CI-only jobs.
- **Repaired:** the local environment repairs made, such as reinstalled dependencies or seeded data.
- **Support material:** each item with its path and the action applied, or listed as pending, as [support material](../flow-close/references/support-material.md) sets.
- **Findings:** the defects not fixed, as [Findings](#findings) sets.
- **Candidates:** the work's repetitions, as [Candidates](#candidates) sets.
- **Next:** what remains, with who owns it, and what comes next, such as the merge left to the user and review bots or base CI still pending.

When the user asks to keep or share this report, or it exceeds what fits on one screen, offer a `create-report` page as one reply of the close footer.

## Findings

A finding is a defect the work found but did not fix: a review finding below P1, a UI defect that was already there, a test that already failed or fails intermittently (keep its error output), or anything else noticed on the way. Every finding carries one priority by its observable impact, whatever scale its source used: P0 blocks the release, an operation or a core flow for every input; P1 is blocking, as [review findings](../hive-rule-review-findings/SKILL.md) defines it, and must be fixed before merge; P2 should be fixed but does not block; P3 is minor. A finding of P2 or higher with a concrete failure scenario is proposed; the rest get one line each in the report. P2 findings that share an area or a cause are proposed as one group that lists each member with its evidence; findings of P1 or higher are proposed one by one. Never open a ticket without the user's answer, which unattended delegation gives at its step 5.

## Candidates

A candidate is a repetition the work did not turn into a mechanism under [defect fix](../hive-rule-defect-fix/SKILL.md) point 5: an inline script or command chain the work ran two or more times for the same purpose, such as the same link check or hash comparison, or a correction the user gave twice in the work. The report lists each one on one line, with what repeated, how many times and the mechanism that point would choose, and asks nothing about it; a work with none has no such part. [Script lifecycle](../project-layout/references/script-lifecycle.md) sets where a script goes at each level.

At step 7, search the project's persistent memory for each candidate by the terms of its purpose, then create or update one record under the topic `candidate/<tool>-<target>` in kebab case, such as `candidate/link-check-docs`, with the sessions that saw it, their dates and its count. A candidate seen in two sessions or more is offered in the close footer; it is not a finding, since it is no defect.
