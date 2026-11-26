# Project structure

Read from [breakdown](breakdown.md) when the user asks for phases or the set has more than 7 children. It sets what groups the children of a set in the tracker; the set, its approval and its writes stay as [breakdown](breakdown.md) sets them.

## Phases

Propose the phases as groups of children delivered together, in delivery order, each with a name and the outcome it reaches; a child belongs to one phase. A phase is the tracker's own grouping for it:

| Tracker | A phase is | Created through |
| --- | --- | --- |
| Linear | a milestone of an existing project: the one the user names, or the one the location of the repository's `Tracker` value names | the Linear tool's milestone write |
| GitHub | a milestone of the repository | `gh api` on the REST route `POST /repos/{owner}/{repo}/milestones`, since `gh` has no milestone command; `gh issue create --milestone` assigns it |
| Jira | a version of the project, set as each ticket's fix version | a tool of the session that creates versions; `acli` has none |

Read the tool's current help or description before the first write, as [trackers](trackers.md) sets. When no tool of the session can create the grouping, or the tracker has none, create nothing: name the phase in each child's text and say in the report what could not be represented. For another tracker, use its own grouping for phases under the same rule.

## Parent

In Linear, when the phases are milestones of a project, the project takes the parent's place: create no parent ticket, link each child to the project and to its milestone, and give the set's shared outcome and what is out of scope in the report, for the user to place in the project, since the breakdown never edits a project. Without a project that the user or the setting names, a Linear set keeps its parent and names each phase in its children's text. In GitHub and Jira the set keeps its parent, as [trackers](trackers.md) «Parent and children» sets, and each child also carries its milestone or version.

## Labels

Propose a label only for a module or area that two or more children of the set share and the project has no label for; reuse an existing label whose name means the same.

## What is created, and when

Create a milestone or a label only when the approved set shows it: the approval of [breakdown](breakdown.md) lists each one by name before any write, and its writes create them before the tickets, one at a time, reading each back. Never create a project, an initiative, a team, a state or a ticket type. To undo a milestone or a label, name the ones created and leave their removal to the user; never delete one that a ticket outside the set uses. While [unattended delegation](../../unattended-delegation/SKILL.md) is active no set is created, so nothing here is created either.
