---
name: "hive-verify-change"
description: "Independently verify a completed change's behavior end to end, including in-vivo and browser runs, and report evidence without fixing source. Use only when the user asks for an independent end-to-end verification of a built change, or when a deployment check walks the functional flows of a deployed environment; not for verifying one task of a retained plan during the build, and not as a routine step of verification."
model_profile: "execution"
access_profile: "verify"
---

# hive-verify-change

1. Read the agreed requirements before choosing checks: with a plan file, its `AC<n>` criteria and `## Design` when present; without one, the request and criteria the brief quotes. Cover every criterion with a behavior observed in the environment your brief assigns, never through a substitute as [Acceptance criteria](skill:flow-plan/SKILL.md#acceptance-criteria) defines it. For each criterion that names an `Observed in:` place, say whether that environment is it, since an observation elsewhere reports the behavior and does not close the criterion. Compare the decisions in the plan's `## Design` with what the code does, and name each check you could not run.

2. When the check drives a browser, use the tool your brief assigns, as [Browser tool](skill:hive-core/ui/ui-verification.md#browser-tool) for a web interface, [Device tool](skill:hive-core/ui/ui-verification/device-tool.md) for a native mobile app. Exercise each flow an acceptance criterion names, plus one negative case per criterion with a denied, invalid or empty input. Inspect commands first: builds and tests can write files, data or remote state; keep those effects within the assignment. Sign in only with the accounts your brief gives, as [test accounts](skill:hive-core/local/test-accounts.md) sets for a child.

3. Check authentication by behavior and by cookie or token names and attributes, never by printing their values. If a secret value reaches your output anyway, report which one and where at the top of your result.

4. Record expected versus actual behavior, reproducible steps, evidence, environment and coverage limits, separating source inspection, passing checks and observed application behavior.

5. Capture selected changed states as [UI verification](skill:hive-core/ui/ui-verification.md) sets, within the evidence destination the brief assigns, and return to the primary producer each capture's path and caption, the observed state and any limits.

6. Report layout breakage you notice (content clipped, overlapping or overflowing its container or viewport, text broken mid-word) as at least P1, though visual judgment belongs to the UI reviewer. Return defects and blocked checks to the parent; do not modify application source to make verification pass or equate a build with acceptance.

Stay within the parent’s assigned scope, write ownership, and output destination. Never run `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean` or `git switch` in a working tree other agents may be writing, and never `git stash` while another agent works in any worktree of the repository, since the stash list is shared. Unless the brief says your dependency installation is not shared, run no command that installs, prunes or relinks dependencies and add none to a manifest or lockfile; pause only the work that needs one, name it in your report, and continue. Run a test suite that sets up, resets or binds a local service only on the instance the brief assigns; with none assigned, do not run it and report it as not run. Run install, build and test commands in the foreground with a timeout of at most 10 minutes; report a timeout as not completed. Stop any process you started before returning. Return evidence, limitations and needed decisions in one final reply; before it, send at most one message to the parent, for a blocker that needs its decision, and make that the final reply when no independent work remains.
