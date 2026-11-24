# Owner stops

Read when you are a `hive-ticket-owner` child of the [swarm mode](../SKILL.md), before any other step, and again at the start of each run. Your brief names your group's tickets, one or more, worktree, branch, base commit, change-folder path, allowed paths, port and the run you start.

You work only in your worktree and in the change-folder path your brief assigns. You launch no children and ask the user nothing: every decision that is the user's goes back to the coordinator at a stop. Your own checks are not the ticket's verification; the coordinator's reviewers and verifiers are.

## Run 1: research and plan

1. Read each ticket of your group with its comments, the code they touch, the project's `AGENTS.md` and `## Hive` settings, and the specs the change touches, as [flow-plan](../../flow-plan/SKILL.md) step 1 sets.
2. Weigh the approaches as flow-plan step 2 sets, writing them in your return instead of asking. Measure each fact outside the repository yourself, as a read-only command and the excerpt of its output that shows the fact.
3. Write one change folder for the group at your assigned path as [change folder](../../hive-core/plan/change-folder.md) sets, named from its first ticket's key, with the acceptance criteria of every ticket, each written `- AC<n> (<key>) — …` with the key of the ticket it belongs to, and the task verification lines flow-plan step 3 sets, running each verification command once in your worktree. Leave `## Delivery` and `## Effects` for the coordinator.
4. Write the checkpoint with `tasks.py <folder> checkpoint`. When the change folder lies in your worktree, commit it with its checkpoint through the scoped-commit helper, so the worktree is clean; then stop with the plan return.

## Run 2: build

1. Read the plan, its decisions and the checkpoint in your change folder; an answer from the coordinator in your brief is a decision of the plan.
2. For a group with a ticket that changes a screen, serve your worktree on your assigned port and take the before captures of each region the plan changes, as [before captures](../../hive-core/ui/before-captures.md) sets for the main thread, before your first edit of it.
3. Build the tasks in the order of the dependencies between your tickets, each as [flow-build](../../flow-build/SKILL.md) step 2 sets, doing inline what it would delegate: its checks, its tests, and a commit per task through `python3 -I <git-workflow-dir>/scripts/git_state.py commit` with explicit paths, as [git commit](../../hive-core/git/git-commit.md) sets. Tick each task with `tasks.py <folder> done`. Install no dependency: a missing one is a blocker.
4. For a screen, take the after captures on your port at the widths, themes and states the plan names.
5. Stop processes you started, commit the change folder and its checkpoint when they lie in your worktree, check that `git status` shows the worktree clean, then stop with the build return.

## Rulings

A decision that criteria a to f of [questions](../../hive-core/messages/questions.md) do not send to the user, you take and record as one line of your next return:

```
Ruling: <what you decided> — <why> — <what it costs if wrong>
```

A decision that one of those criteria sends to the user is never a ruling.

## Returns

**Plan return**, at the end of run 1:

```
Stop: plan · <ticket keys> · <change-folder path>
Approaches: <each, with its tradeoff and the one you recommend>
Decisions: <each, in the questions text format: question, Situation, options with consequences, the recommended one>
Effects: <each external effect a task would take, with its target, or none>
Business rules assumed: <each, or none>
Paths to write: <every path the build will change>
External facts: <each as command · excerpt>
Verification commands: <each as command · result>
Risky: <yes or no, with the Proportionality category>
Rulings: <Ruling lines, or none>
```

**Build return**, at the end of run 2: the exact full head, the commands with their exit results, the evidence and capture paths, the criteria covered, what stays unverified, the processes stopped, and your `Ruling:` lines.

**Blocker**, at any moment a decision of the user's, a missing dependency or a write outside your allowed paths stops the work: the blocker, what you tried, the decision in the questions text format when there is one, your last commit and the checkpoint's next step, then your `Ruling:` lines. Commit the change folder and checkpoint when they lie in your worktree, and commit or describe what else you leave uncommitted; never discard it.
