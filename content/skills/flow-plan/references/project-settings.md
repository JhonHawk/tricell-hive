# Project settings

## When to read

Read before resolving a project setting needed by the task, including the governing repository, base branch, tracker, specs location, delivery or review. Reading this reference does not invoke the planning workflow.

## Declare settings

Each repository's `AGENTS.md` holds a `## Hive` section with its own settings. Required: `Project`, `Base branch`, `Tracker`, and `Specs`, the path to the project's `openspec` directory, followed by ` · <git-url>` when it lives in another repository. Optional: `Environments`; `Review`, the code-review mechanism for work there, or `none` for no dedicated review; `Delivery`, either `direct-base` for a support repository such as specs or mocks, or `automatic` or `interactive` as the delivery mode for a code repository, as [delivery decisions](delivery-decisions.md) sets; and `Hive guidance: required` (every session there, CI included, runs with this guidance). A workspace `AGENTS.md` may hold one for project-wide values and cross-repository work. Sections do not inherit across repositories, because a host started in a repository may not load the level above.

Example for a code repository:

```markdown
## Hive
- Project: ark (workspace `..`)
- Base branch: development
- Environments: development → qa → production
- Tracker: Linear · team sample · project ark · key ARK
- Specs: ../ark-specs/openspec · https://github.com/<org>/ark-specs.git
- Review: /code-review
- Delivery: automatic
```

## Resolve the governing section

**Condition:** the task changes a file or spans repositories.

**Actions:** the section of the Git repository containing the changed file governs; a package inside a repository uses the repository's section. The workspace section governs only cross-repository work and sessions started there without a repository target. When loaded sections conflict, follow the repository's and report the conflict.

## Resolve a missing value

**Condition:** a required value is missing and the task first needs it.

**Actions:** ask for it then, not at session start and not for repositories only read, then record it in that repository's section. A value found only in other documentation is a recommendation for that question, not the setting. Keep each value there rather than repeating it elsewhere in the file. Use the [question procedure](questions.md).

For relative `Specs` paths, missing checkouts, shared specs, and change-folder formats, read [change records](change-records.md).
