# Trackers

Read before the first write to a tracker, at [create-issue](../SKILL.md) step 8, and when a ticket is marked as blocked at its step 6. Each tracker's commands are in the references of [trackers](../../hive-core/tracker/trackers.md): [Jira](../../hive-core/tracker/trackers/jira.md), [Linear](../../hive-core/tracker/trackers/linear.md) and [GitHub](../../hive-core/tracker/trackers/github.md).

## Before writing

Read what the authenticated account may do in the target (create a ticket, set its type and labels, relate it, comment, edit a ticket someone else wrote); a tracker can drop a field the account may not set without an error, so name each field you cannot set before the write. New tickets go in the primary tracker: the `Tracker` that the `## Hive` of the repository the session works in declares, as [project settings](../../hive-core/plan/project-settings.md) sets, unless the user names another tracker or repository for the ticket.

## Type

Set the ticket's type with the tracker's own type when it has one that matches, such as Jira's story, task and bug, or an issue type the GitHub organization defines. Without a matching type, use a label the project already has for that type; without that label either, the type stands only in the ticket's text. A spike without a type of its own is filed as a task. Never create a type or a label for this.

## Parent and children

One ticket is an issue (GitHub, with an issue type where the organization defines types; a repository owned by a user account has none, so create the parent without a type), an issue (Linear) or a story, task or bug (Jira). A parent with children is a parent issue with sub-issues in GitHub and Linear (Linear has no epic type; add tickets to a project only when the user names an existing one) and an epic in Jira, with each child naming it as its parent (Jira sub-tasks are parts of one ticket, not children of a breakdown). Never create a project, milestone or initiative to hold a set. For another tracker, use its own parent and child relation; otherwise list the children in the parent's text and the parent in each child, and say in the report what could not be represented.

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
