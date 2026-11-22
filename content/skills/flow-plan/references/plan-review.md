# Plan review across domains

Read from [flow-plan](../SKILL.md) step 5 when the plan's tasks change files of two or more domains below. A data-entry screen also counts as client, so a client plan with such a screen counts two. A plan in one domain gets one `hive-review-plan` child, and this reference does not apply.

## Domains

A domain counts when a task's `Locations:` change files of it. Tests and documentation count with the domain they check or describe. Repository-wide files, such as manifests, lockfiles, build configuration, the changelog and the change folder, count with the other files of the task that changes them, or with server when that task changes no other file, and any other file no row names counts with server. A task that changes files of two domains counts in both.

| Domain | Files | What its reviewer examines |
| --- | --- | --- |
| server | server code, APIs, jobs, schemas, migrations and queries | API and event contracts and their consumers, data ownership, authorization, migrations and compatibility, failure behavior, and the tests that show each |
| client | web or mobile interface code and its styles | the screen's place in the app and the patterns it reuses, its loading, empty and error states, accessibility, and whether it matches the server contracts it calls |
| data-entry screen | a client screen where the user enters or chooses data | the user's real task through each planned view: values the system already holds or can derive, entries asked again per record, steps per repeated item, and recovery from an error |
| guidance | files under `content/` and other agent instructions, and documentation that states their rules | every sentence, with file and line, that would still contradict a rule the plan changes, and any option or task that contradicts a standing rule the plan does not change |
| infrastructure | CI, deployment, containers and infrastructure as code | environments, identities, secrets handling, deployment order, recovery, and what each gate authorizes |

## How many reviewers

Launch one `hive-review-plan` child per counted domain, together, before step 4's question, at most three. With four or more domains, merge the assignment with the fewest tasks into the one with the next fewest, a merged assignment counting the sum of its tasks and taking its earliest row's place, and repeat until at most three remain; on a tie, take the later one in the table. Never merge the data-entry screen assignment.

## Each brief

Brief each child as [reviewer brief](../../hive-core/delegation/reviewer-brief.md) sets, and add its domain, the questions of its row and the contracts it shares with the other counted domains, such as a payload, a schema or a rule another domain's tasks change. Tell each child to review its domain's tasks and those shared contracts. Give the role's whole-plan checks, the criteria, the external facts and the verification lines, to the first assignment in the table's order only, and tell the other children to skip them, so they are not run three times.

## Reconcile

Merge a finding that several reviewers report into one, naming each reviewer. Then apply what holds, as flow-plan step 5 sets. In the progress section, record each reviewer's duration and tokens when the host reports them.

## Reread a revision

A reviewer rereads a revision, as [reviewer brief](../../hive-core/delegation/reviewer-brief.md) sets, when the revision changed the text, `Closes:` or verification of a task in its domain, or a criterion such a task closes. The first assignment also rereads a revision that changes a criterion, an external fact or a verification line. When a change belongs to no single domain, every reviewer rereads it.
