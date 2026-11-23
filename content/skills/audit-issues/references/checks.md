# Checks

Read at [audit-issues](../SKILL.md) step 2, before reading the tickets of a pass, and give it whole to each child that reads a lane. It sets each ticket's verdict (checks 1 to 8), the tickets no write touches, the authorship rule, and how tickets form groups (checks 9 to 14).

## Evidence

Every verdict carries its evidence, one of: a file and line at the base branch's remote commit, as [audit-issues](../SKILL.md) step 2 fetches it, a commit or merged pull request, a test that covers the path, a decision record entry, or the canonical ticket's link. A verdict without evidence is `not confirmed`. Read the code at that commit, not the ticket's thread alone: reports go stale when later commits remove or change the code they cite.

## Per ticket

Apply the checks in order; the first one that gives a closing or a non-`valid` verdict decides it, except check 7, which only reports.

1. **Valid.** The code the ticket names (a path, a symbol or a behavior) exists at the current commit and behaves as the ticket says. When the ticket names no code, find it from the behavior it describes; when you cannot, the verdict is `not confirmed`.
2. **False positive.** The symptom does not hold.
   - Propose `close: false positive` only when the code shows the failure cannot happen, such as a guard, a constraint or a type that rules it out, cited by file and line.
   - Reproduce only when the verdict would close a Bug: run the ticket's steps, or the nearest test that exercises its path, on that commit, in a local or disposable environment. Steps that need shared, staging or production data are not run: the Bug stays `not confirmed`, and the report names the environment the reproduction needs. A reproduction that fails to show the defect does not prove it absent, since reproducing a reported defect automatically fails in most cases: a Bug whose reproduction fails and whose code does not rule the failure out stays `not confirmed`.
3. **Already fixed.** A merged pull request or a commit changed the failing path after the ticket was filed, or a test now covers it. Cite it. The verdict is `close: already fixed`.
4. **Duplicate.** Search the whole tracker, open and closed tickets, not only the pass, by the distinctive terms of the outcome and of the place it touches, as [create-issue](../../create-issue/SKILL.md) step 2 searches.
   - It is a duplicate only with firm evidence: the same error text, the same steps to reproduce, or the same cause shown in the code. Text similarity alone is not evidence, since search ranks a likely duplicate high without showing that the two reports are the same defect.
   - Without that evidence, the two tickets get a `related` relation and both stay open.
   - The canonical ticket is open: when the matching ticket is closed, the open one reports a defect that returned or was never fixed, so it stays open with a `related` relation to the closed one, as create-issue step 2 relates a returning defect. Among open matches, the canonical one holds the most evidence (steps, logs, discussion); on a tie, the older one. The verdict for the other is `close: duplicate of <canonical>`, and its closing comment names the canonical ticket and the evidence that makes them one defect.
5. **Covered by a change in progress.** An open change folder under the repository's specs, or an open pull request, already changes the path the ticket names. The verdict is `covered by <change>`: the ticket stays open and is related to the change's ticket, or gets a comment with the pull request's link when the change has no ticket. [Ticket state](../../hive-core/tracker/ticket-state.md) closes it at the change's integration only when the change names it among its tickets; otherwise a later pass gives it `close: already fixed`.
6. **Contradicts a recorded decision.** The project's decision record, or an architecture decision record, states that the behavior is intended. This is a false positive with that evidence: `close: false positive`, closed as not planned with the link to the decision.
7. **Type and severity.** The ticket's type and labels match its real effect, such as an authorization finding labelled as technical debt. This check only reports: the report names the type or label that would match, and when the ticket already gets a comment for its author, that comment names it too. It adds no write of its own, and no ticket is relabelled.
8. **Ready.** Apply the ready list of [ticket types](../../create-issue/references/ticket-types.md). When it fails, the verdict is `not ready`: the report names each failed check, and nothing is corrected here; validating the ticket is a later [existing ticket](../../create-issue/references/existing-ticket.md) run.

## Verdicts

| Verdict | Write |
| --- | --- |
| `valid` | None, except moving it under its group's parent |
| `close: false positive` (includes a behavior a recorded decision states as intended) | Close as not planned, with the evidence or the decision's link |
| `close: already fixed` | Close as completed, with the pull request or commit |
| `close: duplicate of <canonical>` | Close as a duplicate, linked to the canonical ticket |
| `covered by <change>` | Related to the change; the ticket stays open |
| `not confirmed` | None |
| `not ready` | None |

Every close carries a comment with its evidence, written before the close.

## Exclusions

These tickets get their verdict in the report and no write:

- a ticket in progress, or one with an open branch or pull request, since someone is working it;
- for a move, a ticket that already has a parent: its group reuses that parent when the whole group fits under it, otherwise the ticket gets a `related` relation to the group's parent, since GitHub and Linear allow one parent per ticket and replacing it would undo someone else's grouping.

## Authorship

Decide whether the user wrote a ticket as [existing ticket](../../create-issue/references/existing-ticket.md) step 3 sets; a ticket the user states is theirs counts as theirs. For a ticket the user did not write, this skill changes two parts of that reference's step 4, which otherwise says not to edit it and to post one comment after the user confirms:

- Within the approved set, it may be moved under a parent and given a `related` or blocked-by relation; its text, type, labels and state stay as they are.
- A closing verdict does not close it: it gets one comment for its author with the verdict and its evidence, and stays open.

The set's approval is the confirmation for that comment.

In Jira, propose a close only when you can read that the ticket's workflow has a transition back from the closing state, as [tracker writes](tracker-writes.md) «Jira reopen» sets; otherwise the ticket gets the comment instead, also when the user wrote it.

## Per group

9. **Key.** Group by the first key that holds, and name the level in the group: root cause (the same defect in the code, shown by file and line), then domain concept (such as time zones or permissions), then module or view. A theme shared without a common cause or module gives a `related` relation between the tickets, not a group, since grouping by theme alone puts unrelated changes under one parent.
10. **Fits one change.**
    - A root-cause group is a parent whose children one change closes: one repository, one pull request and at most 7 acceptance criteria, the limit of create-issue step 3. This is an exception to [breakdown](../../create-issue/references/breakdown.md) «The set», where each child is a change integrated and accepted on its own: here the children are reports of one defect, and the plan of that one change takes the parent as its ticket and closes every child. The set's other rules hold, including at least two children.
    - When a root-cause group does not fit, it becomes a domain or module group, whose children are delivered separately as «The set» sets.
11. **Order.** Record a blocked-by relation between two members only when one cannot start before the other finishes; leave the others unordered.
12. **Existing parent.** When a ticket or an umbrella ticket already covers the group, reuse it as the parent, unless its own verdict closes it: a ticket the set closes is not reused as a parent, since closing it can close the valid children moved under it, so the group gets a new parent. A new parent is drafted as create-issue steps 2 and 4 to 6 set: its outcome is the one the members share, it lists what is out of its scope, and it is shown in the set before any write.
13. **Priority between groups.** Order the groups and the ungrouped tickets as point 4 of [ticket status](../../hive-mode/playbooks/ticket-status.md) «Report» orders recommendations.
14. **Ungrouped.** A ticket that shares neither a cause, a domain concept nor a module with another ticket of the pass stays on its own, with no parent.
