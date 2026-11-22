# Breakdown

Read at [create-issue](../SKILL.md) step 3, when the work names more than one change that integrates separately, touches more than one repository, or needs more than 7 acceptance criteria. It sets the parent ticket and its children, their approval and their writes, in place of that skill's steps 7 and 8; steps 4 to 6 apply to each ticket of the set.

## The set

- A child is a change that is integrated and accepted on its own, with its own outcome and criteria, and someone could deliver it and stop there; a layer, step or file that only makes sense inside another change stays in its ticket, whose plan turns it into a task. Cut children as slices that each deliver behavior end to end, not one per layer ("the API", "the screen").
- The parent states the outcome the children share, what is out of scope for the whole set, and its children. A set has at least two children; with one, write one ticket.
- Each child is drafted and checked as a ticket of its type, as [ticket types](ticket-types.md) sets, and names the repository it changes.
- When a child cannot start before another finishes, record that it is blocked by it; leave the others unordered.
- Several tickets for separate findings, each chosen by the user, are single tickets following the skill's steps 7 and 8; so is a group of findings a flow presents as one item (one ticket that lists its members).

## Approval

Show the whole set before any write: the parent, then each child with its type, title, outcome, criteria, repository and what it is blocked by, and which tickets would be created as blocked. Ask one approval for the whole set, as one decision whose options are to create it or to correct it; a correction changes the set, which is shown again. No ticket of the set is written before that approval, and an earlier answer asking for "a ticket" does not cover a set, also inside a flow.

While [unattended delegation](../../unattended-delegation/SKILL.md) is active, create no set: report the proposed set as unresolved. A single ticket for a finding is still created there without a confirmation.

## Writes

1. Create every ticket of the set in one tracker, the primary tracker of the repository the session works in or the one the user names for the set, also when its children change different repositories.
2. Write one ticket at a time, never in parallel, since trackers limit bursts of writes: the parent first, then each child linked to it, then the blocked-by relations, as [trackers](trackers.md) sets.
3. After each write, add the ticket's link to the list of what was created, in the message, and read the ticket back as the skill's step 8 sets.
4. When a write fails, stop: write nothing more, delete nothing, and report the tickets created with their links and the ones still missing, then ask how to continue.
5. Never repeat a write whose result you could not read. Search the tracker for the ticket by its title first, since it may have been created.
