# Plan review across domains

Read from [flow-plan](../SKILL.md) step 5 when the plan's tasks change files of two or more domains below. A data-entry screen also counts as client, so a client plan with such a screen counts two. A plan in one domain gets one `hive-review-plan` child and needs no more.

## Domains

A domain counts when a task's `Locations:` change files of it. Tests and documentation count with the domain they check or describe. Repository-wide files, such as manifests, lockfiles, build configuration and the changelog, count with the other files of the task that changes them, or with server when that task changes no other file, and any other file no row names counts with server. A task that changes files of two domains counts in both.

| Domain | Files | What its reviewer examines beyond the role's checks |
| --- | --- | --- |
| server | server code, APIs, jobs, schemas, migrations and queries | contracts and their consumers, data ownership, migrations and compatibility |
| client | web or mobile interface code and its styles | the screen's place in the app, its loading, empty and error states, and whether it matches the server contracts it calls |
| data-entry screen | a client screen where the user enters or chooses data | the user's real task through each planned view: values the system already holds or can derive, entries asked again per record, steps per repeated item, recovery from an error |
| guidance | files under `content/` and other agent instructions, and documentation that states their rules | every sentence, with file and line, that would still contradict a rule the plan changes, and any option or task that contradicts a standing rule the plan does not change |
| infrastructure | CI, deployment, containers and infrastructure as code | environments, identities, secrets handling, deployment order, recovery, and what each gate authorizes |

## How many reviewers

Launch one `hive-review-plan` child per counted domain, together, before step 4's question, at most three. With four or more domains, merge the assignment with the fewest tasks into the one with the next fewest, a merged assignment counting the sum of its tasks and taking its earliest row's place, and repeat until at most three remain; on a tie, take the later one in the table. Never merge the data-entry screen assignment.

## Each brief

Brief each child as [reviewer brief](../../hive-core/delegation/reviewer-brief.md) sets, and add its domain, the questions of its row and the contracts it shares with the other counted domains, such as a payload, a schema or a rule another domain's tasks change; it reviews its domain's tasks and those shared contracts. Give the role's whole-plan checks (the criteria, the external facts and the verification lines) to the first assignment in the table's order only, and tell the others to skip them.

## Reconcile

Merge a finding that several reviewers report into one, naming each reviewer, then apply what holds as flow-plan step 5 sets. Record each reviewer's duration and tokens in the progress section when the host reports them.

## Reread a revision

When a revision is reread, as [Revise a plan in progress](../../hive-core/plan/plan-record.md#revise-a-plan-in-progress) step 4 sets with its trigger and its limit, the reviewer of each domain whose tasks or criteria it changed rereads it, as [reviewer brief](../../hive-core/delegation/reviewer-brief.md) sets; the first assignment also rereads a change to an external fact or a verification line, and every reviewer rereads a change that belongs to no single domain.
