## MODIFIED Requirements

### Requirement: Close cleanup sweeps the repositories the work touched
The close's branch cleanup SHALL run in each repository where the work created a branch, a worktree or a commit or wrote its change folder, plus the repository that holds `Specs`, using `flow-close/scripts/cleanup_branches.py` when `python3` runs it to compute the dry-run cleanup plan before applying deletions with `--apply`; a hygiene request SHALL cover the current repository, its `Specs` repository and the repositories that the pending-close records in persistent memory name. Every worktree removal SHALL keep the worktree and both copies of its branch when it has uncommitted or untracked files, or an ignored entry other than these three: a dependency or build folder or an entry in one; an entry whose content the main checkout of the worktree's own repository already holds at the same path; and, at any depth, `*.tsbuildinfo`, `next-env.d.ts`, `*.pyc`, `.turbo/`, `__pycache__/`, `.pytest_cache/`, `.mypy_cache/`, `.ruff_cache/` or `.DS_Store`, or an entry inside one of those folders; it SHALL first stop the processes the work started inside its own worktree, by process ID, and SHALL keep any worktree where a running process has its working directory or where that cannot be checked. A squash- or rebase-merged branch SHALL count as merged only when its pull request's `headRefOid` equals the head of each local and remote copy to be deleted; a remote copy that moved SHALL be reported with the commits after that head. After the work's own branch, the sweep SHALL check the other branches whose pull requests the user authored and merged into the base, the local branches `git branch --merged` lists that a close record in persistent memory, or a plan or checkpoint, names for the user's earlier work in a repository without pull requests, and the temporary or preview branches the work created. It SHALL NOT switch any checkout's branch, and SHALL keep, reporting the reason, a branch checked out in another checkout and a branch that persistent memory, a plan or a checkpoint names for work whose close is not recorded as done, unless a pending-close record for that work is older than 24 hours, in which case it SHALL resume that work's pending close steps in order instead of removing its branch or worktree. The global guidance SHALL authorize deleting merged work branches from this work or the user's earlier work in those repositories, with their clean worktrees, and SHALL keep forbidding the deletion of unmerged branches and of worktrees with uncommitted or untracked files. The deferral of cleanup to the close footer's reply SHALL be unchanged.

#### Scenario: A leftover worktree of earlier work
- **WHEN** a close or hygiene request runs and the specs repository holds a clean worktree whose branch the user merged into the base in earlier work whose pending close is older than 24 hours
- **THEN** that work's pending close steps run in order, including its support decision, and its worktree and branch are removed by its own cleanup

#### Scenario: A merged worktree with only disposable ignored entries
- **WHEN** a merged branch's clean worktree holds only ignored `__pycache__/`, `.DS_Store` and a `.env` identical to the main checkout's
- **THEN** the cleanup removes the worktree and deletes both copies of its branch

#### Scenario: A worktree still in use
- **WHEN** a merged branch's worktree is the working directory of a running process, holds an ignored entry other than those three, or belongs to work in progress or whose close has been pending for less than 24 hours
- **THEN** the cleanup keeps it and reports the reason

#### Scenario: A remote copy moved after the merge
- **WHEN** the remote copy of a merged branch holds a commit pushed after the merged head
- **THEN** that copy is kept and reported with its unmerged commits

#### Scenario: Another session's fresh worktree
- **WHEN** a hygiene sweep finds a clean worktree on a branch with no commits beyond the base, and that branch is named in the checkpoint of work whose close is not recorded as done
- **THEN** the sweep keeps the worktree and its branch and reports the reason

### Requirement: Plan revisions keep the change folder coherent
A revision of a written plan that has a change folder SHALL edit in the same pass every artifact that states what changed, mark a replaced decision or criterion with `**Replaces:** <old IDs>` and a dropped criterion with `**Retired:** <reason>` without deleting the old line, reopen each completed task the revision invalidates, and run the coherence check (`flow-plan/scripts/check_change.py`, which also validates the canonical grammar of lines under `## Business rules` in `proposal.md` when run without `--archive`) until it exits 0 before work continues; a revision that changes a criterion or the delta SHALL get one `hive-review-plan` round.

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

### Requirement: Git and pull-request state read in one call
Before comparing a work branch with its base, before merging and when reading a pull request's reviews, the agent SHALL read the Git and pull-request state with `git-workflow/scripts/git_state.py` when `python3` runs it, and SHALL fall back to the skill's manual commands when the script is missing or exits 2. Invoked without a subcommand, the script SHALL fetch before any comparison when asked to and report no comparison when that fetch fails; report the commits on the branch that the base lacks with their author without deciding which belong to the work; report whether the local head and the pushed head contain the base; read the pull request, its checks and its reviews and comments once each through REST, without waiting for CI or polling; report for each review and comment its author type, whether it is on the pull request's head commit and whether it is later than the push time the agent passes, without deciding whether it still applies; keep the Git fields when the pull-request source fails; never prompt for credentials; bound each command's time; run no shell; and change nothing but the remote-tracking refs a fetch updates. Invoked as `git_state.py commit`, the script SHALL reject wildcard or whole-tree flags (`-a`, `-A`, `.`), hook-skipping flags (`--no-verify`, `-n`) and directory paths, record pre-existing staged paths, temporarily register untracked target files with `git add -N` (rolling them back on failure), scan the target paths' diff for secrets, enforce the `Unreviewed: <reason>` trailer when `--unreviewed-fix` is passed, run `git commit --only -- <paths>`, and verify that pre-existing staged paths outside `<paths>` remain staged. The skill SHALL keep applying the head-commit and push-time filter only to review bots, and an open human thread SHALL still apply.

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

#### Scenario: Path-scoped commit preserves unrelated staged files
- **GIVEN** a repository with `a.txt` already staged in the index and `b.txt` modified
- **WHEN** the agent commits `b.txt` through `git_state.py commit`
- **THEN** `b.txt` is committed after passing the secret scan and `a.txt` remains staged in the index

### Requirement: Business rules a work item adopts are recorded and reach its ticket

A business rule SHALL be a behavior of the product in a situation: what an action requires, who may take it, which state follows it, what it blocks or keeps, or where a setting lives. Each rule a work item decides or assumes that its ticket and the documentation the work read do not state, in any flow, SHALL be recorded in the work's rule record in the same turn: one line naming the rule, the screen it shapes when it has one, its status, and the flow that found it. A rule adopted without the user's decision, by the main thread or by a child that reports it, SHALL be recorded as an assumption and stated in the main thread's next message. A rule that replaces one the ticket, the documentation or the record states SHALL keep the earlier rule's line, marked cancelled with the date, the reason and the rule that replaces it; no line SHALL be deleted. The record SHALL live in the example's record of the design stage (`examples.md`) while the work has no change folder, in the proposal of its change folder once it has one, which then governs, and, for work without a change folder, in the work's checkpoint file once its plan or build starts. The main thread SHALL write the record's rules to the work's ticket when the user approves a design example, when the build starts and at the work item's completion point, using `flow-plan/scripts/business_rules.py` when `python3` runs it to render and splice the section between its localized start and end markers while verifying that all text outside the markers stays byte-for-byte identical: a section of the description, between a heading and a closing line of its own, that it alone replaces, leaving the rest of the description unchanged, holding the rules decided, the assumptions, the rules still to decide and the rules cancelled with their reasons, and a comment naming what that write added and what it cancelled, with each reason. It SHALL skip a write only when the groups and the design location it would write equal those the ticket holds, ignoring the generated write date, and SHALL write nothing to a tracker when the work has no ticket or was abandoned; a close that writes nothing SHALL list the rules in its report. These writes SHALL take no question, as an exception the guidance states to its rule for effects on data the task did not create. Research SHALL read a ticket with its comments and SHALL NOT write the ticket.

#### Scenario: A rule decided during the plan

- **GIVEN** a ticket that does not say whether a search matches by plate or by any field
- **WHEN** the user decides it in the plan and later starts the build
- **THEN** the rule is a line of the proposal's rule record in the turn the user answered
- **AND** the build's start writes it in the ticket's rules section and adds a comment naming it

#### Scenario: A decision replaces a rule the ticket states

- **GIVEN** a ticket whose description says an evidence can satisfy a required document
- **WHEN** the user decides during the plan that it cannot
- **THEN** the record keeps the ticket's rule as a cancelled line with the date, the reason and the rule that replaces it
- **AND** the next write lists it among the cancelled rules of the section, leaves the original sentence of the description unchanged, and its comment says what was cancelled and why

#### Scenario: A child assumes a rule during the build

- **GIVEN** a build child that had to choose which stage a new record starts in
- **WHEN** it reports that choice
- **THEN** the main thread records it as an assumption and states it in its next message
- **AND** the child writes neither the record nor the ticket

#### Scenario: Work without a ticket

- **GIVEN** a work item that names no ticket
- **WHEN** it reaches its close with rules in its record
- **THEN** no tracker is written and the close's report lists the rules

#### Scenario: Nothing changed since the last write

- **GIVEN** a record whose last ticket write was at the build's start
- **WHEN** the section the close would write equals the one the ticket holds
- **THEN** it writes neither the section nor a comment

#### Scenario: The design location changes without changing the rules

- **GIVEN** a ticket section holding the rules and the previous design location
- **WHEN** the user approves a replacement design at another location with the same rules
- **THEN** the next write updates the design location instead of skipping the write

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
- **THEN** `flow-research` reports its impact, complexity and dependencies, and judges its readiness with the skill's readiness list.

#### Scenario: A session without a repository
- **WHEN** no `## Hive` section declares a tracker for the session
- **THEN** the skill asks which tracker and repository to create the ticket in before it searches or writes, and omits the solution hint, saying so in the ticket.

### Requirement: Requirement documents are written and validated through one procedure

The workflow guidance SHALL keep the procedure for writing, restructuring, updating and technically reviewing a requirement document in the `create-prd` skill. The skill SHALL write the document with these sections in order: problem, goals, non-goals, user stories, requirements, acceptance criteria, success metrics, open questions and version history, and SHALL NOT write interface design, the technical how or a schedule. Each requirement SHALL be one rule written as an entry of its own, never as a row of a table, with a stable code that is never reused, a priority of P0, P1 or P2, a state (decided, assumption, to decide or cancelled), its source and its verdict against the code (exists, new, contradicts, blocked or not checked). The skill SHALL check the draft against the fixed readiness checklist with `create-prd/scripts/check_prd.py` when `python3` runs it before launching `hive-review-prd`, SHALL ask at most three questions per round and SHALL never assume a business rule the user does not decide. Used on its own, it SHALL end at the document and only offer the next step. `flow-research` SHALL serve a request whose deliverable is a requirement document by following the skill in the same turn, whether or not it asked questions first, and the global guidance SHALL name that hand-over among the continuations that need no reply from the user.

#### Scenario: A requirement is drafted
- **WHEN** the user asks for the requirement of a module or feature
- **THEN** the skill reads the source, asks at most three questions per round, writes the document in the fixed sections with each requirement as a rule entry, and reports its link, its state, its open questions with who answers each and its assumptions.

#### Scenario: A request enters through research
- **WHEN** a request whose deliverable is a requirement document reaches `flow-research`
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
