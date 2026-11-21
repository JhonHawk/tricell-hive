# Trackers

Read before the first write to a tracker, at [create-issue](../SKILL.md) step 8, and when a ticket is marked as blocked at its step 6. Each tracker's commands are in the references of [trackers](../../hive-core/references/trackers.md): [Jira](../../hive-core/references/trackers/jira.md), [Linear](../../hive-core/references/trackers/linear.md) and [GitHub](../../hive-core/references/trackers/github.md). It names each tracker's own concept; perform the action through the tracker's own CLI or tool, and read that tool's current help for the exact command, since its options change between versions.

## Before writing

Read what the authenticated account may do in the target: create a ticket, set its type and labels, relate it to another, comment, and edit a ticket someone else wrote. A tracker can drop a field the account may not set without reporting an error, so name before the write each field you will not be able to set. New tickets go in the primary tracker: the `Tracker` that the `## Hive` of the repository the session works in declares, as [project settings](../../hive-core/references/project-settings.md) sets, unless the user names another tracker or repository for the ticket.

## Type

Set the ticket's type with the tracker's own type when it has one that matches, such as Jira's story, task and bug, or an issue type the GitHub organization defines. Without a matching type, use a label the project already has for that type; without that label either, the type stands only in the ticket's text. A spike without a type of its own is filed as a task. Never create a type or a label for this.

## Parent and children

| Level | GitHub | Linear | Jira |
| --- | --- | --- | --- |
| One ticket | An issue, with an issue type where the organization defines types | An issue | A story, task or bug |
| Parent with children | A parent issue with sub-issues | A parent issue with sub-issues | An epic, with each child naming it as its parent |

- In GitHub, issue types exist only where the organization defines them. A repository owned by a user account has none: create the parent without a type.
- Linear has no epic type: the parent issue is the epic. Add the tickets to a project only when the user names an existing one.
- In Jira, the epic is an issue type. Sub-tasks are parts of one ticket and are not used for the children of a breakdown.
- Never create a project, a milestone or an initiative to hold a set.
- For another tracker, use its own parent and child relation when it has one. Otherwise list the children in the parent's text and the parent in each child, and say in the report what could not be represented.

## Relating a ticket to its origin

Relate the new ticket to the ticket it came from: blocked-by or blocking when one waits on the other, a sub-issue when it is a deferred part of the origin ticket's scope, related otherwise. When the origin ticket is in another tracker, link each ticket to the other by its URL in a comment. A ticket with no origin ticket gets no relation.

## Blocked

A ticket created as blocked, because a business rule is undecided or it lacks a fact nobody in the session can supply, is marked with the first of these that exists:

1. The tracker's native blocked-by relation, when what blocks it is another ticket.
2. The label or state the project already uses for blocked work or for work that needs information; read the project's labels and states to find it.
3. Otherwise, the first line of the ticket: that it is blocked, the question, and who it waits on, in the ticket's language.

For this, never create a label or a state. In every case the question and who answers it also stand in the ticket's open questions.

## After writing

Read the ticket back, as the skill's step 8 sets. To undo a ticket, close it as not planned; never delete one.
