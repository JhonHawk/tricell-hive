# Project settings

## When to read

Read before resolving a project setting needed by the task, including the governing repository, base branch, tracker, specs location, environments or completion point. Reading this reference does not invoke the planning workflow.

## Declare settings

Each repository's `AGENTS.md` holds a `## Hive` section with its own settings. The section states verifiable facts about the repository, not decisions: the delivery mode, the code-review mechanism and the UI verification level are asked in the round [delivery decisions](../../flow-plan/references/delivery-decisions.md) sets and reused for the session, never declared. Required: `Project`, `Base branch`, `Tracker`, and `Specs`, the path to the project's `openspec` directory, followed by ` · <git-url>` when it lives in another repository. Optional: `Environments`, the ordered environments from the base branch to production, such as `development → qa → production`: the last one is production and every earlier one is non-production, and `local` is never one of them even when listed; `Delivery: direct-base`, which marks a support repository such as specs or mocks and is the only value of the section that authorizes effects, as [delivery decisions](../../flow-plan/references/delivery-decisions.md) sets; `Completion`, the project's later completion point, as `<intermediate state> on <ref> · <final state> on <ref>`, such as `To Deploy on origin/development · Done on origin/qa`, applied as the next section sets; and `Hive guidance: required` (every session there, CI included, runs with this guidance). A workspace `AGENTS.md` may hold one for project-wide values and cross-repository work. Sections do not inherit across repositories, because a host started in a repository may not load the level above.

Retired keys: `Review`, `UI verification`, and `Delivery` with `automatic` or `interactive`. A section that still holds one authorizes nothing and answers no question: ask as if it were absent, and propose removing it, as [harness-audit](../../harness-audit/SKILL.md) does. A standing grant written in prose elsewhere in `AGENTS.md`, such as an administrative bypass or a pre-authorized deployment, has no key; it never stands in for a setting.

Example for a code repository:

```markdown
## Hive
- Project: ark (workspace `..`)
- Base branch: development
- Environments: development → qa → production
- Tracker: Linear · team sample · project ark · key ARK
- Specs: ../ark-specs/openspec · https://github.com/<org>/ark-specs.git
- Completion: To Deploy on origin/development · Done on origin/qa
```

## Ticket moves

The global guidance authorizes moving a tracked ticket to its in-progress state before implementing it and to done at integration. Whenever you change a ticket's state, change each of its sub-issues that the same work covers to the same state in the same step, and name in the report any sub-issue you left and why. The base branch is the one work branches are cut from and merged back into, as the `Base branch` setting declares; with environment branches it is the first of them, such as `development`, and later ones receive changes only by promotion.

## Later completion point

**Condition:** the repository's section declares `Completion`.

**Actions:** the work item still closes when it is integrated into the base branch, and [flow-close](../../flow-close/SKILL.md) closes it then. At integration, move the ticket to the intermediate state `Completion` names, such as To Deploy, not to its final state. Move it to the final state when its ref holds the work: [flow-deploy](../../flow-deploy/SKILL.md) applies that move when it runs the promotion, and when someone else promoted, the next session that observes the condition applies it. The global ticket rule authorizes both tracker writes. A promotion never holds the work item open.

## Resolve the governing section

**Condition:** the task changes a file or spans repositories.

**Actions:** the section of the Git repository containing the changed file governs; a package inside a repository uses the repository's section. The workspace section governs only cross-repository work and sessions started there without a repository target. When loaded sections conflict, follow the repository's and report the conflict.

## Resolve a missing value

**Condition:** a required value is missing and the task first needs it.

**Actions:** ask for it then, not at session start and not for repositories only read, then record it in that repository's section. A value found only in other documentation is a recommendation for that question, not the setting. Keep each value there rather than repeating it elsewhere in the file. Use the [question procedure](questions.md). To create, complete or update the whole section, or move a value written in prose into it, use [harness-audit](../../harness-audit/SKILL.md).

For relative `Specs` paths, missing checkouts, shared specs, and change-folder formats, read [change records](../../flow-plan/references/change-records.md).
