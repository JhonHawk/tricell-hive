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

### Requirement: Tickets are created through one procedure

The workflow guidance SHALL keep the procedure for creating, breaking down and validating a tracker ticket in the `create-issue` skill, and every other flow SHALL point to it instead of stating its steps. The global guidance SHALL keep when a ticket is proposed and that creating one needs the user's answer. The skill SHALL create a ticket of one of four types (story, bug, task, spike) with that type's sections, SHALL ask at most three questions per round, each with a recommended answer, SHALL write what it does not ask as an assumption in the ticket, and SHALL check the ticket against its fixed readiness list in the same thread using `create-issue/scripts/check_issue.py` when `python3` runs it, without a reviewer child. Run on its own, the skill SHALL end at the ticket and only offer the next step; followed from a flow, it SHALL return to that flow's next step. It SHALL NOT add a statement that an agent wrote the ticket unless the target repository's contribution rules ask for one.

#### Scenario: A flow needs a ticket
- **WHEN** a flow's finding or a separate topic gets the user's answer to create a ticket
- **THEN** the flow follows the `create-issue` skill in the same turn, which searches every declared tracker for an existing one, creates no other when an open one exists, relates the new ticket to its origin and creates it in the primary tracker without another confirmation, shows the created ticket with its link, and returns to the flow's next step.

#### Scenario: The skill is used on its own
- **WHEN** the user asks for a ticket outside a flow
- **THEN** the skill shows the draft, creates the ticket after the user confirms, and ends by offering the next step.

#### Scenario: An undecided business rule
- **WHEN** a ticket is created while one of its business rules is undecided, or while it lacks a fact nobody in the session can supply
- **THEN** the missing part is not invented and the ticket is created marked as blocked, through the tracker's native blocked-by relation, a blocked label or state the project already has, or else the first line of the ticket with the question and who it waits on, and no label or state is created.

#### Scenario: A ticket chosen for research
- **WHEN** the user picks a ticket for research
- **THEN** `hive-mode` reports its impact, complexity and dependencies, and judges its readiness with the skill's readiness list.

#### Scenario: A session without a repository
- **WHEN** no `## Hive` section declares a tracker for the session
- **THEN** the skill asks which tracker and repository to create the ticket in before it searches or writes, and omits the solution hint, saying so in the ticket.

### Requirement: Requirement documents are written and validated through one procedure

The workflow guidance SHALL keep the procedure for writing, restructuring, updating and technically reviewing a requirement document in the `create-prd` skill. The skill SHALL write the document with these sections in order: problem, goals, non-goals, user stories, requirements, acceptance criteria, success metrics, open questions and version history, and SHALL NOT write interface design, the technical how or a schedule. Each requirement SHALL be one rule written as an entry of its own, never as a row of a table, with a stable code that is never reused, a priority of P0, P1 or P2, a state (decided, assumption, to decide or cancelled), its source and its verdict against the code (exists, new, contradicts, blocked or not checked). The skill SHALL check the draft against the fixed readiness checklist with `create-prd/scripts/check_prd.py` when `python3` runs it before launching `hive-review-prd`, SHALL ask at most three questions per round and SHALL never assume a business rule the user does not decide. Used on its own, it SHALL end at the document and only offer the next step. `hive-mode` SHALL serve a request whose deliverable is a requirement document by following the skill in the same turn, whether or not it asked questions first, and the global guidance SHALL name that hand-over among the continuations that need no reply from the user.

#### Scenario: A requirement is drafted
- **WHEN** the user asks for the requirement of a module or feature
- **THEN** the skill reads the source, asks at most three questions per round, writes the document in the fixed sections with each requirement as a rule entry, and reports its link, its state, its open questions with who answers each and its assumptions.

#### Scenario: A request enters through research
- **WHEN** a request whose deliverable is a requirement document reaches `hive-mode`
- **THEN** research follows the `create-prd` skill in the same turn, and the skill asks its own questions.

#### Scenario: A document arrives already written
- **WHEN** the user hands over a requirement document someone else wrote
- **THEN** the skill returns it in the fixed sections without changing any of its author's rules, with each rule pointing to where it stood in the original, delivers a questionnaire for the author with a recommended answer per question, and never edits the original.

#### Scenario: An undecided business rule
- **WHEN** a rule belongs to someone else or the user does not know it yet
- **THEN** the rule is written as to decide, with who answers it, and is never filled in.

#### Scenario: A correction arrives
- **WHEN** answers or corrections change the document
- **THEN** the skill writes a new version with its list of changes, and a cancelled rule keeps its entry with the version that cancelled it and the rule that replaces it.
