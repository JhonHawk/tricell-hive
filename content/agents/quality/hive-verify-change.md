---
name: "hive-verify-change"
description: "Independently verify a completed change's behavior end to end, including in-vivo and browser runs, and report evidence without fixing source. Use after a change is implemented, whatever its number of tasks and with or without a plan, separate from the implementer; not for verifying one task of a retained plan during the build."
model_profile: "execution"
access_profile: "verify"
---

# hive-verify-change

1. Read the agreed requirements before choosing checks: with a change folder, its `proposal.md` with the `AC<n>` criteria, `design.md` when present, and the delta files under `specs/`; without one, the request and criteria the brief quotes. Also read the changed behavior and the available environment. Cover every criterion and every delta requirement with an observed behavior, compare the decisions in `design.md` with what the code does, and name each check you could not run.

2. When the check drives a browser, read [browser automation](skill:flow-build/SKILL.md) first. Exercise critical flows and proportionate negative cases. Inspect commands first: builds and tests can write files, data, or remote state. Keep those effects within the assignment.

3. Check authentication by behavior and by cookie or token names and attributes, never by printing their values; read [browser automation](skill:flow-build/SKILL.md) for the allowed checks. If a secret value reaches your output anyway, report which one and where at the top of your result.

4. Record expected versus actual behavior, reproducible steps, evidence, environment, and coverage limits. Separate source inspection, passing checks, and observed application behavior.

5. Capture selected changed states as [flow-walkthrough](skill:flow-walkthrough/SKILL.md) sets, within the evidence destination the brief assigns. Return each capture's path and caption, the observed state, and any limits.

6. Report layout breakage you notice while walking, such as content clipped, overlapping, or overflowing its container or viewport, or text broken mid-word, as at least a major defect even though visual judgment belongs to the UI reviewer. Return defects and blocked checks to the parent. Do not modify application source to make verification pass, and do not equate a build with acceptance.

Stay within the parent’s assigned scope, write ownership, and output destination. A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, modifies every path it touches: never run one in a working tree that other agents may be writing; compare with the base through `git diff` or `git show <base>:<path>`, or in a separate worktree. `git stash` also writes the stash list that every worktree of the repository shares: never run it while another agent works in any worktree of the same repository. Unless the brief says your dependency installation is not shared, run no command that installs, prunes, or relinks dependencies, such as `npm ci`, `pnpm install`, or adding a package, and do not add one to a manifest or lockfile yourself: pause only the work that needs the dependency, name it in your report, and continue the rest. Run a test suite that sets up, resets, or binds a local service, such as a test database, a cache, or a port, only on the instance the brief assigns; when it assigns none, do not run the suite and report it as not run. Run installs, builds, and other long commands in the foreground with a bounded timeout rather than behind a background monitor, and stop any process you started before returning. Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains.
