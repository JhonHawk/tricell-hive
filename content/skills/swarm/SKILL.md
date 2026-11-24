---
name: swarm
description: "Coordinate two to four tickets in one session through the swarm mode: related tickets form groups, one hive-ticket-owner child per group researches, plans and builds in its own worktree, stopping so the coordinator can ask the user, and each group is verified and published on its own. Use only when the user types this skill's command or names the swarm mode; never on your own initiative, and not for a single ticket."
disable-model-invocation: true
---

# Swarm mode

The main thread is the coordinator of two to four tickets. Tickets that are related form a group, and a ticket with no relation is a group of one. For each group, one `hive-ticket-owner` child researches, plans and builds its tickets in a worktree the coordinator created, and stops where the user must decide; the coordinator asks the user, verifies each group independently and publishes it on its own branch and pull request. This skill and the `hive-ticket-owner` role hold every exception the mode makes to the ordinary flows: outside a session that started this skill, those flows apply unchanged.

The mode starts only from the user's command or a request in which the user names it. It is not [unattended delegation](../unattended-delegation/SKILL.md): the user's silence does not extend it, and it ends when every group is published, returned to the user or stopped, or when the user stops it.

## Exceptions the mode makes

Each holds only inside this mode, and each names what it changes:

- **Who plans.** The owner does [flow-plan](../flow-plan/SKILL.md) steps 1 to 3 for its group, which the main thread does in ordinary work. The coordinator keeps that skill's question rounds, its effects round and its plan review. A group's plan gets one `hive-review-plan` child only when it is a risky change, as the global guidance's «Proportionality» defines it, in place of the one review every plan with a plan file gets.
- **Who writes and commits.** The owner writes its group's plan file and its checkpoint at the one plan path the coordinator assigns in the coordinator's checkout, under its ignored `workspace/` folder, where nothing commits them, and creates local commits through the scoped-commit helper in its assigned worktree, branch and paths. This changes the global delegation rule that children stay in their working tree, [child brief](../hive-core/delegation/child-brief.md)'s rules that only an E2E owner commits and that an ordinary child works in the main thread's checkout, and [E2E coordination](../flow-build/references/e2e-coordination.md)'s rules that only `hive-build-owner` builds and commits and that the coordinator alone writes the plan, whose file never lies in an owner's worktree.
- **The delivery question.** Round 0 asks the delivery again for a repository that holds a ticket's code whose answer in force is `hold` or `direct-base`, and offers neither there, since each group hands back a committed head and is published on its own branch and pull request; [delivery questions](../hive-core/git/delivery-ask.md) otherwise asks once per session and repository.
- **Who creates worktrees.** The coordinator creates one worktree per group on the answer to round 0, which is the user's request for those worktrees. Git-workflow step 3 otherwise creates a worktree only for the work in progress.
- **Who captures a screen before it changes.** The owner of a group with a ticket that changes a screen serves its worktree on the port the coordinator assigned and takes the before captures there, before its first edit of that screen, and the after captures once it is built. [Before captures](../hive-core/ui/before-captures.md) otherwise has the main thread capture and a child's brief ask for none.
- **The execution record.** Each group has its own [execution record](../flow-plan/references/execution-record.md), one unit with the role `hive-ticket-owner`, in the coordinator checkout's `_support/workspace`, created when the group's plan is approved rather than before the first dispatch.
- **What the owner delegates.** Nothing: it launches no children, so it does inline what an ordinary flow would give a child, and the coordinator launches every reviewer and verifier.
- **Task verification.** Each group gets one `hive-verify-task` child for all its tasks and criteria at the returned head, in place of the one per risky task [flow-build](../flow-build/SKILL.md) step 3 sets; its verdict is the unit's verifier on the record.

## 1. Round 0

Read each ticket with its comments, and the code it names, enough to give its verdict and estimate the files it changes. A ticket is eligible when all of these hold:

- it passes the [ready list](../create-issue/references/ticket-types.md#ready-list);
- its estimated scope stays within one repository and ten files;
- it writes no stored data and runs no migration;
- when it changes a screen, the project's dev server accepts a port set by a flag or an environment variable, read from its own scripts;

Then propose the groups. Two eligible tickets are related when their estimated files cross, when they change the same module, or when one depends on the other. Files nearly every change touches, such as a changelog, lockfiles, manifests and generated indexes, do not count as crossing. Tickets whose estimated files cross go in one group or are not taken together. A group's tickets share one repository: related tickets in different repositories are not taken together: the round names the pair with that reason, and the user keeps at most one of them in the batch. A group holds at most three tickets and ten estimated files in all; a larger group is ineligible, named with that reason.

Ask one round, as [questions](../hive-core/messages/questions.md) sets, with the verdict of each ticket and each proposed group with its reason in its report, and each ineligible ticket or group named with the failed condition, left for its own session. The round asks: the groups to confirm or change, and the eligible tickets to take, at least two and at most four tickets in all, grouped or not; the [delivery and review](../hive-core/git/delivery-ask.md) of each repository that holds their code, unless answers are in force, with `hold` and `direct-base` left out as «The delivery question» sets (when the user keeps either, recommend the ordinary flow); the caps on simultaneous owners (at most four) and on owner runs per group (recommend three: plan, build and one resume after a blocker); and the creation of one worktree per group. With fewer than two eligible tickets, or when they all form one group, recommend the ordinary flow instead of the mode.

## 2. Prepare

- Create each group's worktree and branch in the host's native location, as git-workflow step 3 sets for an accepted worktree.
- Install each worktree's dependencies before any owner serves or builds it; owners install nothing.
- Assign each group with a screen ticket a distinct free port, recorded as its exclusive resource.
- Start one [decision log](../unattended-delegation/SKILL.md#decision-log) for the batch in the coordinator checkout's support home; only the coordinator writes it.
- Write the batch checkpoint, a file in that log's folder: each group with its tickets, worktree, branch, base commit, plan-file path, execution record once created, port, runs used and current stop, and the caps. The group's own plan file names its record from its `## Checkpoint`. Update it after every launch and every return.

## 3. Owner runs

Before every launch, read the caps from the batch checkpoint. Launch owners together up to the cap on simultaneous owners; the other groups wait in the batch checkpoint, and each starts when a running owner stops. Launch the owners of a round together, as [delegation](../hive-core/delegation/delegation.md#when-children-run-at-the-same-time) sets for children that write disjoint files: here the files are disjoint across worktrees, and a conflict in a common file is settled when the base is merged before each merge, as [merge gate](../hive-core/git/merge-gate.md) sets. Brief each as [child brief](../hive-core/delegation/child-brief.md) sets, naming its tickets, worktree, branch, base commit, plan-file path, allowed paths, port and the run it starts; the role's [owner stops](references/owner-stops.md) set what it does and returns.

To continue an owner after a stop, message the same child when the host documents resuming a child with its history; otherwise launch a new `hive-ticket-owner` that reads the plan and checkpoint. Either counts as a run of its group.

## 4. At the plan stop

For each group that returns its plan:

1. Write each `Ruling:` line it returned to the decision log, one row each.
2. Check the worktree with `git status`: a write outside the assigned paths blocks the group.
3. Compare its paths to write with those of every other group not yet integrated: the paths to write of a returned plan, else the allowed paths of its record, else its round-0 estimates; repeat the comparison for every group whose plan arrives later in the round. When they cross outside the common files, stop this group, stop the other at its next return as a blocker, and hold the publication of another group already past its build stop, then ask the user whether to regroup them or set one aside.
4. Run `python3 -I <flow-plan-dir>/scripts/check_change.py <plan-file>` on its plan file.
5. When the plan is a risky change, launch its `hive-review-plan` child, briefed as [reviewer brief](../hive-core/delegation/reviewer-brief.md) sets.
6. Ask the decisions of every group that has stopped, together, in rounds as questions sets: number them on through the session, and name the group's tickets in each question's context line. Ask each group's [effects round](../hive-core/plan/external-effects.md) in the same rounds.
7. Once a group's plan is answered and approved, move each of its tickets to in progress as [ticket state](../hive-core/tracker/ticket-state.md) sets, write its business rules to each ticket as [ticket write](../hive-core/business/business-ticket.md) sets, create its execution record, and continue its owner.

A blocker an owner returns during its build follows steps 1, 2 and 6 for that decision alone and step 3 for each path the blocker adds, then continues the owner unless step 3 stopped it.

## 5. At the build stop

For each group that returns a head, write its `Ruling:` lines to the decision log and run `validate-handoff` on its record. Launch the group's one `hive-verify-task` child, briefed with every task and criterion of its plan at that head, and record its verdict as the unit's verifier as [E2E coordination](../flow-build/references/e2e-coordination.md#check-correct-and-integrate) sets. Then take the group through [flow-verify](../flow-verify/SKILL.md) with its worktree head as the candidate: the owner's own checks are not its evidence, and `hive-review-ux` reviews a group that changes a screen, served on that group's port. Publish the owner's before and after captures as [walkthrough](../hive-core/ui/walkthrough.md) sets, and carry them into the pull request's before and after table. Each group is integrated on its own branch and a pull request that names each of its tickets, under the delivery in force, each ticket moves when it integrates, and the group is closed with [flow-close](../flow-close/SKILL.md). A group that fails or is blocked does not stop the others.

## Caps and end

When a group reaches its cap on runs, launch no further owner for it, and report it with its tickets, worktree, branch, last head and checkpoint, and the decision log's path; the other groups continue. The final report lists each ticket with its group, state, pull request or reason, and what the user still owns.

## Recovery

After a compaction, read the batch checkpoint and each group's plan-file checkpoint before any launch. Replace an owner that stopped answering as [E2E coordination](../flow-build/references/e2e-coordination.md#replace-and-recover) sets. A plan an owner revised after its record exists reports plan drift; reconcile it as the execution record sets before continuing.
