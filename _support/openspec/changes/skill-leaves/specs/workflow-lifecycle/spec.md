## ADDED Requirements

### Requirement: Entry router with playbooks
Hive SHALL provide an entry skill, `hive-mode`, that replaces `flow-research`. It SHALL read before asking, play the request back, ask what is missing and settle what the user needs, and SHALL NOT edit repository files itself; for a small understood change it MAY hand the work to `flow-build` in the same turn. It SHALL classify the work and open at most one playbook under `playbooks/<kind>.md`, record its name in the change folder's `proposal.md` (written by `flow-plan`) or, for work without a change folder, in a `- Playbook:` line of the checkpoint file, and, when no playbook fits or the work is large and cross-cutting, follow the general route from research to plan or build. A playbook names the order of the stages, what is specific to that kind and the leaves it uses; the stages (`flow-figma-design`, `flow-plan`, `flow-build`, `flow-verify`, `flow-close`, `flow-deploy`) SHALL remain commands that read the playbook's specific steps.

#### Scenario: A reported defect with a fix request
- **WHEN** the user reports a defect and asks for its fix
- **THEN** `hive-mode` opens the `bugfix` playbook, and build and verify fix the defect at its cause as «Defects are fixed at their cause» sets, with its regression evidence.

#### Scenario: An investigation confirms a defect
- **GIVEN** work running the `investigate` playbook
- **WHEN** the findings confirm a defect and the user asks for its fix
- **THEN** the work hands off explicitly to the `bugfix` playbook as a scope change, and the new playbook is recorded

### Requirement: Verify complements and adversarial review option
Before the code review, `flow-verify` SHALL read the leaves that remove narrating comments, check the change's impact beyond the diff, and check the prose of the pull request, commits and documentation. The review question SHALL offer an adversarial cross-vendor review as an option that states it spends another provider's quota, under the reuse and spending limits the user set for it; when another provider's command-line agent is unavailable, verify SHALL report it as unavailable and run the chosen review only.

#### Scenario: The adversarial review is chosen
- **WHEN** the user picks the adversarial review in the review question for a work item
- **THEN** verify runs it on that work item's candidate within its limit, reruns each reported finding before keeping it, and runs no such review on work items where it was not chosen.

### Requirement: Decision log for unattended and coordinated work
Unattended work and bounded E2E coordination SHALL keep one decision log at `_support/workspace/YYYY-MM-DD-<work>/decisions.tsv`, a tab-separated file with the columns `ts`, `phase`, `decision`, `why`, `evidence` and `result`, one single-line row per decision and one per external effect, whose evidence names the target and the change and whose result names how to revert it. Evidence SHALL be a link or path, never a paragraph. Its format SHALL have one home, in `unattended-delegation`, which E2E coordination links; in E2E only the coordinator writes it. The log SHALL NOT be committed or published unless the user asks, and the final report SHALL cite it by path. The task record SHALL keep only progress.

#### Scenario: An unattended run takes an external effect
- **WHEN** an unattended run changes a resource outside the repository within its authorized ceiling
- **THEN** `decisions.tsv` gains one row whose evidence names the target and the change and whose result names how to revert it, and the task record gains no copy of it

## MODIFIED Requirements

### Requirement: Plan revisions keep the change folder coherent
A revision of a written plan that has a change folder SHALL edit in the same pass every artifact that states what changed, mark a replaced decision or criterion with `**Replaces:** <old IDs>` and a dropped criterion with `**Retired:** <reason>` without deleting the old line, reopen each completed task the revision invalidates, and run the coherence check (`flow-plan/scripts/check_change.py`, which also validates the canonical grammar of lines under `## Business rules` in `proposal.md` and of the delivery-unit boxes under `## Delivery units` in `tasks.md` when run without `--archive`) until it exits 0 before work continues; a revision that changes a criterion or the delta SHALL get one `hive-review-plan` round.

#### Scenario: A task still names a replaced decision
- **GIVEN** a change folder whose revision replaced a decision and marked the new one with `**Replaces:**`
- **WHEN** the coherence check runs and a task still names the replaced decision
- **THEN** the check prints that task's file and line and exits 1, and the revision is not complete until it exits 0

#### Scenario: A revision changes a criterion
- **GIVEN** a change folder whose plan was already reviewed
- **WHEN** a revision changes a criterion or the delta
- **THEN** one `hive-review-plan` child reviews the revised folder, and a gap that survives the round goes to the user as a decision

#### Scenario: Malformed business rule line in proposal
- **GIVEN** a change folder whose `proposal.md` holds a line under `## Business rules` that violates the canonical prefix or field grammar
- **WHEN** `check_change.py` runs without `--archive`
- **THEN** the check prints the file and line of the malformed entry and exits 1

#### Scenario: A live box without its predicate
- **GIVEN** a `tasks.md` whose `## Delivery units` section holds `- [ ] live: <scenario>` without `· pass: <predicate>` and without `n/a:`
- **WHEN** `check_change.py` runs without `--archive`
- **THEN** the check prints the file and line and exits 1

### Requirement: Task progress written by a script
When a task of a change folder is done, the agent SHALL record it with `flow-plan/scripts/tasks.py`, which ticks only that task's boxes, adds one dated line with its commit and what stayed unverified under the progress section without repeating an existing line, and prints the task count; the agent SHALL keep the build checkpoint of a change folder in a `## Checkpoint` section of `tasks.md` written by the same script with its stage, step, last commit, open findings and next action, and, at a pause, the optional list of what is not yet written to any file, followed by an `### Evidence` list of checks that a later update keeps. An optional field SHALL be kept by a later call that omits it and removed by the value `none`; existing calls without the new options SHALL behave as before. Delivery-unit boxes under `## Delivery units` SHALL NOT be ticked by recording a task. The coherence check SHALL exempt the checkpoint section as it exempts the progress section. When the script cannot run, the agent SHALL make the same edits by hand; a plan kept in the conversation keeps its checkpoint file as before.

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

#### Scenario: A pause records what is not yet written
- **GIVEN** a checkpoint written at a pause with `--unwritten "the retry idea for T4"`
- **WHEN** a later call updates only the stage
- **THEN** the unwritten field is still there, and a call with `--unwritten none` removes it

### Requirement: Archive confirms unchecked tasks
Before merging a change folder's deltas into the specs and moving it to the archive, the close SHALL run the coherence check with `--archive` and fix each coherence line it prints; a delivery-unit box whose predicate reads `n/a: <reason>` SHALL count as resolved; when the check lists unchecked task boxes, the close SHALL show their count and ask whether to archive anyway before moving the folder, and a close under unattended delegation SHALL leave the folder unarchived and record it as pending.

#### Scenario: Unchecked boxes remain at archive
- **GIVEN** an integrated change folder whose `tasks.md` holds unchecked boxes
- **WHEN** the close runs the check with `--archive`
- **THEN** the close shows the number of unchecked boxes and asks before archiving, and the folder moves only on the user's answer

#### Scenario: An unattended close with unchecked boxes
- **GIVEN** a close running under unattended delegation and a folder with unchecked boxes
- **WHEN** the check lists them
- **THEN** the folder stays unarchived and is recorded as pending

#### Scenario: A delivery-unit box that does not apply
- **GIVEN** a `## Delivery units` box `- [ ] perf: n/a: no hot path changes`
- **WHEN** the close runs the check with `--archive`
- **THEN** that box is not counted as unchecked
