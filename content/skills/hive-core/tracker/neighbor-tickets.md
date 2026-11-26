# Neighbor tickets

Read before assessing, planning or building a ticket whose parent, project, milestone or epic holds other open tickets, and before commenting on another ticket about work moved between them. This leaf is the only home of the rule. A ticket with no such container gets one line that says so, and none of the steps below.

## Query

1. **Name the capabilities.** List what the ticket's criteria need and its own outcome does not deliver, such as sending email, storing a file, a permission or a table another ticket creates, each as a term of two to four words.
2. **Run one query per declared tracker**, filtered by that parent, project, milestone or epic, for the tickets in a non-terminal state. It returns only short fields (identifier, title, state, labels, parent, milestone and blocking relations), is capped at 50 tickets and requests no descriptions, since a description costs context that matching by title and label does not need. Reuse a listing this session already holds, such as the status report's. When the filter returns more than 50 tickets, narrow it by milestone or by parent; when it still does, say in the report that the list was cut and by which filter.
3. **Open only the matches.** Read whole, with its comments, each ticket whose title or labels match a capability of step 1, and each ticket in progress in a module this ticket changes: five at most, the nearest first by blocking relation and then by milestone, naming the rest without opening them. A list of five tickets or fewer may be opened whole. Do not open the tickets of a longer list one by one: the query exists so the rest are judged by their title and labels.

## What a match means

Classify each ticket opened. Report each class among the dependencies of the assessment ([hive-mode](../../hive-mode/SKILL.md#assess-a-chosen-ticket)) or in the plan's verified context, next to the relations the ticket already declares:

- **A later ticket owns a capability this one needs.** Ask it as a decision, as [questions](../messages/questions.md) sets, with this recommendation: star pulling the work into this ticket when a criterion of this ticket cannot be observed without the capability and deferring it would make this work build a substitute the later ticket replaces, or change again the files this work changes; star deferring otherwise. An external effect the pulled work adds does not change the star, since the [effects round](../plan/external-effects.md) asks it apart. Each option's description names the tickets it would comment on or edit.
- **A neighbor in progress touches the same resource**, such as the same migration number, sequence or module. Name the ticket and the resource in the report, do not take that ticket's work, and pick a number or name that does not collide with the one it holds.
- **No ticket owns a capability this one needs.** Say so. It joins this ticket when it passes the test of the first class, asked as that decision; otherwise propose its ticket as [scope changes](../plan/scope-changes.md) sets.

A query that finds no match reports that, with the filter it used.

## Leave a trace

When a decision or the delivered change pulls forward, defers or delivers work that another ticket of the same project states, comment on that ticket, in its language: what moved, from which ticket to which, and the pull request once it exists. The global guidance's «Scope and authorization» authorizes this comment, and [trackers](trackers.md) sets how to write it. Write it:

- when the build starts, with the ticket's move to in progress, for a decision taken in research or in the plan, so work that never reaches a build leaves no comment;
- in the same turn, for a decision taken during the build;
- at the delivery point, when the change delivered part of another ticket's scope without a decision.

Edit another ticket's description only when the option the user chose names that edit and the round's report showed each section it changes, as it is and as it would be, which is the confirmation [existing ticket](../../create-issue/references/existing-ticket.md) asks for. Never edit a ticket the authenticated account did not write: comment instead.

A child neither asks the decision nor writes the comment: it returns the decision as one its parent must ask, with the comment's text, and the parent writes it. An operational task writes none unless the user asks.
