---
name: audit-issues
description: Audit a tracker's open tickets against the current code, proposing to close the ones that no longer apply (false positive, already fixed, duplicate) and grouping the related ones under parent tickets by root cause, domain or module so they are fixed together, then close and move them after one approval of the whole set. Use when the user asks to audit, triage, deduplicate, group, consolidate or clean up a project's open tickets or issues; not for creating, breaking down or validating one ticket, which create-issue does, nor for a report of the project's status or what to work on next, which hive-mode's ticket status gives.
---

# Audit open tickets

Check each open ticket against the code as it is now, propose closing the ones that no longer apply, and group the rest under parent tickets so related work is planned and fixed together. The skill runs in the main thread, since a child never writes to a tracker; it changes no repository file and writes nothing to the tracker before step 6's approval. Read [checks](references/checks.md) before step 2: it holds the criterion of each verdict and group, the exclusions and the authorship rule. Read [tracker writes](references/tracker-writes.md) before step 5: it holds how each write is made in each tracker. The audit ends with its report, as [create-issue](../create-issue/SKILL.md) does, and starts no work on any ticket.

## Pass

1. **Count and suggest a pass.** Read the repository's `## Hive` settings for its trackers as [project settings](../hive-core/plan/project-settings.md) sets; in a session without a repository, ask which tracker and project before any read, as create-issue step 2 does. List the open tickets through the tracker's tool, reading only their fields: identifier, title, type, labels, project or module, state, assignee, author and parent. Read no ticket text and no code before the user picks the pass. Report:
   - the open tickets, and how many of them are Bugs;
   - the lanes, each with its count and its Bugs: a lane is a module, label or project the tracker already uses, otherwise a term several titles share;
   - the tickets step 3's exclusions will leave unwritten, by their fields.

   Suggest passes of at most 30 tickets each, one lane per pass, starting with the lane with the most Bugs, and ask which pass to run as one decision, as [questions](../hive-core/messages/questions.md) sets. The answer is the pass; the duplicate search of check 4 still reads the whole tracker.

## Verdicts

2. **Read each ticket of the pass** whole, with its comments and each document it links as far as the ticket depends on it, and the code it names at the base branch's remote commit, fetched before the first read and named in the set's header, not the commit the ticket was filed against, since reports go stale. Decide whether children do these reads as [delegation](../hive-core/delegation/delegation.md) «When a child does the work» sets. When they do, give each `hive-research` child at most 15 tickets of the pass, splitting a larger lane across children, briefed as [child brief](../hive-core/delegation/child-brief.md) sets, with [checks](references/checks.md) as its criterion. The child returns its rows in its final reply, since that role writes no file: one row per ticket as `identifier · verdict · evidence · group key`, plus each candidate duplicate it found. A child writes nothing to the tracker. Check each returned row's evidence before using it, as [review evidence](../hive-core/review/review-evidence.md) sets for a finding: a row whose evidence does not hold gets its ticket read again here, or `not confirmed`.
3. **Give each ticket a verdict** with checks 1 to 8 of [checks](references/checks.md), each with the evidence that check requires, then apply its «Exclusions» and «Authorship». Reproduce a defect only when its verdict would close a Bug, as check 2 sets.
4. **Group** the `valid` tickets with checks 9 to 14; the other verdicts take no group, as the verdict table of [checks](references/checks.md) sets: each group names its key, its level (root cause, domain or module), its parent and the order of its members. Find an existing parent first (check 12); draft a new parent as create-issue steps 2 and 4 to 6 set, with the group's shared outcome as its outcome.
5. **Read the child auto-close.** For each parent the set would close or that a root-cause group would close later, read whether closing it closes its open children in this tracker, as [tracker writes](references/tracker-writes.md) «Child auto-close» sets. It is `on`, `off`, or `unknown` with the place where it is configured.

## Approval

6. **Show the set and ask one approval.** Show the whole set in the message before any write, in this form, in the session language:

   ```
   Pass <lane> · <base>@<commit> · <n> tickets read · <n> to close · <n> groups · <n> unwritten

   <Group title> · <level>: <key> · parent <link | new: title> · child auto-close <on | off | unknown: where>
     <ID>  <verdict>  <evidence>  → <write>
     <ID>  <verdict>  <evidence>  → <write>

   Ungrouped
     <ID>  <verdict>  <evidence>  → <write>

   Unwritten
     <ID>  <reason: in progress | open branch or pull request | not confirmed | not ready>  <what it lacks>
   ```

   Each `<write>` names the exact write the ticket gets: `move under <parent>`, `close as <duplicate of ID | not planned | completed> + evidence comment`, `comment for its author`, `comment with <pull request link>`, `related to <ID>`, `blocked by <ID>`, or `none`. A new parent is shown with its title, type, outcome and what is out of its scope.

   Then ask one decision, as [questions](../hive-core/messages/questions.md) sets, whose options are to apply the set or to correct it, such as taking a ticket out of a group, changing a verdict or splitting a group. A correction changes the set, which is shown again whole. This approval names each write on each ticket and is the instruction on named targets that the global guidance's critical-effect rule asks for: closing or moving a ticket the task did not create changes shared data. No write happens before it; an earlier answer, also one asking for the audit, does not cover it.

   While [unattended delegation](../unattended-delegation/SKILL.md) is active, write nothing: report the set as unresolved, as create-issue's [breakdown](../create-issue/references/breakdown.md) «Approval» sets for a set of tickets.

## Writes

7. **Write the approved set** one ticket at a time, never in parallel, in this order: the new parents; each move under a parent; the relations (related, blocked by); each close, its evidence comment first; each comment for another author and each `covered by` comment. Make each write as [tracker writes](references/tracker-writes.md) sets, add its link to the list of what was written in the message, and read the ticket back, reporting a difference instead of writing again. When a write fails, stop as [breakdown](../create-issue/references/breakdown.md) «Writes» steps 4 and 5 set.

## Report

8. **Report** in the session language:
   - each ticket written, with its link and the write it got;
   - each group with its level, parent and members, in the order of check 13;
   - the tickets left unwritten, each with its reason and what it lacks: a `not confirmed` Bug names the reproduction that failed, a `not ready` ticket the ready-list checks it fails;
   - each child auto-close that was `unknown`, with where to set it.

   Then offer the next step as a reply of [the reply block](../hive-core/messages/report-endings.md#the-reply-block), such as planning the first group, with its parent as the work's ticket, or validating a `not ready` ticket through create-issue, and stop.
