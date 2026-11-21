# Playbook: infra-ci

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work changes infrastructure, a CI or CD pipeline, build or deployment configuration, or developer tooling.

**Stages.** `hive-mode`, then [flow-plan](../../flow-plan/SKILL.md), [flow-build](../../flow-build/SKILL.md), [flow-verify](../../flow-verify/SKILL.md), [flow-deploy](../../flow-deploy/SKILL.md) when the change reaches an environment, and [flow-close](../../flow-close/SKILL.md). A mechanical change, such as configuration that changes no permission, security setting or dependency, skips the plan and the dedicated review ([flow-build](../../flow-build/SKILL.md) sets what it still checks).

**Specific to infrastructure and pipelines**
- Measure each fact about the outside system before planning, with read-only reads ([external facts](../../hive-rule-external-facts/SKILL.md)), and plan each effect as [effects round](../../hive-rule-external-effects/SKILL.md) sets: spending real money or paid quota, destroying or replacing infrastructure and changing shared or remote resources each need an answer before they run.
- Reproduce the repository's CI locally before the push and read each exit status ([checks](../../hive-rule-checks/SKILL.md)). A pipeline change shows its real effect only when the path it changes runs, so name that run as the place a criterion is observed.
- A change that touches permissions, process execution, secrets or what a job can read goes through [security boundaries](../../hive-rule-security-boundaries/SKILL.md). Secrets stay in the project's untracked secret location.
- A new package or tool from a registry passes the package check of the global guidance's «Scope and authorization» before it is added.
- A check that is weakened, skipped or excluded to make a pipeline green needs its justification: [weakened checks](../../hive-rule-suppression-review/SKILL.md).
- Environment branches change only through the declared path ([promotion path](../../hive-rule-promotion-path/SKILL.md)); each stop, push, pull request and merge happens under the delivery mode in force, and promotion follows `flow-deploy` with its authorization per environment.

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-rule-close-report/SKILL.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-rule-report-endings/SKILL.md) that [close modes](../../hive-rule-close-modes/SKILL.md) sets; a promotion's cleanup close uses the deployment row of [close report](../../hive-rule-close-report/SKILL.md#variants).

**Handoffs.** A defect the work exposes is a separate topic or [bugfix](bugfix.md); a project verification command or journey the work adds follows [project verification](../../hive-rule-project-verification/SKILL.md). Either is a scope change, as [scope changes](../../hive-rule-scope-changes/SKILL.md) sets.

**Leaves.** [verification floor](../../hive-rule-verification-floor/SKILL.md), [external facts](../../hive-rule-external-facts/SKILL.md), [effects round](../../hive-rule-external-effects/SKILL.md), [checks](../../hive-rule-checks/SKILL.md), [security boundaries](../../hive-rule-security-boundaries/SKILL.md), [suppression review](../../hive-rule-suppression-review/SKILL.md), [promotion path](../../hive-rule-promotion-path/SKILL.md).
