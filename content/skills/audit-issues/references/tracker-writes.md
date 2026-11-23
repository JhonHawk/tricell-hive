# Tracker writes

Read at [audit-issues](../SKILL.md) step 5, before reading the child auto-close and before any write of step 7. It sets how each write of an approved set is made in GitHub Issues, Linear and Jira. Each tracker's tool order and commands are in [trackers](../../hive-core/tracker/trackers.md) and its references: [GitHub](../../hive-core/tracker/trackers/github.md), [Linear](../../hive-core/tracker/trackers/linear.md) and [Jira](../../hive-core/tracker/trackers/jira.md). Read the tool's help or description before its first use in a session, since flags and fields change between releases.

## Before writing

Read what the authenticated account may do in the tracker, as create-issue's [trackers](../../create-issue/references/trackers.md) «Before writing» sets, and name in the set each write it cannot make. Every write below is one ticket at a time, and every ticket is read back after its write, as «Read back» sets.

## New parent

Create it as create-issue's [trackers](../../create-issue/references/trackers.md) «Type» and «Parent and children» set: a parent issue in GitHub and Linear, an epic in Jira. Never create a project, milestone, label, type or state to hold a group.

## Move under a parent

The representation is the one create-issue's [trackers](../../create-issue/references/trackers.md) «Parent and children» sets.

- **GitHub:** `gh issue edit <n> --parent <parent>`. A parent holds at most 100 sub-issues; a group past that is split.
- **Linear:** set the issue's parent (`parentId` in the Linear MCP tool). Read back its project and cycle too, since a move may change them.
- **Jira:** set the ticket's parent field to the epic through the tool that edits it, after reading its help.

## Close

Write the evidence comment first, as «Evidence comment» sets, then close. Never delete a ticket.

| Close | GitHub | Linear | Jira |
| --- | --- | --- | --- |
| As a duplicate | `gh issue close <n> --duplicate-of <canonical>` | Set the duplicate-of field to the canonical issue (`duplicateOf` in the Linear MCP tool); read the state back, and when it did not change, move it to the team's state whose `type` is `canceled` | The link type for duplicates the project has, from its link types, to the canonical ticket; then the transition to the state the project uses for cancelled work |
| As not planned | `gh issue close <n> --reason "not planned"` | Move it to the team's state whose `type` is `canceled` | The transition to the state the project uses for cancelled or rejected work |
| As completed (already fixed) | `gh issue close <n> --reason completed` | Move it to the team's state whose `type` is `completed` | The transition to the state the project uses for integrated work |

Every move to a state, in any tracker, follows [trackers](../../hive-core/tracker/trackers.md) «Move»: the states cache, the ticket's transitions read live, and the exact state name. In Jira, read the project's link types and the ticket's transitions before the first close, since both differ by project; when neither the duplicate link type nor a cancelled state exists, the ticket gets the comment and no close.

## Jira reopen

Propose a close in Jira only when you can read that the ticket's workflow has a transition back from the closing state, from the workflow or from a closed ticket of the same type and project. A transitions read lists only those of the ticket's current state, and reading the workflow may need administration permission: when you cannot read it, the ticket gets the comment and no close.

## Related and blocked by

| Relation | GitHub | Linear | Jira |
| --- | --- | --- | --- |
| Related | No native relation: a comment on each ticket with the other's link | The issue's related relation (`relatedTo` in the Linear MCP tool) | The project's «Relates» link type |
| Blocked by | `gh issue edit <n> --add-blocked-by <m>` | The issue's blocked-by relation | The project's «Blocks» link type |

A relation to a ticket in another tracker is a comment on each ticket with the other's URL, as create-issue's [trackers](../../create-issue/references/trackers.md) «Relating a ticket to its origin» sets.

## Evidence comment

Every close, every comment for another author and every `covered by` without a ticket gets one comment, written as [trackers](../../hive-core/tracker/trackers.md) «Comment and attach» sets, in the ticket's language: the verdict, its evidence (file and line, commit, pull request, test, decision or canonical ticket, each linked), and, for another author's ticket, that the audit leaves the decision to them. Paste no secret value and no account identifier.

## Child auto-close

Before the approval, read whether closing a parent closes its open children:

- **Linear:** a team setting closes open sub-issues when the parent closes. The Linear MCP tool's team read does not return it: report `unknown` and name the team's issue settings in Linear as the place to check it, unless another tool shows it.
- **GitHub and Jira:** neither documents it as a default; Jira automation rules can do it. Report `unknown` unless the tracker shows a rule or setting that does, and `off` when you can read that none exists.

A parent whose children would close with it is named in the set with that state, so the user approves knowing it.

## Read back

After each write, read the ticket back and compare its state, parent, relations, project and the comment with what you sent, since a tracker can drop a field without an error. Report a difference in the list of what was written instead of writing again. When a write fails, or its result cannot be read, stop as create-issue's [breakdown](../../create-issue/references/breakdown.md) «Writes» steps 4 and 5 set.
