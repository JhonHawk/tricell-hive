---
name: hive-rule-verification-floor
description: "Hive rule leaf: read when a plan picks each task's checks, at a build's start and at verify's entry. What always runs for each kind of change, the defect's output and the try-it block."
disable-model-invocation: true
---

# Verification floor

What always runs for each kind of change. Each check keeps its own rule, linked here. The floor holds in every delivery mode, `hold` included: a mode decides only what happens after the checks.

| Kind of change | Always runs |
| --- | --- |
| Mechanical, as «Proportionality» of the global guidance defines it | The checks that already cover it; no verify and no code review |
| Defect fix | Its regression test seen failing, then passing, as [tests](../hive-rule-tests/SKILL.md) and [defect fix](../hive-rule-defect-fix/SKILL.md) set; the full suite and local CI, as [checks](../hive-rule-checks/SKILL.md) sets; [blast radius](../hive-rule-blast-radius/SKILL.md); the code review |
| New behavior without a screen | The owner test of each behavior, as [tests](../hive-rule-tests/SKILL.md) sets; the full suite and local CI; the walk of each criterion in its place, as [flow-verify](../flow-verify/SKILL.md) step 1 sets; [blast radius](../hive-rule-blast-radius/SKILL.md); the code review |
| A screen added or changed | Its kind's row, plus [before captures](../hive-rule-before-captures/SKILL.md), the captures [UI verification](../hive-rule-ui-verification/SKILL.md) sets, the `hive-review-ux` child and the [walkthrough](../hive-rule-walkthrough/SKILL.md) |
| Risky, as «Proportionality» defines it | Its kind's row, plus a `hive-verify-task` child per risky task of a plan with a change folder, as [flow-build](../flow-build/SKILL.md) step 3 sets; for a change that touches a [security boundary](../hive-rule-security-boundaries/SKILL.md), the recommended `hive-review-security` |
| A measured speed or size goal | Before and after measurements read as [explain the number](../hive-principle-explain-the-number/SKILL.md) sets, against the threshold of the [perf playbook](../hive-mode/playbooks/perf.md) |

## In the final report

- **A defect's proof:** the regression test's output seen failing on the code before the fix, then passing after it, trimmed to the lines that show both. When the failure could not be reproduced before the fix, say so instead.
- **Try it:** when the change can be seen or run, end the final report, and any request for validation, with a block headed in the session language, such as **Try it**: the URL or command to open it, the test account by its row in the account index (never a secret), the steps, and the result to expect. It asks nothing and appears in every delivery mode.
- A check of this floor that did not run is named with its reason among the report's omitted checks.
