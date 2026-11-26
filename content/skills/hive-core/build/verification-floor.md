# Verification floor

Read when a plan picks each task's checks, at a build's start and at verify's entry.

What always runs for each kind of change; each check keeps its own rule, linked here. The floor holds in every delivery mode, `hold` included: a mode decides only what happens after the checks.

| Kind of change | Always runs |
| --- | --- |
| Mechanical, as [proportionality](../plan/proportionality.md) defines it | The checks that already cover it; no verify and no code review |
| Every other change | The full suite and local CI, as [checks](checks.md) sets; [blast radius](../review/blast-radius.md); the code review; and the rows below that apply |
| Defect fix | Its regression test seen failing, then passing, as [tests](tests.md) and [defect fix](defect-fix.md) set; without a test runner, the direct check [no test runner](tests-no-runner.md) sets |
| New or changed behavior | The check [tests](tests.md) sets for each behavior, and the walk of each criterion in its place, as [flow-verify](../../flow-verify/SKILL.md) step 1 sets |
| Restructuring that keeps behavior | The existing tests, or a characterization test first, as [tests](tests.md) sets |
| A screen added or changed | [Before captures](../ui/before-captures.md), the captures [UI verification](../ui/ui-verification.md) sets, the `hive-review-ux` child with its [accessibility audit](../ui/a11y-audit.md) and the [walkthrough](../ui/walkthrough.md) |
| Risky, as the global guidance's «Proportionality» defines it | A `hive-verify-task` child per risky task of a plan with a plan file, as [flow-build](../../flow-build/SKILL.md) step 3 sets; for a change that touches a [security boundary](security-boundaries.md), the security review the user accepted |
| A measured speed or size goal | Before and after measurements read as [explain the number](../principles/explain-the-number.md) sets, against the threshold of the [perf playbook](../../hive-mode/playbooks/perf.md) |

## What the report carries

- **Try it:** when the user can see or operate the change as its users do, a block headed in the session language, such as **Try it**, written as the steps those users take rather than the commands that run the app:
  - **Start:** the URL, as a link, when the candidate is still served; otherwise one line naming the start command, and the seed command when the walk needs seeded data, that its `README.md` or a `HIVE.local.md` documents, with the URL it serves. Setup that documentation does not ask its reader to type, such as environment files, variables or flags, stays out. For a change used from a command line, such as a CLI, the command its users run is the start and its steps.
  - **Account:** its row in the account index (never a secret).
  - **Steps:** one block per flow, headed by the ticket, if any, and a short description, such as `ABC-123 - <description>`, with bullet steps from a named starting point: preconditions such as the role and signed-in state in parentheses on the first step, navigation with `>` between screens by the labels the interface shows, in its language, such as `Claims > New`, the action, and what the user should see. Condense obvious steps. When several tickets share one flow, cite the ticket in the step it concerns; separate flows get separate blocks.

  A command that reruns a check, such as a test run, is not this block, since the report already names the checks that passed; a change with nothing else to try has no block. It asks nothing and appears in every delivery mode, where [close report](../close/close-report.md) places it in the final report, and in a request to validate such a change.
- A check of this floor that did not run is named with its reason among the report's omitted checks.
