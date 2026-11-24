---
name: swarm
description: "Coordinate two to four tickets in one session through the swarm mode: one hive-ticket-owner child per ticket researches, plans and builds in its own worktree, stopping so the coordinator can ask the user, and each ticket is verified and published on its own. Use only when the user types this skill's command or names the swarm mode; never on your own initiative, and not for a single ticket."
disable-model-invocation: true
---

# Swarm mode

The main thread is the coordinator of two to four tickets. For each ticket, one `hive-ticket-owner` child researches, plans and builds in a worktree the coordinator created, and stops where the user must decide; the coordinator asks the user, verifies each ticket independently and publishes it on its own branch and pull request. This skill and the `hive-ticket-owner` role hold every exception the mode makes to the ordinary flows: outside a session that started this skill, those flows apply unchanged.

The mode starts only from the user's command or a request in which the user names it. It is not [unattended delegation](../unattended-delegation/SKILL.md): the user's silence does not extend it, and it ends when every ticket is published, returned to the user or stopped, or when the user stops it.

## Exceptions the mode makes

Each holds only inside this mode, and each names what it changes:

- **Who plans.** The owner does [flow-plan](../flow-plan/SKILL.md) steps 1 to 3 for its ticket, which the main thread does in ordinary work. The coordinator keeps that skill's question rounds, its effects round and its plan review. A ticket's plan gets one `hive-review-plan` child only when it is a risky change, as the global guidance's «Proportionality» defines it, in place of the one review every plan with a change folder gets.
- **Who writes and commits.** The owner writes its ticket's change folder and its checkpoint, and creates local commits through the scoped-commit helper in its assigned worktree, branch and paths. When the project's specs live in another repository, the owner writes the change folder in the one change-folder path of that repository's main checkout the coordinator assigns, and never commits there; the coordinator commits it, as [git-workflow](../git-workflow/SKILL.md) step 3 sets for a change folder in a `direct-base` specs repository. This changes the global delegation rule that children stay in their working tree, [child brief](../hive-core/delegation/child-brief.md)'s rules that only an E2E owner commits and that an ordinary child works in the main thread's checkout, and [E2E coordination](../flow-build/references/e2e-coordination.md)'s rules that only `hive-build-owner` builds and commits and that the coordinator alone writes the plan, whose change folder never lies in an owner's worktree.
- **Who creates worktrees.** The coordinator creates one worktree per eligible ticket on the answer to round 0, which is the user's request for those worktrees. Git-workflow step 3 otherwise creates a worktree only for the work in progress.
- **Who captures a screen before it changes.** The owner of a ticket that changes a screen serves its worktree on the port the coordinator assigned and takes the before captures there, before its first edit of that screen, and the after captures once it is built. [Before captures](../hive-core/ui/before-captures.md) otherwise has the main thread capture and a child's brief ask for none.
- **The execution record.** Each ticket has its own [execution record](../flow-plan/references/execution-record.md), one unit with the role `hive-ticket-owner`, in the coordinator checkout's `_support/workspace`, created when the ticket's plan is approved rather than before the first dispatch.
- **What the owner delegates.** Nothing: it launches no children, so it does inline what an ordinary flow would give a child, and the coordinator launches every reviewer and verifier.
- **Task verification.** Each ticket gets one `hive-verify-task` child for all its tasks and criteria at the returned head, in place of the one per risky task [flow-build](../flow-build/SKILL.md) step 3 sets; its verdict is the unit's verifier on the record.

## 1. Round 0

Read each ticket with its comments, and the code it names, enough to give its verdict. A ticket is eligible when all of these hold:

- it passes the [ready list](../create-issue/references/ticket-types.md#ready-list);
- its estimated scope stays within one repository and ten files;
- it writes no stored data and runs no migration;
- when it changes a screen, the project's dev server accepts a port set by a flag or an environment variable, read from its own scripts;
- when the project's specs live in another repository, that repository declares `Delivery: direct-base`.

Then ask one round, as [questions](../hive-core/messages/questions.md) sets, with the verdict of each ticket in its report and each ineligible ticket named with the failed condition, left for its own session. The round asks: the eligible tickets to take, at least two and at most four; the [delivery and review](../hive-core/git/delivery-ask.md) of each repository they touch, unless answers are in force, with `hold` left out, since owners hand back committed heads (a repository whose answer in force is `hold` is asked again, and when the user keeps `hold`, recommend the ordinary flow); the caps on simultaneous owners (at most four) and on owner runs per ticket (recommend three: plan, build and one resume after a blocker); and the creation of one worktree per ticket. With fewer than two eligible tickets, recommend the ordinary flow instead of the mode.

## 2. Prepare

- Create each worktree and branch in the host's native location, as git-workflow step 3 sets for an accepted worktree.
- Install each worktree's dependencies before any owner serves or builds it; owners install nothing.
- Assign each screen ticket a distinct free port, recorded as its exclusive resource.
- Start one [decision log](../unattended-delegation/SKILL.md#decision-log) for the batch in the coordinator checkout's support home; only the coordinator writes it.
- Write the batch checkpoint, a file in that log's folder: each ticket with its worktree, branch, base commit, change-folder path, execution record once created, port, runs used and current stop, and the caps. The ticket's own `tasks.md` names its record only when both lie in the same checkout. Update it after every launch and every return.

## 3. Owner runs

Before every launch, read the caps from the batch checkpoint. Launch owners together up to the cap on simultaneous owners; the other tickets wait in the batch checkpoint, and each starts when a running owner stops. Launch the owners of a round together, as [delegation](../hive-core/delegation/delegation.md#when-children-run-at-the-same-time) sets for children that write disjoint files. Brief each as [child brief](../hive-core/delegation/child-brief.md) sets, naming its ticket, worktree, branch, base commit, change-folder path, allowed paths, port and the run it starts; the role's [owner stops](references/owner-stops.md) set what it does and returns.

To continue an owner after a stop, message the same child when the host documents resuming a child with its history; otherwise launch a new `hive-ticket-owner` that reads the plan and checkpoint. Either counts as a run.

## 4. At the plan stop

For each ticket that returns its plan:

1. Write each `Ruling:` line it returned to the decision log, one row each.
2. Check the worktree and, for an assigned path in another repository, `git status` of that path alone: a write outside the assigned paths blocks the ticket.
3. Run `python3 -I <flow-plan-dir>/scripts/check_change.py <folder>` on its change folder.
4. When the plan is a risky change, launch its `hive-review-plan` child, briefed as [reviewer brief](../hive-core/delegation/reviewer-brief.md) sets.
5. Ask the decisions of every ticket that has stopped, together, in rounds as questions sets: number them on through the session, and name the ticket in each question's context line. Ask each ticket's [effects round](../hive-core/plan/external-effects.md) in the same rounds.
6. Once a ticket's plan is answered and approved, move the ticket to in progress as [ticket state](../hive-core/tracker/ticket-state.md) sets, write its business rules as [ticket write](../hive-core/business/business-ticket.md) sets, create its execution record with the ticket's change-folder path among the unit's allowed paths when the folder lies in its worktree, commit a change folder that lies in the specs repository as git-workflow step 3 sets for a `direct-base` specs repository, and continue its owner.

A blocker an owner returns during its build follows steps 1, 2 and 5 for that decision alone, then continues the owner.

## 5. At the build stop

For each ticket that returns a head, write its `Ruling:` lines to the decision log and run `validate-handoff` on its record. Launch the ticket's one `hive-verify-task` child, briefed with every task and criterion of its plan at that head, and record its verdict as the unit's verifier as [E2E coordination](../flow-build/references/e2e-coordination.md#check-correct-and-integrate) sets. Then take the ticket through [flow-verify](../flow-verify/SKILL.md) with its worktree head as the candidate: the owner's own checks are not its evidence, and `hive-review-ux` reviews a ticket that changes a screen, served on that ticket's port. Publish the owner's before and after captures as [walkthrough](../hive-core/ui/walkthrough.md) sets, and carry them into the pull request's before and after table. Each ticket is integrated on its own branch and pull request under the delivery in force, and closed with [flow-close](../flow-close/SKILL.md). A ticket that fails or is blocked does not stop the others.

## Caps and end

When a ticket reaches its cap on runs, launch no further owner for it, and report it with its worktree, branch, last head and checkpoint, and the decision log's path; the other tickets continue. The final report lists each ticket with its state, pull request or reason, and what the user still owns.

## Recovery

After a compaction, read the batch checkpoint and each ticket's `tasks.md` checkpoint before any launch. Replace an owner that stopped answering as [E2E coordination](../flow-build/references/e2e-coordination.md#replace-and-recover) sets. A plan an owner revised after its record exists reports plan drift; reconcile it as the execution record sets before continuing.
