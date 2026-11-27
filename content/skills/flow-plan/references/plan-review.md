# Plan review

Read from [flow-plan](../SKILL.md) step 5 when the plan's tasks change files in two or more repositories. The `hive-review-plan` role reads «Domains» itself.

## How many reviewers

One `hive-review-plan` child reviews the plan, whatever the number of domains it touches. When the tasks' `Locations:` change files in two or more repositories, launch one child per repository instead, together, before step 4's question, at most three. A task that changes files of two repositories counts in both. With four or more repositories, merge the one with the fewest tasks into the one with the next fewest, a merged assignment counting the sum of its tasks, and repeat until at most three remain; on a tie, merge the one the plan names later.

Several reviewers of one repository read the same files: in real sessions from 2026-11-21 to 2026-11-26, plans reviewed by two or three domain reviewers in one repository cost 2.4 times a plan with one reviewer, with fewer findings per reviewer.

## Domains

The reviewer adds the questions of each domain the tasks it reviews count. A domain counts when a task's `Locations:` change files of it. Tests and documentation count with the domain they check or describe. Repository-wide files, such as manifests, lockfiles, build configuration and the changelog, count with the other files of the task that changes them, or with server when that task changes no other file, and any other file no row names counts with server. A task that changes files of two domains counts in both. A data-entry screen also counts as client, so a client plan with such a screen counts two.

| Domain | Files | Questions the reviewer adds to the role's checks |
| --- | --- | --- |
| server | server code, APIs, jobs, schemas, migrations and queries | contracts and their consumers, data ownership, migrations and compatibility |
| client | web or mobile interface code and its styles | the screen's place in the app, its loading, empty and error states, and whether it matches the server contracts it calls |
| data-entry screen | a client screen where the user enters or chooses data | the user's real task through each planned view: values the system already holds or can derive, entries asked again per record, steps per repeated item, recovery from an error |
| guidance | files under `content/` and other agent instructions, and documentation that states their rules | every sentence, with file and line, that would still contradict a rule the plan changes, and any option or task that contradicts a standing rule the plan does not change |
| infrastructure | CI, deployment, containers and infrastructure as code | environments, identities, secrets handling, deployment order, recovery, and what each gate authorizes |

## Each brief

Brief each child as [reviewer brief](../../hive-core/delegation/reviewer-brief.md) sets, and add the repository whose tasks it reviews and the contracts that repository shares with the others, such as a payload, a schema or a rule another repository's tasks change; it reviews its repository's tasks and those shared contracts. Give the role's whole-plan checks (the criteria, the external facts and the verification lines) to one child only, the one whose assignment counts the most tasks, a merged assignment counting the sum of its tasks; on a tie, the one the plan names first. Tell the others to skip them.

## Reconcile

Merge a finding that several reviewers report into one, naming each reviewer, then apply what holds as flow-plan step 5 sets. Record each reviewer's duration and tokens in the progress section when the host reports them.

## Reread a revision

When a revision is reread, as [Revise a plan in progress](../../hive-core/plan/plan-record.md#revise-a-plan-in-progress) step 4 sets with its trigger and its limit, the reviewer rereads it, as [reviewer brief](../../hive-core/delegation/reviewer-brief.md) sets. With several reviewers, the reviewer of each repository whose tasks or criteria it changed rereads it; the one with the whole-plan checks also rereads a change to an external fact or a verification line, and every reviewer rereads a change that belongs to no single repository.
