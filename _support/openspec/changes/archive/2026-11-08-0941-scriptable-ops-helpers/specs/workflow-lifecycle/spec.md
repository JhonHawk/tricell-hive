## ADDED Requirements

### Requirement: Git and pull-request state read in one call
Before comparing a work branch with its base, before merging and when reading a pull request's reviews, the agent SHALL read the Git and pull-request state with `git-workflow/scripts/git_state.py` when `python3` runs it, and SHALL fall back to the skill's manual commands when the script is missing or exits 2. The script SHALL fetch before any comparison when asked to and report no comparison when that fetch fails; report the commits on the branch that the base lacks with their author without deciding which belong to the work; report whether the local head and the pushed head contain the base; read the pull request, its checks and its reviews and comments once each through REST, without waiting for CI or polling; report for each review and comment its author type, whether it is on the pull request's head commit and whether it is later than the push time the agent passes, without deciding whether it still applies; keep the Git fields when the pull-request source fails; never prompt for credentials; bound each command's time; run no shell; and change nothing but the remote-tracking refs a fetch updates. The skill SHALL keep applying the head-commit and push-time filter only to review bots, and an open human thread SHALL still apply.

#### Scenario: A branch behind its base
- **GIVEN** a work branch 2 commits ahead and 1 behind its base on the remote
- **WHEN** the agent reads the state with `--fetch`
- **THEN** the output lists the 2 commits with their authors, reports 1 behind and that the head does not contain the base, and the agent names the commits it did not make before the push

#### Scenario: The fetch fails
- **GIVEN** a remote that cannot be reached
- **WHEN** the agent reads the state with `--fetch`
- **THEN** the comparison fields are empty, the output names the fetch error and exits 1, and the agent does not merge on that state

#### Scenario: A human comment on an older commit
- **GIVEN** a pull request whose head moved after a person commented on an older commit
- **WHEN** the agent reads the reviews
- **THEN** the comment is reported as not on the head, and the agent still treats its open thread as a review finding

#### Scenario: GitHub CLI unavailable
- **GIVEN** a repository where `gh` is missing
- **WHEN** the agent reads the state
- **THEN** the output keeps the Git fields, marks the pull request unavailable with its reason and exits 1, and the agent reads the pull request with the skill's manual commands

### Requirement: Task progress written by a script
When a task of a change folder is done, the agent SHALL record it with `flow-plan/scripts/tasks.py`, which ticks only that task's boxes, adds one dated line with its commit and what stayed unverified under the progress section without repeating an existing line, and prints the task count; the agent SHALL keep the build checkpoint of a change folder in a `## Checkpoint` section of `tasks.md` written by the same script with its stage, step, last commit, open findings and next action, followed by an `### Evidence` list of checks that a later update keeps. The coherence check SHALL exempt that section as it exempts the progress section. When the script cannot run, the agent SHALL make the same edits by hand; a plan kept in the conversation keeps its checkpoint file as before.

#### Scenario: A task is done
- **GIVEN** a change folder with four tasks, one already done
- **WHEN** the agent records the second task with its commit
- **THEN** only that task's boxes are ticked, one progress line names the task and the commit, and the script prints `✓ T2 · 2/4` with the task's title

#### Scenario: The checkpoint is updated after checks were recorded
- **GIVEN** a checkpoint whose evidence lists the full suite's command, result and commit
- **WHEN** the agent updates the checkpoint's stage and next action
- **THEN** the evidence line is still there

#### Scenario: The checkpoint names a replaced decision
- **GIVEN** a checkpoint whose open findings name a decision a revision replaced
- **WHEN** the coherence check runs
- **THEN** it exits 0 for that section
