# workflow-lifecycle

## Purpose

The lifecycle of authorized work, its viewer record and deferred session hygiene.

## Requirements

### Requirement: Viewer finalization before deferred hygiene

The normal completion phase SHALL publish and read back the viewer's actual terminal or waiting lifecycle before the report's hygiene footer. Integrated or abandoned completed work SHALL close explicitly; a ready/human-review stop SHALL remain in_review, a real pause SHALL remain paused, and awaiting validation SHALL remain explicit. Support-material decisions, change-record archival and resource/branch cleanup MAY remain deferred without keeping a completed viewer record active. A later hygiene phase SHALL NOT duplicate an already confirmed closure. The procedure SHALL preserve the change-folder commit and intentional-pending-work rules of the prerequisite #311 candidate.

An unavailable viewer or missing producer authority SHALL preserve inputs, report the still-recorded state and continue independent work without starting the service, silently recovering authority or claiming a successful close. Upload cleanup SHALL retain its exact committed-readback requirement.

#### Scenario: Delivered work awaits optional cleanup
- WHEN authorized work reaches its completion point and the viewer accepts its close
- THEN the initial report observes the work closed while the hygiene footer can still wait for the user.

#### Scenario: An open pull request waits for merge
- WHEN delivery stops ready for the user's merge
- THEN the viewer remains in_review and is not made eligible for terminal autoarchive.

#### Scenario: Later hygiene
- WHEN the user requests the deferred hygiene
- THEN its support/archive/cleanup steps use the previous closure evidence without appending another closure.

### Requirement: Tickets follow the tracker their key declares
When a repository's `## Hive` declares `Other trackers`, the workflow guidance SHALL read and search tickets in every declared tracker, SHALL move and comment each ticket, and write its business rules, in the tracker that declares its key prefix, using the primary `Tracker` for a key no tracker declares, and SHALL create new tickets only in the primary tracker. Declaring a tracker in `## Hive` SHALL authorize those moves, comments and rule writes as it does for the primary tracker.

#### Scenario: Ticket of an additional tracker reaches a later environment
- **WHEN** a deployment is verified and `Completion` holds a value for the prefix `XYZ`
- **THEN** the guidance moves each `XYZ-` ticket whose change that ref holds to the declared state in the tracker that declares `XYZ`.

#### Scenario: Mirror tickets
- **WHEN** one work item names a ticket in the primary tracker and another in an additional tracker
- **THEN** the guidance moves and comments both, and writes the work's business rules to both, each in its own tracker.

#### Scenario: New ticket
- **WHEN** a flow proposes or creates a ticket in a repository with additional trackers
- **THEN** it searches every declared tracker for an existing one, as the `create-issue` skill sets, and the new ticket is created in the primary tracker.

### Requirement: UX review in one pass

A UI change's `hive-review-ux` child SHALL receive, with the walk's captures, the means to reach each state its role always checks, or the states that cannot be reached. It SHALL list a state it cannot reach as not covered instead of failing its verdict. Its verdict SHALL judge only the defects the change introduced in the reviewed regions: passed when the reviewed states show no introduced defect of any priority, and not passed when no state could be reviewed. Defects already there SHALL be reported in a separate list. A correction of one of its findings SHALL be reread by a follow-up limited to the corrected regions, the layouts that contain them and their loading, empty and error states, with the automated audit and the layout, keyboard, focus and accessible-name checks on the longest values, never by a new full UX review. In a deployment walk, the reviewer SHALL get the same brief with the environment's URLs in place of the captures.

#### Scenario: State the brief cannot reach

- **GIVEN** a UI change whose error state needs a fixture the brief does not provide
- **WHEN** the UX reviewer cannot reach that state
- **THEN** it lists the state as not covered with what blocked it, and its verdict covers the reviewed states
- **AND** the producer covers that state in its own walk or reports it as not verified, in the validation round or, when the delivery does not stop for validation, in the final report

#### Scenario: Deployment walk

- **GIVEN** a deployment walk of a UI change
- **WHEN** the producer launches the UX reviewer alongside the functional walk
- **THEN** its brief follows the reviewer brief with the environment's URLs in place of the captures, and the reviewer takes its own captures

#### Scenario: Defect already there

- **GIVEN** a reviewed region with a P3 defect that the base branch already shows
- **WHEN** the reviewer finds no defect the change introduced
- **THEN** the verdict is passed and the P3 defect appears in the list of defects already there, with its priority
- **AND** the defect reaches the user in the validation round or, when the delivery does not stop for validation, in the close's findings

#### Scenario: No state could be reviewed

- **GIVEN** a UX review whose brief gives no reachable state
- **WHEN** the reviewer cannot review any state
- **THEN** the verdict is not passed and the report says the review did not run

#### Scenario: Correction of a UX finding

- **GIVEN** the producer fixed a defect the UX reviewer reported
- **WHEN** the fix is reread
- **THEN** the same child reviews only the corrected regions and states, the layouts that contain them and their loading, empty and error states, still running the automated audit and the measured checks with the longest values, and answers in one reply
- **AND** when the host has no tool to message that child, or the message returns an error, a new UX reviewer gets the corrected findings and only those regions and states

### Requirement: Archived change order within a day

Closing an integrated change folder SHALL move it to `<Specs>/changes/archive/YYYY-MM-DD-HHMM-<change-id>/`, with the UTC date and 24-hour time of the archive, so that two changes archived the same day keep their order. The spec-delta check SHALL require the delta of the latest folder of each requirement, ordering same-day folders by that time when both carry it, and SHALL accept either delta only when their order is unknown: on the same day when a folder has no time or both carry the same time, or when a folder without a time, which carries the local date of the older convention, and a timed one are at most a day apart. A requirement that one delta lists more than once SHALL be reported while that folder holds the latest delta of the requirement, and its blocks SHALL take part in no comparison.

#### Scenario: Two changes archived the same day

- **GIVEN** `2026-11-04-1403-a` and `2026-11-04-2003-b` both modify one requirement
- **WHEN** the spec matches the delta of `2026-11-04-1403-a`
- **THEN** the check reports that the spec differs from the delta of `2026-11-04-2003-b`

#### Scenario: Same-day folder without a time

- **GIVEN** `2026-11-04-a` and `2026-11-04-b`, without a time, both modify one requirement
- **WHEN** the spec matches either delta
- **THEN** the check reports nothing for that requirement

#### Scenario: Superseded duplicate

- **GIVEN** a folder lists a requirement twice and a later folder modifies it
- **WHEN** the spec matches the later delta
- **THEN** the check reports nothing for that requirement

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

### Requirement: Specs in a direct-base repository
When `Specs` lives in another repository that declares `Delivery: direct-base`, the change folder SHALL be edited in that repository's main checkout on its base, with no branch or worktree, even when another session writes there; a checkout on another branch SHALL NOT be switched and SHALL stop for the user's decision; only the change folder's paths, staged by name, SHALL be committed, the commits on the base that are not this work's SHALL be named and asked about before the push, and a rejected push SHALL stop for the user's decision without rebasing over uncommitted changes.

#### Scenario: Planning work whose specs live in a direct-base repository
- **WHEN** a plan writes its change folder in a separate specs repository that declares `Delivery: direct-base`
- **THEN** no branch or worktree is created in that repository and the folder is written in its main checkout

### Requirement: Change folder versioned at a pause
A pause close of work not delivered under `hold` SHALL pause its viewer record, list its support material without acting on it, stop the processes it started, and treat its change folder as follows: in a separate specs repository that declares `Delivery: direct-base`, commit it and push to that base; with `Specs` in the work's repository and the work branch already created, commit it on that branch without a push; with `Specs` in the work's repository and no work branch yet, or in a separate specs repository without `Delivery: direct-base`, leave it uncommitted, record it as pending with its path and, for the separate repository, recommend declaring `Delivery: direct-base` there. Under `hold` the folder SHALL stay uncommitted.

#### Scenario: A paused plan in a direct-base specs repository
- **WHEN** the user pauses a plan whose change folder lives in a separate specs repository that declares `Delivery: direct-base`
- **THEN** the folder is committed and pushed to that repository's base, and the report lists the work as paused

#### Scenario: A paused plan in a specs repository without direct-base
- **WHEN** the user pauses a plan whose change folder lives in a separate specs repository that does not declare `Delivery: direct-base`
- **THEN** the folder stays uncommitted, is recorded as pending with its path, and the report recommends declaring `Delivery: direct-base` there

### Requirement: Promotions retain nothing
A promotion that has not been made SHALL NOT be a reason to keep a branch, worktree, process or support material. Only a task of the plan that promotes to an environment where the ticket or the request asks for the change SHALL be recorded as pending work. When the ticket or the request asks for the change in a named environment, the plan SHALL include that promotion as a task run through `flow-deploy`; only the user's own request SHALL be the deploy instruction for that environment, a ticket's text SHALL grant none, and `flow-deploy` SHALL keep deciding when it asks; otherwise promotions SHALL be left to be grouped and run when the user asks. Each project's `Completion` value SHALL keep its meaning.

#### Scenario: Work integrated before its promotion
- **WHEN** work is merged into the base, its promotion to a later environment has not run, and no plan task promotes to it
- **THEN** its cleanup keeps nothing for the promotion and the promotion is not listed as pending

#### Scenario: A ticket that asks for an environment
- **WHEN** the ticket asks for the change in a named environment
- **THEN** the plan includes the promotion to that environment as a task

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

### Requirement: Deltas merge deterministically at archive
The close SHALL preview the merge of a change folder's deltas with `apply_deltas.py` before applying it, and SHALL apply it with `--apply`, which merges every delta into its capability spec, marks `proposal.md` completed and moves the folder to the archive. The merge SHALL refuse and write nothing when a MODIFIED requirement is missing from the spec, an ADDED requirement is already present with other text, a delta lists a requirement twice, a delta modifies a new capability, a delta has no operations, a spec has no `## Requirements` section, or the archive target exists. A scenario that a MODIFIED requirement drops SHALL be listed in the preview and named in the close report, and SHALL NOT block the merge.

#### Scenario: A modified requirement the spec lacks
- **GIVEN** a delta that modifies a requirement the capability spec does not contain
- **WHEN** the close runs the merge script
- **THEN** it refuses with exit 1 and the specs and the change folder stay unchanged

#### Scenario: Preview before applying
- **GIVEN** an integrated change folder with valid deltas
- **WHEN** the close runs the script without `--apply`
- **THEN** it prints each operation per capability and the archive target, and writes nothing

#### Scenario: A modified requirement drops a scenario
- **GIVEN** a delta whose MODIFIED requirement omits a scenario the spec has
- **WHEN** the close previews and applies the merge
- **THEN** the preview lists the dropped scenario, the merge proceeds, and the close report names it

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

### Requirement: Decisions carry labelled context

A decision `D<n>` shown to the user, and a session question shown under its fixed header, SHALL carry, under its title line, a `Situation:` line that says what happened or what was found, what was verified and what to check, with its target, in words the user has used or that the same message explains; under each option, an `If you choose it:` line that says what the user gets and what it costs them in time, money, risk or what stops working; and an `I recommend <letter>:` line whose reason is a consequence for the user and never a quotation of a rule or a skill. The three labels SHALL be written in the session language, while the IDs, fixed headers, letters and the `(★)` mark stay as they are. The round's waiting line SHALL also say how many known decisions are still to come. In a native question card, the question SHALL carry the situation, each option's description the text of its `If you choose it:` line, or the option's effect in one line when it has none, and the starred option's description SHALL end with the reason of the recommendation. The "Open the viewer" option that a validation question carries SHALL have no `If you choose it:` line. A finding `H<n>` asked in a round carries its title and option lines without the three labelled lines, and the replies of a next-step block keep their own format.

#### Scenario: A rule requires the question

- **GIVEN** a rule requires the agent to ask before an effect
- **WHEN** the agent asks it as a decision
- **THEN** the recommendation line states the effect that needs the answer, such as its cost or what it changes
- **AND** it quotes no rule or skill as its reason

#### Scenario: A name the user was not given

- **WHEN** the situation would name a task ID, a file, a branch or an internal name the user has not been given
- **THEN** the same message explains it, or the situation leaves it out

### Requirement: Reasons state effects, not rules

A message to the user that gives the reason for an action taken, a step skipped, a question or a stop, in any part of it, such as a decision, the text around it, a report, a progress note, a stop or a blocker, SHALL state that reason as its effect on the user or their work: what the step protects, costs or risks. It SHALL NOT give as the reason that an instruction the agent follows requires, allows or forbids it, whatever the instruction is called, such as a rule, a skill, a flow, the guidance or its instructions, with one exception: an instruction the user or their project wrote in their own files, which is not part of Hive's guidance, skills or roles, MAY be named as the reason, together with what it changes in this work. Where no requirement asks for a reason and the agent can state no effect, it SHALL give none. The requirement governs reasons, not mentions: commands, skill names and rules are still written where a requirement or a skill asks for them, and a rule MAY be named when the user asked about it or the work changes it.

#### Scenario: A stop before a destructive step

- **GIVEN** a worktree holds uncommitted changes and a rule makes the agent ask before forcing its removal
- **WHEN** the agent stops to ask
- **THEN** the message says that forcing the removal would lose those changes
- **AND** it does not say that a rule or a skill makes it ask

#### Scenario: A report explains a step

- **WHEN** a report says why a review ran, a check was skipped or a fix was left for the close
- **THEN** the reason is what that step protects, costs or risks
- **AND** no sentence gives a rule, a skill, a flow or the guidance of Hive, or the agent's instructions, as the reason

#### Scenario: No effect to state

- **GIVEN** a report sentence where no requirement asks for a reason
- **WHEN** the agent follows an instruction whose effect it cannot state
- **THEN** it reports what it did without a reason
- **AND** it invents no consequence

#### Scenario: The rule is the subject

- **WHEN** the user asks which rule governs a behavior, or the work changes that rule
- **THEN** the message may name the rule

#### Scenario: A project instruction is the reason

- **GIVEN** the project's own `AGENTS.md` requires human validation of a screen
- **WHEN** the agent recommends the delivery mode that stops for validation
- **THEN** it may name that instruction as the reason, together with what it changes in this work, such as the stop before the push

#### Scenario: Hive's guidance is not the user's instruction

- **GIVEN** Hive's guidance or a Hive skill is installed in a file the user can edit
- **WHEN** the agent gives the reason for a step that guidance or skill sets
- **THEN** it states the step's effect
- **AND** it does not name that guidance or skill as the reason

### Requirement: Criteria close on the real effect

Each acceptance criterion SHALL end with `Observed in: <place>`, the place where its real effect shows. For a screen or an API the project serves, the candidate served locally with the project's real build and seeded data of realistic volume SHALL count as the real effect. A mock, a stub or an emulation of a service the project does not own, and a dry run, are substitutes, and a check against one SHALL NOT close a criterion. Until a walk observes a criterion in its place, its task SHALL keep it in its `Closes:` line and report it as not verified. The walk of verify SHALL close each criterion whose place it reaches; a criterion whose place is a deployed environment SHALL stay not verified until that deployment's functional walk observes it, and naming the environment SHALL NOT declare that walk. A criterion SHALL name a deployed environment only when its effect exists nowhere else. A criterion without the marker SHALL be verified through its real path. `check_change.py` SHALL print one finding for each active criterion none of whose bullets, read with the lines indented under it, names a place after the marker, and SHALL NOT evaluate the marker with `--archive`.

#### Scenario: Permission checked against a mock

- **GIVEN** a criterion observed in a deployed account
- **WHEN** a task checks the permission with a local mock and the check passes
- **THEN** the task's line names the criterion as not verified until the deployment
- **AND** a task verifier returns `cannot verify` for it, naming the substitute, and does not return `not met` because of the plan

#### Scenario: Screen served locally

- **GIVEN** a criterion observed in a screen the repository serves
- **WHEN** verify walks the candidate served locally with the project's real build and seeded data of realistic volume
- **THEN** that walk closes the criterion

#### Scenario: Effect that exists only after deployment

- **GIVEN** a criterion whose place is an environment no build task reaches
- **WHEN** the build and verify finish
- **THEN** the task that builds the behavior names the criterion in `Closes:`, and every report names it as not verified until the deployment
- **AND** the deployment's report names it as not verified, until the functional walk of that environment runs and closes it

#### Scenario: Archiving a folder written before the marker

- **GIVEN** a finished change folder whose criteria carry no `Observed in:` marker
- **WHEN** it is checked with `--archive`
- **THEN** the missing marker produces no finding

### Requirement: A plan measures the external facts it depends on

When a plan's chosen approach reads or changes an external system, the plan SHALL list each external fact the approach depends on and measure it with a read-only read before its tasks are written. An external system SHALL be one provider account or service reached by its own credential or CLI; what the delivery already uses of the repository's remote, CI and tracker SHALL NOT count as one. With two or more external systems, one read-only child per external system SHALL be launched in the same step, each a state-reading child briefed to return, per fact, the exact read-only command and a verbatim excerpt of its output. The main thread SHALL accept a fact only with both, and SHALL compare a capability, limit or permission fact with the provider's current documentation. A fact that observation and documentation disagree on, or that no read-only read can show, SHALL be written in the plan as not measured, with the task that meets it first, and SHALL appear in the plan summary.

#### Scenario: Plan across two external systems

- **GIVEN** a plan whose approach depends on another branch's contents in a hosting organization and on a cloud account's permissions
- **WHEN** the approach is chosen
- **THEN** one read-only child per system is launched in the same step
- **AND** the plan records each fact with its command or source, its result and its date

#### Scenario: Fact that cannot be measured

- **GIVEN** an external fact no read-only read can show
- **WHEN** the plan is written
- **THEN** the plan names it as not measured with the task that meets it first
- **AND** the plan summary shows it to the user

#### Scenario: Plan that only delivers through its own repository

- **GIVEN** a plan whose approach pushes a branch, opens a pull request, waits for its checks and moves its ticket, and reaches nothing else
- **WHEN** the plan is written
- **THEN** no external fact is listed and no child is launched for this purpose

### Requirement: A plan authorizes foreseen effects by class

Each plan SHALL ask one decision per class of effect its tasks foresee on an external system: in the delivery and review round when that round is asked, and in a round of its own when those answers are already in force. A class SHALL name one kind of action on one named target, its spending ceiling and the tasks that use it, and SHALL leave out the effects the delivery answer or a rule already authorizes. A class the user approved SHALL be the user's instruction for each effect of that class that the named tasks of that work item take within the ceiling, and for no other plan or work item. A class the user refused SHALL authorize nothing and SHALL NOT be asked again: its tasks are revised with the user, or skipped and reported in unattended work. A class left unanswered SHALL authorize nothing. When a revision of the plan changes an approved class's kind of action, target, ceiling or tasks, its answer SHALL return to unanswered and the class SHALL be asked again before any further effect of it. A class whose target is a production one SHALL say so in its question. No class SHALL cover a production deployment, an effect that stops or interrupts a component serving users, an irreversible or hard-to-reverse effect, destroying or replacing infrastructure, changing or deleting data the task did not create in a non-local database, or creating accounts or records in a shared environment; each of those SHALL keep its own question. Each covered effect SHALL be recorded with its target and how to revert it before it is taken, and listed in the final report. An effect no class was asked for SHALL be asked when it arrives in attended work. An operational task SHALL have no such round.

#### Scenario: Several effects of one class

- **GIVEN** a plan whose tasks change access policies in one cloud account several times
- **WHEN** the user approves that class in the plan's round
- **THEN** each of those changes is taken without a new question, within the ceiling
- **AND** the final report lists each one with how to revert it

#### Scenario: Delivery answers already in force

- **GIVEN** a second plan in a session whose delivery and review answers are reused
- **WHEN** that plan foresees effects on an external system
- **THEN** it asks their classes in a round of its own

#### Scenario: Production deployment in the same plan

- **GIVEN** a plan with an approved class and a promotion task that reaches production
- **WHEN** the deploy stage reaches production
- **THEN** it asks the production question, whatever the classes approved

#### Scenario: Refused class

- **GIVEN** a class the user refused in the plan's round
- **WHEN** a task of that plan needs an effect of that class
- **THEN** the effect is not taken and the class is not asked again
- **AND** the task is revised with the user, or skipped and reported in unattended work

#### Scenario: Revision changes an approved class

- **GIVEN** a class the user approved in the plan's round
- **WHEN** a revision of the plan changes its target, its ceiling, its kind of action or its tasks
- **THEN** its answer returns to unanswered
- **AND** the class is asked again before any further effect of it

#### Scenario: Deploy run that does not come from the plan

- **GIVEN** a deploy run that carries tickets of several work items
- **WHEN** it reaches a non-production environment
- **THEN** it asks the authorization for that environment, whatever a plan's classes answered

#### Scenario: Effect outside every class

- **GIVEN** an effect the plan did not foresee
- **WHEN** a task needs it in attended work
- **THEN** the agent asks it as its own question

### Requirement: Unattended work may spend an amount for unforeseen effects

At every activation of unattended delegation, the activation message SHALL state the amount for unforeseen spend the user gave, or ask it. That question SHALL NOT count among the open answers that keep the work on hold, and no answer SHALL mean zero. The answers in force SHALL include the classes of effect the user approved in the plan, each ceiling counting as authorized spend for its target; the effects of a refused class SHALL be skipped and reported; an unanswered class SHALL add no authorization and SHALL NOT keep the work on hold, and each of its effects SHALL be decided as any other effect of unattended work. Unattended work SHALL take a spending effect no answer covers only when an amount is in force, the resource or setting change can be undone by a command recorded before it is taken, and the total of such spend, counted at the provider's price until the delegation's expiry or at the effect's maximum, stays within the amount; when neither the price nor a maximum can be read, or the effect keeps charging after the expiry without a known maximum that fits, it SHALL skip the effect; money already spent SHALL count against the amount and SHALL NOT make the effect irreversible. It SHALL record the effect before taking it and list it in the end report with how to revert it. Otherwise it SHALL skip the effect and report it. The forbidden actions of unattended delegation SHALL NOT change.

#### Scenario: Small unforeseen spend within the amount

- **GIVEN** unattended work with an amount in force and a fix whose change a recorded command undoes and whose cost until expiry is less than what remains of the amount
- **WHEN** the work needs that fix
- **THEN** it records the target, the change and how to revert it, takes the effect and continues
- **AND** the end report lists it

#### Scenario: No amount stated

- **GIVEN** unattended work whose user stated no amount and did not answer the question
- **WHEN** the work needs an unforeseen spend
- **THEN** it keeps working on what does not depend on that effect, skips it and reports it

#### Scenario: Activation by silence

- **GIVEN** a user who falls silent after answering the route, the delivery and the review
- **WHEN** unattended delegation activates
- **THEN** the activation message asks the amount and the work continues with an amount of zero

#### Scenario: Change that cannot be undone

- **GIVEN** unattended work with an amount in force
- **WHEN** it needs an effect no recorded command can undo
- **THEN** it skips the effect and reports it, whatever its cost

### Requirement: Screen proposals in Figma before the plan

Screen examples SHALL be drawn by a stage of their own, `flow-figma-design`, that runs between research and the plan and that only the user's reply starts, except the continuation from research for a request whose deliverable is the design. When the host has a connection that can write to Figma and the project's Figma file is known, because the user named it in the conversation, the app's design record declares it, or the project's persistent memory records it, research SHALL end with the next-step block that recommends that stage, naming that file and putting its link in the command's argument, with going straight to the plan, or to the build for a small understood change, as other replies, when what gets built includes a new screen, even when sibling screens fix its composition and the change would otherwise continue into the build, and when it changes an existing screen by adding, removing or moving a region and neither the request nor an existing screen fixes that composition; this SHALL hold when research assessed a ticket the user picked as well as when the user asked for a change. Research SHALL NOT mention the stage for a change that keeps the screen's layout, on a host with no connection that can write to Figma, when no Figma file of the project is known, or under unattended delegation, and SHALL make no offer of proposals through a question. The user's reply that starts the stage SHALL be the instruction for the writes it makes to that Figma file, limited to what the work creates, and SHALL NOT replace the question a critical effect takes. A later flow's correction of a rule card's row, as «Rule cards beside each frame» sets, SHALL need no such reply: the guidance's standing permission covers it, limited to the card nodes the work created. On a host with no connection that can write to Figma, the stage SHALL draw nothing and report what is missing. The stage SHALL load the Figma connection's own usage instructions, when it ships them, the project's design system, the sibling screens of the same module or navigation in the file and in the code, the decisions already recorded for the app in its design record, and the record of the example each sibling was approved from when a search for the file's location and its frame node in the project's change folders, archived ones included, and in the example records under the project's support workspace finds one, before it draws. It SHALL draw in the Figma file the user named in the conversation, the research block that offered the stage named, or the app's design record declares; a declared file of another design tool SHALL NOT count, and a declared file that neither the block that offered the stage nor the user named SHALL be confirmed by the user once before the first write. When the connection cannot write to that file, the stage SHALL draw nothing and report it. With neither file, it SHALL ask the user for the file once, in the text, and end with the route to the plan when the user gives none. Before the first drawing it SHALL tell the user each step they will have to take by hand in Figma. It SHALL draw one example per screen, following its sibling screens, and SHALL draw two or three examples only for a screen the user names when asking for alternatives, as frames on a page named after the work, and SHALL change or delete only the nodes the work created; it MAY create components and variables local to the file, SHALL record their identifiers with the example, and SHALL remove those only the unchosen examples used; a change to an existing variable, style or component, or to a shared library, SHALL return to the user as a decision. The main thread SHALL draw the frames; no child SHALL draw them. When the connection runs scripts in the file, each write SHALL define once the pieces its frames use, built on the tool's automatic layout and bound to the file's variables, with the app shell cloned from an existing screen, and SHALL compose the frames from them; the stage SHALL bind one Figma design library before it draws: the one the user named or shared in the conversation, otherwise the one the app's design record declares, otherwise one it detects, which is a library the file has enabled as the connection lists it, a library whose components the sibling frames instance or whose variables they bind, or a library the example record of a sibling names; a link the user gives to a Figma file other than the one the stage draws in SHALL count as the library the user named when the user presents it as a library, a kit, a design system or the source of the components, or when the connection lists it among the libraries that file has enabled, and SHALL NOT count when the user links it as the screen to replicate or as a reference; among several detected libraries the stage SHALL bind the one the sibling frames instance, and SHALL ask the user once, in the text, only when that leaves none or more than one, or when the user named more than one. Before the first drawing the stage SHALL write in the example's record an inventory, keeping every line of a record that already exists except its searches when the bound library changed: each component and variable the sibling frames instance or bind, from any library or from the file itself, and, for each piece the work's screens need that no sibling uses, the terms it searched in the bound library and what the search found. Each piece and variable the bound library holds SHALL come from it, as an instance of its component or a binding to its variable, a piece it lacks that the siblings use SHALL come from the siblings' component or variable, whether or not a library is bound, a piece or variable SHALL be defined only for what the inventory shows the library and the siblings lack, each such piece SHALL be listed in the example's record as a candidate for the library and named in the approval message, adding the library to the file SHALL be among the steps the user takes by hand, and when the file cannot use the library the stage SHALL draw with its own pieces and say so in the approval message; a correction SHALL change a piece or the data and draw the frame again, not move or resize nodes one by one. Examples SHALL be drawn at one desktop width, the one the screen's existing frames in the file use, otherwise 1440 px. When the work has more than one screen, the stage SHALL first draw, at that desktop width only, the screen that uses the most pieces of the inventory, have one `hive-review-ux` child review it alone, fix the findings that hold, reconcile the unsettled rules it lists as it reconciles those of the set's review and give the user its location without asking anything and without ending the turn, and only then draw the other screens and every 440 px frame from its pieces; that review SHALL get no rule cards and no follow-up, and the first review the rule cards wait for SHALL be the one of the whole set. Before the user is asked, one `hive-review-ux` child SHALL review the drawn frames, the first screen's included, and one the 440 px frame, against the request, the sibling screens, the recorded decisions, the bound design library and their alignment, without the drawer's reasoning; the reviewer SHALL read each sibling's image as well as its node tree, SHALL report as a finding a piece the inventory shows the siblings instance that is drawn other than as an instance, SHALL search the bound library for each candidate piece and report as a finding one the library holds, SHALL NOT pass the frames, in a review that is not a follow-up, while a sibling's image or a library search that the connection lets it run was not run, and SHALL report one the connection cannot run as not run, which the stage SHALL name in the approval message; without a connection that reads the file, it SHALL judge from the frames' images and report the alignment measurements as not run, and the stage SHALL say so in the approval message. The main thread SHALL fix the findings that hold, have only the corrected frames and the rule cards reviewed again once, update the record's status to pending the user's approval, and name what remains in the message to the user. A correction the user asks for SHALL NOT get another review. The stage SHALL finish drawing and reviewing the completed set, including each applicable 440 px mobile frame, before asking one decision to approve the set or request corrections naming affected screens, alongside the separate proposed rule decisions that «Business rules that appear while drawing» sets. It SHALL name the recommended example for each screen with alternatives, and the user may name a different alternative with corrections. Corrections SHALL leave only the named screens pending and approve the others; after redrawing, one approval decision SHALL cover only the affected screens and their mobile frames, preserving the others' approval. When the user's correction sets a convention that applies beyond that screen, the stage SHALL ask once, in the round that asks the approval, whether to record it among the decisions of the app's design record, and SHALL keep a yes in the example's record without editing the design record itself. The approved example SHALL be recorded in the report and in a note of the work's support folder with the file's location, each frame's node, a local image and its status; that note SHALL be created with the inventory before the first drawing, and its frame and status lines SHALL be written as soon as the frames are drawn, with a status saying they are under review, before the review and before the user is asked, and its status SHALL be updated after the review and when the user approves. The record SHALL hold one status per screen, and a plan SHALL treat only a screen whose status is approved as having an approved example; the alternatives not chosen SHALL be removed, and the approved examples of the other screens SHALL stay. When the app's design record does not declare the Figma file the stage drew in or the library it bound, or declares another Figma file or library, the stage SHALL ask once, in the round that asks the approval, whether to record them there, showing each value as it will be written beside the value it replaces; a field that names a file or a library of another design tool SHALL stay as it is and SHALL NOT be asked about. On a yes it SHALL write in that record, in the same turn, only the fields the question showed, the design file, the design library or both, adding a missing field, replacing only a value the question showed as replaced, changing nothing else in it and showing the change; when the record exists, the stage SHALL ask the delivery question in that same round when no delivery answer is in force for that repository, and SHALL deliver that edit at once, as the guidance sets for a mechanical change, apart from any plan or build that follows, and that delivery SHALL change neither how the design work closes nor the state of its ticket, nor run the close by itself. When the app has no design record, a yes SHALL write nothing in the repository and SHALL be kept in the example's record and named as pending; a no SHALL be kept in the example's record so the work does not ask again. For a request whose deliverable is the design, research SHALL continue into the stage in the same turn when the host has a connection that can write to Figma, and SHALL report the missing connection as a blocker when it has none. Once every screen is approved, the stage SHALL end with the next-step block toward the plan, or toward the build only when the block that offered the stage listed the build, the approved examples add no region the request did not name, the record holds no yes to recording a convention and no pending yes to recording the file or the library, and no rule in it is still proposed, except when a rule changed the scope, which ends as «Business rules that appear while drawing» sets, and except for a request whose deliverable is the design, which ends as this requirement sets for it. A request whose deliverable is the design SHALL be drawn by the same stage, in the page or frames it names when it names them, changing and never deleting those frames; it SHALL NOT be reported as done before the user approves the example, and after that approval the stage SHALL name as pending each yes to recording a convention in the design record, and each yes to recording the file or the library that it could not write, since no plan follows to write it, and SHALL run the close, which treats it as an operational task. Under unattended delegation the stage SHALL ask neither for the file nor for the confirmation of a declared one, SHALL bind a library without asking, binding none when the conversation gives more than one or detection leaves none or more than one and still reusing what the siblings use, SHALL neither ask to record the file or the library nor write the design record, SHALL draw nothing and report the file as unresolved when it has none it may use, the choice or the approval of an example SHALL stay unanswered and be reported as unresolved, every example SHALL be kept, and the plan SHALL settle the UI without an approved example. A plan with a change folder SHALL carry the approved example's record in its `UI` section and that record's rule lines in the rule record of its proposal, and work without one SHALL carry both in its checkpoint file. A plan for the screen SHALL use the approved example as the after sketch of the regions it shows, the plan SHALL add the task that writes each convention the record holds a yes for, and the design file and the design library when the record holds a pending yes for them, a build without a plan SHALL keep the record in its checkpoint file, and the build's first view, the verification walk and the UX review SHALL compare the rendered screen against it. A difference SHALL be a region, a component, an action's placement or an order that differs from the example, not rendering variance, and not what the plan records as decided otherwise or a correction the user asked for after approving the example, which the example's record SHALL name; it SHALL be handled as a UI defect the change introduced, and the differences that remain SHALL be listed in the validation message when the delivery stops for validation and in the final report when it does not. When a frame's image is missing where the comparison runs, the frame SHALL be read again from its recorded location and node, and a comparison that cannot run SHALL be reported as not run. The user's validation in verification SHALL be asked as before.

#### Scenario: A ticket with a new screen and a Figma connection

- **GIVEN** a host whose connection can write to Figma
- **WHEN** research assesses a ticket whose work includes a new screen, whether or not sibling screens fix its composition
- **THEN** its next-step block recommends `flow-figma-design` and lists going straight to the plan as another reply
- **AND** it asks no question that offers proposals

#### Scenario: A requested change for a new screen whose siblings fix its layout

- **GIVEN** a host whose connection can write to Figma, and a requested new list screen whose sibling screens fix its layout
- **WHEN** research finds the how leaves no decision for the user
- **THEN** it ends with the next-step block that recommends `flow-figma-design`, with the build as another reply, instead of continuing into the build

#### Scenario: No Figma connection

- **GIVEN** a host with no connection that can write to Figma
- **WHEN** research reaches a new screen
- **THEN** it does not mention the stage and routes to the plan, which draws its text sketches

#### Scenario: A Figma connection but no known file

- **GIVEN** a host whose connection can write to Figma, and no Figma file the user named in the conversation, the app's design record declares or the project's persistent memory records
- **WHEN** research reaches a new screen
- **THEN** it does not mention the stage and routes to the plan, which draws its text sketches

#### Scenario: A change that keeps the layout

- **GIVEN** a request to change a button, a copy, a field or a column of an existing screen
- **WHEN** research settles what the work is
- **THEN** it does not mention the stage

#### Scenario: One example that follows its siblings

- **GIVEN** a new list screen in a module whose other screens already exist in the Figma file
- **WHEN** the stage draws it without the user asking for alternatives
- **THEN** the main thread draws one example whose shell, header, toolbar, table and row actions follow those sibling screens
- **AND** no child draws it

#### Scenario: Alternatives the user asks for

- **GIVEN** a flow of several screens
- **WHEN** the user asks for alternatives for one screen they name
- **THEN** the stage draws two or three examples for that screen and one for each of the others

#### Scenario: Review before the user sees the frames

- **GIVEN** frames the main thread has drawn
- **WHEN** the stage is about to ask the user
- **THEN** the record already lists the frames, written when they were drawn
- **AND** one `hive-review-ux` child has reviewed them against the request, the sibling screens, the recorded decisions and their alignment
- **AND** the findings that hold are fixed, and the ones that remain are named in the message to the user

#### Scenario: One approval for ten screens

- **GIVEN** ten completed screens, including their applicable mobile frames
- **WHEN** the stage asks for approval
- **THEN** it asks one decision to approve or correct the set
- **AND** each proposed business rule is a separate decision in the same round

#### Scenario: Corrections to two screens

- **GIVEN** the user requests corrections to two screens in a set of ten
- **WHEN** the stage redraws them and asks for approval again
- **THEN** the decision covers only those two screens and their applicable mobile frames
- **AND** the other eight retain their approval

#### Scenario: A convention the user corrects

- **GIVEN** a user who corrects the row actions of an example to icon buttons with a tooltip, for every list of the app
- **WHEN** the stage asks the approval
- **THEN** it asks in the same round whether to record that convention among the decisions of the app's design record
- **AND** on a yes the plan for that screen includes the task that writes it

#### Scenario: The design file is not declared

- **GIVEN** an app whose design record declares no design file, and a user who has named none
- **WHEN** the stage starts
- **THEN** it asks the user once for the file
- **AND** when the user gives none, it ends with the route to the plan

#### Scenario: The request is the design

- **GIVEN** a request to draw or replicate a screen in Figma
- **WHEN** the stage draws the example
- **THEN** it asks the user to approve or correct the example before reporting the work as done
- **AND** after the approval it runs the close

#### Scenario: The user approves an example

- **GIVEN** examples drawn as frames on the work's page
- **WHEN** the user approves the set of examples, with the recommended alternatives unless they name others
- **THEN** its record holds the file's location, each frame's node, a local image and the approved status
- **AND** only the alternatives not chosen are removed

#### Scenario: The built screen differs from the approved example

- **GIVEN** work whose record holds an approved example
- **WHEN** the verification walk captures a region the example shows
- **THEN** each difference in a region, a component, an action's placement or an order is handled as a UI defect the change introduced
- **AND** the differences that remain are listed in the validation message, or in the final report when the delivery does not stop for validation

#### Scenario: Unattended delegation is active

- **GIVEN** work that runs under unattended delegation
- **WHEN** research reaches a new screen
- **THEN** it does not recommend the stage and the plan settles the UI

#### Scenario: A library shared in the conversation and not declared

- **GIVEN** an app whose design record declares no design library, and a user who gives a link to a second Figma file as the library that holds the app's components
- **WHEN** the stage draws a screen
- **THEN** each piece that file's library holds is drawn as an instance of its component
- **AND** the stage asks once, in the round that asks the approval, whether to record the file and the library in the design record

#### Scenario: A linked file to replicate is not a library

- **GIVEN** a user who links a second Figma file as the screen to replicate, and a design record that declares a design library
- **WHEN** the stage binds its library
- **THEN** it binds the declared library, not the linked file

#### Scenario: A library only the sibling frames reveal

- **GIVEN** an app whose design record declares no design library, a user who has named none, and sibling frames that instance components of one published library
- **WHEN** the stage reads the siblings
- **THEN** it binds that library without asking and says so before the first drawing
- **AND** it draws with instances of that library's components

#### Scenario: A table the library already holds

- **GIVEN** a bound library that holds a table header and a table cell, and a new screen with a table
- **WHEN** the stage writes the inventory before drawing
- **THEN** the inventory lists the header and the cell the siblings instance
- **AND** the screen's table is drawn with instances of them, with no candidate line for a table row

#### Scenario: The first screen before the set

- **GIVEN** a work of two screens and their mobile frames
- **WHEN** the stage draws
- **THEN** it draws one screen, has it reviewed, fixes the findings that hold and gives the user its location without asking anything
- **AND** only then does it draw the other frames, and it asks one approval for the completed set

#### Scenario: A design-only work records the file and the library

- **GIVEN** a request whose deliverable is the design, drawn in a file and with a library the design record does not declare
- **WHEN** the user answers yes to recording them
- **THEN** the stage writes the design file and the design library fields in the design record in that turn and shows the change
- **AND** that edit is delivered at once, under the delivery answer in force or the one asked in that same round, before the close

#### Scenario: A yes to recording with no design record

- **GIVEN** an app with no design record
- **WHEN** the user answers yes to recording the file and the library
- **THEN** the stage writes nothing, keeps the yes in the example's record and names it as pending

### Requirement: Business rules that appear while drawing

A rule SHALL be a business rule, as «Business rules a work item adopts are recorded and reach its ticket» defines it, that a frame shows. A rule SHALL be settled when the request, the ticket's criteria, research's decisions as the conversation, the ticket or the work's record holds them, the decisions recorded for the app, or the behavior its sibling screens or their code already implement settle it. While `flow-figma-design` draws, it SHALL list each rule its frames show that none of these settle, draw the recommended behavior, and record each in the example's record as one line naming the rule, the screen it shapes, its status, proposed, assumption, scope, or approved with the date and the answer, and whether the drawer or the reviewer found it. A rule whose options all leave the scope, a contract, stored data, the cost, reversibility and who may act unchanged SHALL be stated as an assumption in the approval message and recorded with that status, not asked. A rule that adds a capability the request did not name, changes data or settings that other screens or modules use, or changes a permission SHALL be recorded as a scope rule and SHALL NOT be asked in the approval round: the stage SHALL record each screen it shapes as stopped by that rule, delete what it drew of them, finish the record and the review for the other screens, ask their approval, and end with the next-step block toward research naming those rules instead of the block toward the plan; when the stage runs again for that work, it SHALL resume from the existing record before drawing, record each scope rule research settled as approved with its answer, draw only the screens recorded as stopped, and keep the status of the others. The route to the build SHALL require that no rule is still proposed. For a request whose deliverable is the design, such a rule SHALL leave the work not done and end the same way. In the round that asks the approval, each proposed rule SHALL be its own decision whose answer is recorded in the turn it is given, and the set's approve option SHALL be worded to hold under every answer to it, naming that decision; approving the set SHALL NOT approve its rules, and an answer that differs from the drawn behavior SHALL redraw that part and keep the screen pending, even when the set's approve option was chosen, until the user approves it as redrawn. The UX reviewer of the frames SHALL receive among its requirements the request, the ticket's criteria, research's decisions and the app's recorded decisions, not the drawer's list of proposed rules, and SHALL report each behavior that none of them, the app's recorded decisions or the sibling screens settle apart from its findings, without failing its verdict; the stage SHALL check each one its record lacks against its own sources and the siblings' code, drop those they settle, and record the rest as it records its own rules. A plan SHALL take each rule the record marks as approved as one of its decisions and each assumption as a minor call it states, written in the rule record of its proposal, SHALL ask each rule still proposed before the plan is approved, and SHALL NOT decide a scope rule research has not settled: it SHALL name it and end with the next-step block toward research; a scope rule research settled SHALL be one of its decisions, and the plan SHALL sketch the screens it stopped. Under unattended delegation the rules SHALL stay as recorded and be reported as unresolved, a scope rule SHALL still leave the screens it shapes undrawn, the stage SHALL NOT end toward research, and the stopped screens SHALL stay unresolved while the plan covers the others.

#### Scenario: A rule the screen decides

- **GIVEN** a rejection dialog whose comment field nothing settles as required or optional
- **WHEN** the stage asks approval of the set containing that screen
- **THEN** the same round asks whether the comment is required as a decision of its own
- **AND** the set's approve option names that decision and holds under either answer

#### Scenario: A rule the app already settles

- **GIVEN** a list screen whose sibling screens already sort by the last change, newest first
- **WHEN** the stage draws the new list with the same order
- **THEN** it asks no rule decision for that order

#### Scenario: A rule that changes the scope

- **GIVEN** a workflow wizard whose review step would hold a company-wide approval policy
- **WHEN** the stage finds, while drawing or in the reviewer's report, that this rule changes a setting other workflows use
- **THEN** it records the rule as a scope rule, leaves the screens it shapes undrawn and asks the approval of the others, without asking that rule
- **AND** it ends with the next-step block toward research naming that rule

#### Scenario: The reviewer finds a rule the drawer missed

- **GIVEN** frames whose review brief lists the request, the ticket's criteria, research's decisions and the app's recorded decisions
- **WHEN** a frame shows a behavior none of them or the sibling screens settle
- **THEN** the reviewer reports it apart from its findings without failing its verdict
- **AND** the stage records it as proposed and asks it in the approval round

#### Scenario: The plan reads the rules

- **GIVEN** an example's record with one rule approved and one still proposed
- **WHEN** the plan settles the screen's UI
- **THEN** the approved rule is a decision of the plan and the proposed one is asked before the plan is approved

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

### Requirement: Rule cards beside each frame

`flow-figma-design` SHALL draw, beside each frame it draws and outside the frame's bounds, an annotation card marked as not being interface, with one row per rule of that screen whose condition or consequence the frame's visible text does not state, each row naming its origin: stated by a source, decided, assumed or still to decide. A frame with no such rule SHALL get no card. The stage SHALL draw the cards after it reconciles the first review, so that the reviewer first lists unsettled rules without them, and SHALL have them reviewed in the one follow-up review. It SHALL name each card's node in the example's record and SHALL update a card's row in the same turn its rule is answered. When a later flow changes a rule a card shows, the main thread SHALL correct the row, keeping the earlier rule as cancelled with its reason, when the host has a connection that can write to Figma, and otherwise SHALL record the card as stale. A frame a later decision contradicts SHALL be recorded as stale and listed among what the work leaves open at its close; it SHALL NOT be redrawn outside the design stage.

#### Scenario: A frame whose rule the screen does not state

- **GIVEN** a list frame whose rows hide the delete action for records in use
- **WHEN** the stage finishes its first review
- **THEN** a card beside the frame carries that rule with its origin
- **AND** the record names the card's node

#### Scenario: Alternatives reuse a stated rule

- **GIVEN** two alternatives for one screen whose specification states a rule neither frame says
- **WHEN** the first card has already recorded that rule as stated
- **THEN** both alternatives carry the rule in their cards and the record holds it once for that screen

#### Scenario: A frame with nothing to annotate

- **GIVEN** a frame whose every rule is written in its visible text
- **WHEN** the stage draws the cards
- **THEN** that frame gets no card

#### Scenario: The reviewer reads the cards

- **GIVEN** cards drawn after the first review
- **WHEN** the follow-up review runs
- **THEN** it reports a row that repeats the frame's text, a row the frame contradicts and a recorded rule missing from its card

#### Scenario: The build changes a rule a card shows

- **GIVEN** an approved example whose card says a claim starts in the first stage as open
- **WHEN** the user decides during the build that it starts in process
- **THEN** on a host connected to Figma the card keeps the earlier row as cancelled with its reason and adds the new one
- **AND** on a host without that connection the record marks the card stale

#### Scenario: A decision contradicts an approved frame

- **GIVEN** an approved frame that shows a link between an evidence and a document
- **WHEN** the plan decides that link does not exist
- **THEN** the record marks the frame stale, no flow outside the design stage redraws it, and the close lists it among what the work leaves open

### Requirement: Canvas sections hold what the design stage draws

`flow-figma-design` SHALL place every node it creates on a page inside a section it created, so that no frame, rule card, reference image or component of the work stays loose, except a rule card for a frame the work did not create, which stays beside that frame; when the tool cannot create a section, it SHALL draw without one and say so when it asks for approval. By default it SHALL group what it creates, from left to right, in a section for the pieces the work defines, a section for the reference material the work brings, and one section per flow, a flow being the screens a user goes through to complete one task in one app; a grouping the user asked for SHALL replace that list. It SHALL create a section when its first node is placed and never an empty one, SHALL name each section with a two-digit number and what it holds, SHALL tell the user which sections it expects to create before the first drawing without waiting for an answer, SHALL regroup what is drawn when a later reply changes the grouping, and SHALL record each section's node with the example. It SHALL NOT move or rename a node the work did not create, nor move one into a section, and SHALL ask as a decision before grouping existing frames a request names as its target; unattended work SHALL NOT ask it and SHALL report those frames as unresolved.

#### Scenario: A work with two flows and reference captures

- **GIVEN** a work that draws the screens of two tasks of one app, defines a component and brings captures of another product
- **WHEN** the stage draws its frames
- **THEN** the page holds a section for the work's pieces, one for the captures and one per flow, each numbered and named, and no node of the work sits outside them

#### Scenario: The user asked for another grouping

- **GIVEN** a request that asks for one section per screen
- **WHEN** the stage draws its frames
- **THEN** it groups them as asked, and the other section rules still apply

#### Scenario: Rules added to frames that already existed

- **GIVEN** an attended request to add rule cards to existing frames that sit loose on the page
- **WHEN** the stage draws the cards
- **THEN** each card stays beside its frame, no existing frame is moved, no empty section is created, and the approval round asks whether to group those frames

### Requirement: Research audits a codebase on request

When the user asks to audit a codebase, or to find what to improve in it, without naming one question, research SHALL run a code audit and change no repository file. Before judging, it SHALL record the project's verification commands from its CI files or scripts, the decision records that exist and the rejected findings in force; a repository with no command that tells whether the code works SHALL get that as its first finding. It SHALL audit the categories the user names, otherwise correctness, security, performance, tests, debt and architecture, dependencies, tooling, docs and direction, starting from the most changed files. Every finding SHALL cite one to five locations and state its impact, effort, fix risk, confidence and fix sketch; the main thread SHALL open each cited location before reporting the finding, whoever found it, and SHALL drop behavior a platform convention or a decision record settles, misattributed evidence, duplicates and findings a rejection in force covers. The audit SHALL NOT run the test suite, a dependency audit command that sends the dependency list to an outside service, or code of a repository outside the current project unless the user asks for it in that audit. Findings SHALL be ranked by impact level, then confidence, effort and fix risk, after the findings that unblock others, which come first whatever their impact; when the audit covers direction, direction findings SHALL be reported apart, each citing the repository, and the report SHALL state what was not audited. The report SHALL end with research's block for an analysis, whose replies include creating tickets for chosen findings, investigating or planning one, fixing a confirmed defect, rejecting findings and stopping. Findings the user rejects SHALL be recorded in persistent memory per repository in the same turn, with their locations and the audited commit, and a later audit SHALL suppress a recorded finding while the files it cites are unchanged since that commit, and SHALL show it again, naming the earlier rejection, when they changed or that commit no longer exists; findings the main thread drops in vetting SHALL NOT be recorded.

#### Scenario: A finding rejected earlier

- **GIVEN** a finding the user rejected at a recorded commit
- **WHEN** a later audit finds it again and the files it cites are unchanged since that commit
- **THEN** the report omits it and the coverage counts it among the suppressed findings
- **AND** when those files changed, or that commit no longer exists, the report shows it again and says it was rejected before, with the reason

#### Scenario: A child reports a by-design behavior

- **GIVEN** a child that reports honoring a proxy environment variable as a security flaw
- **WHEN** the main thread opens the cited location
- **THEN** the finding is dropped as behavior by design and counted in the coverage, not recorded as a rejection

### Requirement: Defects are fixed at their cause

Before the first edit of a fix for a defect, in a build task, a round of fixes, an urgent production fix or a child's task, once a failing check is judged to show wrong code, and when a plan weighs the approaches to such a fix, the agent SHALL name the defect's cause, the first place the code does what it should not or fails to handle a state it should handle, shown by a reproduction, a regression test seen failing or the code read, and SHALL write it in the task's line or the report. It SHALL fix the defect at that cause with the smallest change that removes it, also when a finding names only the place where the symptom surfaced, wherever it lies in the repository, and SHALL search the cause's literal values and callers for the same fault, fixing each instance in the files the work already changes and reporting the others as findings. A guard, retry, fallback, default value, clamp, swallowed error or discarded state added where the symptom surfaced SHALL NOT count as a fix, unless the code's documented contract, schema or specification allows that state or it comes from input that crosses a trust boundary. While an incident needs its impact reduced, the agent SHALL mitigate first within existing authority. Otherwise, when fixing the cause is a choice the global «Scope and authorization» makes a blocker for a defect that blocks work, such as a contract, stored data or a fix that is not a small understood change, the agent SHALL report it before writing any workaround, so the user chooses between fixing the cause now and a workaround; when the cause lies in a project outside the current one, the agent SHALL ask the question «Scope and authorization» sets for code in an external project, with a workaround here as the other option. A mitigation or a workaround the user accepts SHALL be the smallest fix that holds; a code change SHALL be marked with one comment line that names the cause; each SHALL be named in the report as a workaround with the cause's location, and fixing the cause SHALL be proposed as a ticket unless the same work fixes it. A child SHALL fix only within the paths its brief names and SHALL report other instances, workarounds and outside causes to its parent. New code SHALL NOT copy an existing pattern that is a workaround, such as one a comment marks or a guard the rule does not count as a fix, and SHALL name that pattern in the report. The code review SHALL check for a fix placed where the symptom surfaced instead of at its cause and for new code that copies a workaround, taking as their failure scenario the bad state the cause still produces for its other consumers or the workaround's spread, SHALL NOT rate a fix at the cause, with the other instances fixed with it, as a change the task did not ask for, and SHALL rate each finding by its observable impact. The rule SHALL live in a reference of `flow-build` that `flow-build`, its «Reuse before writing» section, the round of fixes of `flow-verify` and the urgent fixes of `flow-deploy` link with the condition that requires reading it.

#### Scenario: A null crash in a view

- **GIVEN** a view crashes on a null field that the API should never return
- **WHEN** the agent fixes the defect
- **THEN** it names the code that produces the null as the cause and fixes it there
- **AND** a null check added in the view alone is not reported as the fix

#### Scenario: A cause in another project

- **GIVEN** the cause of a defect is in a service of a project outside the current one
- **WHEN** the agent is about to fix it
- **THEN** before writing any workaround it asks whether to continue there from this session or open a session in that project, with a workaround here as the other option
- **AND** a workaround the user accepts is marked with one comment line naming the cause, named as a workaround in the report, and a ticket is proposed for the cause

#### Scenario: A cause in an untouched module of the same repository

- **GIVEN** a view crashes because a helper in a module the work has not touched returns a value its contract forbids
- **WHEN** the agent fixes the defect
- **THEN** it fixes the helper, without a guard in the view and without proposing a ticket

#### Scenario: Copying an existing workaround

- **GIVEN** the surrounding code repeats a retry that a comment marks as a temporary workaround
- **WHEN** a task writes new code in that area
- **THEN** the new code does not copy the retry and the report names it

### Requirement: The user decides whether a worktree is created

When the work is already routed to build and authorized and the user asked for a worktree, the agent SHALL create it; a question from the user about using a worktree is not such a request. When the current checkout is on the base or on this work's own branch and holds no changes to tracked files that the agent did not make, the agent SHALL work there, creating the work branch only when the delivery makes commits. Otherwise, before creating any branch or switching, the agent SHALL ask the user as a decision whether to create a worktree, stating what it observed, such as another session's changes or another branch checked out, unless the session already works in one; when such changes appear later in the work, it SHALL ask then. Under unattended delegation the agent decides as that skill's step 5 sets. An accepted worktree SHALL go in the host's native worktree location; when the host has none, or it is not verified, the agent SHALL ask for the path. An accepted worktree inside the repository SHALL be added to `$(git rev-parse --git-common-dir)/info/exclude`. The work's checkpoint, in its change folder or its temporary checkpoint file, SHALL name the worktree's path and branch, and work without a checkpoint SHALL find it through `git worktree list`; the agent SHALL NOT write persistent memory because a worktree was created or removed. A delegated child isolated in its own worktree is a worktree the agent creates, and a delegated child SHALL NOT create a worktree: it compares with the base through `git diff` or `git show <base>:<path>`. Under unattended delegation the agent SHALL NOT ask: it SHALL decide at that moment whether to create the worktree, choosing the option that keeps the work going, and SHALL name its choice and reason in its report.

#### Scenario: Unattended work meets another session's changes
- **WHEN** unattended delegation is active and the current checkout holds changes the task did not make
- **THEN** the agent decides without asking, such as creating a worktree in the host's native location so the work continues, and its report names that choice and its reason.

#### Scenario: Another session writes in the checkout
- **WHEN** a build finds uncommitted changes it did not make in the current checkout
- **THEN** the agent asks whether to work in a worktree, with those changes as the situation, and creates none before the answer.

#### Scenario: The user asked for a worktree
- **WHEN** the work is routed to build and authorized and the user's request says to take it in a worktree
- **THEN** the agent creates it in the host's native location without asking again.

#### Scenario: The user only asks about a worktree
- **WHEN** the user asks whether a ticket can be taken in a worktree
- **THEN** the agent answers and creates no worktree.

### Requirement: Commands run from the checkout that holds the change

The agent SHALL give each command that acts on one repository, such as a package manager, build, test or Git command, the repository or worktree that holds the change by an absolute path, through the tool's own directory option, such as `git -C`, `pnpm --dir` or `go -C`, and only for a tool without one through `cd <absolute path> && <command>` in the same call or, where the host refuses that, absolute paths for the files it reads. It SHALL NOT rely on a `cd` made in an earlier call, and SHALL NOT run the command from the root of a declared workspace or after a chain of relative `cd ..`. A change that spans several repositories SHALL run each command in its own repository. Reads that span the workspace and operational tasks that change no repository are outside this requirement.

#### Scenario: A package manager command in a workspace
- **WHEN** a session started at a declared workspace root that holds several repositories runs `pnpm install` for one of them
- **THEN** it runs the command with that repository's or worktree's absolute path as working directory, and no `package.json`, lockfile or `node_modules` appears at the workspace root.

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

### Requirement: A ticket states what is needed, not how it is built

A ticket the skill creates SHALL state the outcome, how it will be verified and what is out of scope, and SHALL carry at most one solution hint, in a separate section marked as non-binding, with at most five locations and one sentence of direction. It SHALL NOT carry a list of implementation steps, a choice between approaches, a contract definition or a file-by-file change list. `flow-plan` SHALL treat a ticket's hint as a claim to verify.

#### Scenario: The hint would need a decision
- **WHEN** writing the hint would require choosing between approaches or defining a contract
- **THEN** the ticket says that it requires a plan and carries no hint.

### Requirement: Work is broken down into a parent ticket with children on one approval

The skill SHALL propose a breakdown when the work names more than one change that integrates separately, touches more than one repository, or has more than 7 acceptance criteria. It SHALL show the whole set, with the parent, each child and their dependencies, and SHALL ask one approval for the whole set before the first write. It SHALL write the tickets one at a time, all in the primary tracker and each child naming its repository, SHALL record each one as it is created and read each one back, and SHALL stop and report what was created when a write fails. A parent ticket with children SHALL be created as Jira's epic type, as a parent issue with sub-issues in GitHub, and as a parent issue with sub-issues in Linear, without creating a project.

#### Scenario: Unattended work reaches a breakdown
- **WHEN** unattended delegation is active and the work would need a parent ticket with children
- **THEN** the skill creates none of them and reports the proposed set as unresolved.

#### Scenario: Unattended work opens a ticket for a finding
- **WHEN** unattended delegation is active and a finding needs a single ticket
- **THEN** the skill creates it without asking for a confirmation.

#### Scenario: A child is not separately integrable
- **WHEN** a part of the work cannot be integrated and accepted on its own
- **THEN** it is not proposed as a child ticket.

### Requirement: An existing ticket is validated without rewriting another person's content

When the skill validates a ticket that already exists, it SHALL first check whether the authenticated account is the ticket's author and what the account may write. When the account is not the author, it SHALL NOT edit the ticket: it SHALL show a comment with what is missing and the open questions, each with a recommended answer, and post it after the user confirms. When the account is the author, it SHALL show the change and edit after the user confirms. When the account cannot write to the tracker, it SHALL hand the report to the user. An undecided business rule SHALL always go to the ticket's owner as a question.

#### Scenario: A developer validates a product owner's ticket
- **WHEN** the authenticated account did not write the ticket and its criteria are missing
- **THEN** the skill shows a comment naming the missing criteria and its questions, and posts it only after the user confirms.

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

### Requirement: A requirement document's state is set by validation

The state of a requirement document SHALL be one of draft, iterating, approved by business, technical review and ready for development, and SHALL be set by what validation found, never by its author's label. The document SHALL reach approved by business only when the fixed checklist passes, an independent `hive-review-prd` child has reviewed it, no rule is still to decide and the user approves it explicitly. It SHALL reach ready for development only through the technical review, run in a session with the repository, when no rule is not checked and every rule the code contradicts or blocks has its owner's answer. A technical reviewer who is not the author SHALL NOT rewrite a rule.

#### Scenario: Approval by business
- **WHEN** the checklist passes, the independent review has run, no rule is to decide and the user approves
- **THEN** the document is approved by business and that version is frozen.

#### Scenario: A rule is still to decide
- **WHEN** the user asks to approve a document that holds a rule still to decide
- **THEN** the document stays iterating and the skill reports the rule and who answers it.

#### Scenario: Technical review with the repository
- **WHEN** a session with the repository takes a version approved by business
- **THEN** the skill gives each rule a verdict against the code, writes a question with a recommended answer for each contradiction or block, and marks the document ready for development only when none is unanswered and none is not checked.

#### Scenario: A session without the repository
- **WHEN** a session without the repository is asked to mark a document ready for development
- **THEN** the skill does not, and says the technical review needs a session with the repository.

### Requirement: A requirement document lives in a shared document and its approved versions are frozen

The skill SHALL write the document in a document tool the session can write to, at the destination the user names, asking for it when none was named, and SHALL name for Notion, Google Docs and Claude Docs how to create, copy, comment and obtain the link. Without such a tool it SHALL write the document as a Markdown file, at `_support/docs/requirements/<module>/prd.md` inside a repository and in the working folder outside one, and render it to PDF beside it with a browser already installed, installing nothing. The rendering script SHALL treat the Markdown file as untrusted text: it SHALL escape any HTML the file carries, emit a page that loads no remote resource and runs no script, and launch the browser with a temporary profile. At each approval it SHALL freeze the version as a copy in the same tool, titled with its version, state and date, which it never writes again, and link it from the live document's version history; in the Markdown mode the copy SHALL be `prd-v<N>.md` with its PDF.

#### Scenario: A document tool is connected
- **WHEN** the session can write to a document tool
- **THEN** the document is created there and each approved version is a titled copy linked from the live document's history.

#### Scenario: No document tool, inside a repository
- **WHEN** the session has no document tool and works inside a repository
- **THEN** the document is written to `_support/docs/requirements/<module>/prd.md` and each approved version to `prd-v<N>.md` with its PDF beside it, without a commit, and the skill reports the path and that it is not committed.

#### Scenario: A document that carries HTML
- **WHEN** the Markdown file holds a script tag or a remote image
- **THEN** the PDF shows them as text, nothing is fetched and no script runs.

#### Scenario: No installed browser
- **WHEN** the PDF is rendered on a machine with no Chromium-family browser
- **THEN** the script leaves the printable page, exits with a non-zero code and says how to print it.

### Requirement: Bounded E2E coordination through existing flows

Hive SHALL offer explicitly activated E2E coordination through its existing flows and one canonical coordination procedure. Its activation SHALL bind an objective, acceptance criteria, delivery endpoint, allowed effects, isolation authorization and stopping conditions. It SHALL NOT infer implementation authority from a research or planning request. It SHALL continue already authorized stage transitions without requiring a separate user command at each stage. Scope changes and effects outside the grant SHALL retain their existing user decision. Unattended status SHALL remain separate from E2E activation.

#### Scenario: Authorized multiunit objective
- **WHEN** an E2E grant and ready plan cover independent deliverables
- **THEN** the coordinator assigns owners, selects required skills, arranges independent verification and continues toward the authorized endpoint without asking the user which tool or stage comes next.

#### Scenario: Missing isolation authority
- **WHEN** the proposed coordination needs worktrees but the activation did not cover their creation
- **THEN** the coordinator resolves that decision before creation, and the existing ordinary-worktree rule remains in force.

### Requirement: Local deliverable owner exception

Only a designated E2E local-owner role SHALL be allowed to create local commits within its assigned repository, worktree, branch and write paths, using the existing scoped-commit and secret-scan mechanism. This named exception SHALL NOT expand ordinary implementation roles. The coordinator SHALL create and assign worktrees, arrange exclusive resources, launch independent verifiers and retain all publication, pull-request, tracker and integration operations. Owners SHALL NOT launch children, publish, integrate remotely, alter another assignment or infer authority from record fields.

#### Scenario: Local handoff
- **WHEN** an owner finishes an assigned unit
- **THEN** it returns the exact local head, command results, evidence and limitations, stops its own processes, and waits for the coordinator's next assignment rather than publishing its branch.

### Requirement: Root-owned execution record

Multiunit execution SHALL have one versioned local execution record linked from the existing checkpoint. The coordinator SHALL be its only writer. Record operations SHALL reject unknown versions, duplicate keys/IDs, invalid task/criterion/dependency references, cycles, conflicting active assignments and stale generations without losing the prior valid state. Machine paths, native IDs and raw run evidence SHALL remain outside versioned change records. Existing progress/checkpoint commands SHALL remain compatible.

#### Scenario: Stale update
- **WHEN** an update names a generation that no longer matches the record
- **THEN** the operation fails with a correction naming the stale state and preserves the current record.

### Requirement: Exact-head handoff and independent evidence

Before a handoff may be recorded verified, the helper SHALL check local Git identity, assigned branch/worktree, full head, ancestry, cleanliness and allowed changed paths. Verification metadata SHALL identify the checked head, criteria, commands, results, available evidence and a verifier distinct from the author. Missing, failed, unavailable or stale evidence SHALL prevent verified status. The coordinator SHALL inspect evidence; structurally valid declarations SHALL NOT be treated as proof of execution or semantic acceptance. Changed code SHALL invalidate affected verdicts.

#### Scenario: Reused verdict after a fix
- **WHEN** an owner returns a new head with only an older head's verdict
- **THEN** the verified transition is rejected and the coordinator obtains evidence for the new head before continuing.

### Requirement: Recoverable replacement without overlapping writers

Replacement SHALL preserve the unit's work and history, require observable completion or stop acknowledgment for the former writer, reconcile Git and resources, and create a new assignment before another writer begins. A missing native ID or elapsed time alone SHALL NOT prove the old writer stopped. Active or ownership-uncertain branches, worktrees and resources SHALL remain protected from cleanup. This contract SHALL NOT promise background scheduling or persistence beyond the host's verified capabilities.

#### Scenario: Partial return
- **WHEN** an owner returns partial work and native completion is confirmed
- **THEN** the coordinator assigns a replacement from the recorded head, limitations and next action without asking the human to reconstruct the session.

### Requirement: Project verification contract and feature map

Hive SHALL generate and maintain a project-owned verification contract indexing existing Launch, Doctor, Drive, Evidence and Cleanup capabilities and the features their checks cover. Initialization SHALL refuse to overwrite. Validation SHALL check versions, required fields, feature references, referenced files and confined paths, without executing commands. Unknown or untested capabilities SHALL remain explicitly unverified; unsupported and non-applicable entries SHALL state reasons. The runtime flow SHALL inspect effects and actual outputs before accepting or reusing evidence, and SHALL avoid creating a competing contract when the project already has an equivalent source.

#### Scenario: Valid shape, unverified application
- **WHEN** a project contract passes structural validation but Doctor or a feature's Drive has not passed
- **THEN** the flow reports that behavior unverified and does not infer readiness from the validator's exit status.

### Requirement: Deterministic correction and suppression review

For a repeatable error addressed by the work, the correction procedure SHALL select the earliest suitable enforcement layer among architecture, types, actionable lint, tests and prose. It SHALL retain one scoped rule-to-enforcer-to-negative-check mapping and remove procedural prose replaced by the mechanism. Flow-verify SHALL require an explicit inventory of added or widened suppressions, skipped tests and configuration exclusions in the diff, and a justified disposition of each candidate, with any selected review mechanism including native review. This procedure SHALL have one canonical home, referenced by review roles and supplied in native review requirements. Missing review coverage SHALL remain unverified. A textual match alone SHALL NOT be a defect; justified exceptions and authorized workaround explanations SHALL remain permitted.

#### Scenario: Suppression hides the failing criterion
- **WHEN** a proposed fix suppresses the check that exposes the defect without fixing its cause
- **THEN** review demonstrates the remaining failure and returns the unit for correction rather than accepting green output alone.

#### Scenario: Legitimate exception
- **WHEN** an exception has a demonstrated scope and justification consistent with requirements
- **THEN** review does not reject it solely because it uses a suppression directive or explanatory comment.

### Requirement: Local integration of verified owner heads

The coordinator SHALL combine independently verified unit heads into the declared integration target, read back the resulting Git head and run the check spanning the units before marking them integrated. The execution record SHALL distinguish local integration from remote publication. The bounded pilot SHALL exercise local integration without a push, pull request or remote merge. Existing publication rules SHALL continue to govern actual delivery.

#### Scenario: Two individually verified units
- **WHEN** both unit heads have valid independent verdicts
- **THEN** the coordinator still checks their combined local result before recording integrated, and the record does not claim remote delivery from that local result.

### Requirement: Pull requests carry before and after captures of changed screens

Before its first edit of anything an existing screen renders, a build SHALL have its main thread serve the unchanged checkout and capture each region its plan, or the request when it has no plan, changes, with the widths, themes and states the verification will check, before any child task edits that surface; those captures SHALL NOT be published to the viewer, and the checkpoint SHALL record their paths or their absence. When the verification produced captures, the remote is github.com and the installed `gh` offers `--attach`, the pull request body SHALL hold a before and after table per region, width, theme and state, with the images uploaded through `--attach`, none added to a commit, and a cell that names a new screen or a missing before capture. In a public repository or on GitHub Enterprise Server, it SHALL attach only captures that show seeded data and test accounts, without internal URLs or real email addresses, and only with the user's one-time authorization for that work. Otherwise the body SHALL keep its text-only form and name the captures as kept locally. At the close, a capture SHALL be treated as replaced by the pull request only when its uploaded URL was read back in the merged body; other captures SHALL be kept.

#### Scenario: Existing screen changed in a private repository
- **WHEN** a build changes an existing screen in a private github.com repository and `--attach` is available
- **THEN** the pull request body shows each changed region's before and after images, uploaded and not committed

#### Scenario: Child edits the screen
- **WHEN** the plan delegates the screen's task to a child
- **THEN** the main thread captures the before state before launching it, and the child's brief asks for no capture

#### Scenario: New screen
- **WHEN** the change adds a screen that did not exist
- **THEN** its row has no before image and says the screen is new

#### Scenario: No attach option
- **WHEN** the installed `gh` lacks `--attach` or the remote is neither github.com nor GitHub Enterprise Server
- **THEN** the body keeps its text-only form, names the captures as kept locally, and the close keeps them

#### Scenario: Public repository without authorization
- **WHEN** the repository is public and the user has not authorized the public upload for this work
- **THEN** the captures stay local, the body keeps its text-only form, and the close keeps them

#### Scenario: Partial upload
- **WHEN** some attachments fail while the pull request is created
- **THEN** the agent attaches the missing ones with `gh pr edit --attach` and edits the body only from the published body

#### Scenario: Close after the merge
- **WHEN** the work is integrated and the agent read back each uploaded image URL in the merged body
- **THEN** the close proposes deleting those local captures and keeps any capture without a read-back URL

### Requirement: The close lists the project's determinism candidates

A work item's final report SHALL list, only when the work has one, each determinism candidate of the project: a repetition the work did not turn into a mechanism, namely an inline script or command chain the work ran two or more times for the same purpose, or a correction the user gave twice in the work, with what repeated, how many times and the proposed mechanism, without asking a new question. The close SHALL search the project's persistent memory for the candidate by its purpose and create or update its record, named after its tool and target, with the session and date. The close footer SHALL offer turning a candidate seen in two sessions into a mechanism as a reply of its own; a candidate is not a finding.

#### Scenario: Repeated inline script

- **GIVEN** a work item ran the same inline link-check script three times
- **WHEN** its close writes the final report
- **THEN** the report lists one candidate naming the script's purpose, the count and a proposed project mechanism, and the project's memory holds its record with this session

#### Scenario: Candidate seen in a second session

- **GIVEN** the project's memory already records a candidate from an earlier session
- **WHEN** a later work item's close finds the same candidate
- **THEN** the record counts two sessions and the close footer offers turning it into a mechanism

#### Scenario: No candidate

- **GIVEN** a work item repeated no script, command chain or correction
- **WHEN** its close writes the final report
- **THEN** the report has no candidates part and the close writes no candidate record

### Requirement: Pending questions are not duplicated across tools

The agent SHALL NOT issue the same pending question through both asynchronous and waiting tools, or issue another card while its answer is pending. A tool's acceptance receipt SHALL NOT count as a user answer.

#### Scenario: An asynchronous question is still pending

- **GIVEN** a permitted asynchronous question was issued
- **WHEN** its answer is still pending
- **THEN** the agent does not send the same question again through a waiting tool or a second card

### Requirement: Decision information survives a required change of form

When no waiting tool carries a question and the host requires another form than Hive's decision format, the agent SHALL retain the decision's ID or fixed header, the alternatives with their letters, their consequences and the recommended option marked `(★)` with its reason, expressed in the form the host permits. It SHALL NOT interpret a different form as removing that information.

#### Scenario: The host requires a permission request in concise text

- **GIVEN** the host requires text for a permission request and no waiting tool carries it
- **WHEN** the agent asks permission
- **THEN** it states the recommended option with `(★)` and why, and the alternative with its letter and consequence
- **AND** it does not issue a duplicate card

### Requirement: Decision rounds go through the card alone

When the selected native question tool waits for the answer and gives each option a description field, the agent SHALL ask a round of decisions or findings through that tool without writing them out in the text before it, also on a host whose own instructions reserve that tool for optional questions, such as Codex in its Default mode: the text before the first card SHALL hold the context line, the report or explanation the round needs and the waiting line, which says what is awaited and how many decisions are still to come, and SHALL NOT hold a decision's or a finding's title line, `Situation:` line, option lines or `I recommend` line. Each card SHALL stand alone: its question SHALL carry the decision's or finding's ID and title with the situation, or the fixed header of a session question. Where no waiting tool carries the question, or the selected card has no description field per option, the agent SHALL write the whole round in the text in its full format; when a card that follows a round written in the text comes back without an answer, the agent SHALL add one line naming the decisions still open and SHALL NOT write the round again. A plain acceptance SHALL apply the starred options only of decisions the user was shown, in a card that opened or in the text.

#### Scenario: A round answered through the card

- **GIVEN** a host whose question tool waits and shows a description per option
- **WHEN** the agent asks a round of three decisions
- **THEN** the text before the card holds the context line, the report and the waiting line, with no decision block
- **AND** the card asks the three decisions and the turn ends with it

#### Scenario: Codex in its Default mode

- **GIVEN** a Codex session in Default mode that offers `request_user_input`
- **WHEN** the agent asks the delivery round before its first edit
- **THEN** it asks the round through `request_user_input` and the turn waits for the answer
- **AND** it does not write the round as a plain-text question

#### Scenario: A card without option descriptions

- **GIVEN** a host whose question tool has no description field per option and permits the full round in text
- **WHEN** the agent asks a round
- **THEN** the whole round is written in the text, and the card follows it
- **AND** when that card comes back without an answer, one line names the decisions still open and the round is not written again

### Requirement: An unanswered card is handled by who left it unanswered

The host's required response SHALL take precedence. Where the host leaves that response to Hive, when a waiting card that carried the round alone comes back without an answer, the agent SHALL ask no further card of that round before acting as follows. The answers the card did return SHALL stand. When the card closed by itself, because the tool returned an empty answer or a timeout while the turn continued and its result held no rejection, dismissal or decline, the agent SHALL write, in its next message, every decision and finding of the round still open in its full format, those of the cards not yet asked included, with a waiting line saying that "ok" applies every starred option and that a reply such as `D2 B` changes one. When the user left it unanswered, because the tool returned a rejection, a dismissal or a decline, also as an error and also on a host that returns a decline after its own timeout, or the host ended the turn while the card was open, or the card returned a request to talk, answers to only some of its questions, or free text that adds, removes or replaces a decision or an option, the agent SHALL NOT write the round out in the text on its own and SHALL act on what the user said, in the card or in their next message when the host ended the turn, by the first of these that fits: answers by ID or a plain acceptance are applied and the decisions left open are asked again through the card; a cancel made by mistake, or a request for the questions, gets the same round again through the card with the same IDs; a statement that the direction is wrong, a change to what the decisions rest on, or free text that adds, removes or replaces a decision or an option, gets the open decisions it touches dropped, reworked and asked as a new round with new IDs; a request to drop the questions gets them dropped and not asked again, except one a later step cannot proceed without; a change of subject, a question or an investigation that leaves the decisions as they were gets its answer followed by one line naming the decisions still unanswered and how to get them back, which are asked again through the card when the user says so or asks to continue work that cannot go on without them. When the user has said nothing about the open decisions yet, as after a decline while the turn continues, a card answered in part or a request to talk that names no subject, the agent SHALL reply with that one line alone and wait. A decision asked again SHALL keep its ID, and one line saying which decisions return SHALL precede a card that returns. Later rounds SHALL go through the card again.

#### Scenario: A card closes by itself

- **GIVEN** a card that returns empty after the host's timeout and no native instruction requiring another response
- **WHEN** the agent replies
- **THEN** its message writes each decision still open with its situation, its options and its recommendation
- **AND** it asks none of that round's remaining cards
- **AND** the next round goes through the card

#### Scenario: The user cancels by mistake

- **GIVEN** the user dismissed a card and the host ended the turn
- **WHEN** the user's next message says it was a mistake or asks for the questions again
- **THEN** the agent asks the same round again through the card, with the same IDs and options

#### Scenario: The user cancels because the direction is wrong

- **GIVEN** the user dismissed a card
- **WHEN** the user's next message says the work is going in the wrong direction
- **THEN** the agent drops the open decisions, reworks them with what the user said and asks a new round with new IDs

#### Scenario: The user cancels and changes the subject

- **GIVEN** the user dismissed a card
- **WHEN** the user's next message asks a question that leaves the decisions as they were
- **THEN** the agent answers it and ends the reply with one line naming the decisions still unanswered
- **AND** it does not write the round out in the text
- **AND** it asks the round again through the card when the user says so

#### Scenario: The user cancels and answers by ID

- **GIVEN** the user dismissed a card of three decisions
- **WHEN** the user's next message answers two of them by their ID
- **THEN** those two answers apply and the third is asked again through the card

#### Scenario: The user cancels and asks to drop the questions

- **GIVEN** the user dismissed a card
- **WHEN** the user's next message says to drop those questions
- **THEN** the agent drops them and does not ask them again, unless a later step cannot proceed without one

#### Scenario: Free text replaces an option

- **GIVEN** a card whose answer is free text that replaces one option of a decision
- **WHEN** the agent replies
- **THEN** it reworks that decision with what the user said and asks it again through the card with a new ID

#### Scenario: The host directs continuation after an unanswered optional question

- **GIVEN** the host requires best judgment after an unanswered optional waiting card
- **WHEN** that card returns no answers
- **THEN** the agent follows that instruction without re-asking the question
- **AND** the empty response grants no permission for an action requiring authorization

### Requirement: Entry router with playbooks
Hive SHALL provide an entry skill, `hive-mode`, that replaces `flow-research`. It SHALL read before asking, play the request back, ask what is missing and settle what the user needs, and SHALL NOT edit repository files itself; it MAY hand the work to `flow-build` in the same turn only for a small understood change that the user's own request named with its target, never after a reply to its question or proposal. It SHALL classify the work and open at most one playbook under `playbooks/<kind>.md`, record its name in the change folder's `proposal.md` (written by `flow-plan`) or, for work without a change folder, in a `- Playbook:` line of the checkpoint file, and, when no playbook fits or the work is large and cross-cutting, follow the general route from research to plan or build. A playbook names the order of the stages, what is specific to that kind and the leaves it uses; the stages (`flow-figma-design`, `flow-plan`, `flow-build`, `flow-verify`, `flow-close`, `flow-deploy`) SHALL remain commands that read the playbook's specific steps.

#### Scenario: A reported defect with a fix request
- **WHEN** the user reports a defect and asks for its fix
- **THEN** `hive-mode` opens the `bugfix` playbook, and build and verify fix the defect at its cause as «Defects are fixed at their cause» sets, with its regression evidence.

#### Scenario: An investigation confirms a defect
- **GIVEN** work running the `investigate` playbook
- **WHEN** the findings confirm a defect and the user asks for its fix
- **THEN** the work hands off explicitly to the `bugfix` playbook as a scope change, and the new playbook is recorded

### Requirement: Verify complements and adversarial review option
Before the code review, `flow-verify` SHALL read the leaves that remove narrating comments, check the change's impact beyond the diff, and check the prose of the pull request, commits and documentation. A review of the candidate by another provider's agent SHALL run only when the user explicitly asks for it: no delivery round SHALL ask or suggest it and no agent SHALL probe other hosts before that request. When the user asks, the agent SHALL list the installed and signed-in command-line agents of other providers for the user to choose, under the reuse and spending limits the user set; when none is available, it SHALL report the review as unavailable.

#### Scenario: Delivery round without an adversarial question
- **WHEN** the agent asks the delivery and review questions for a work item
- **THEN** the round has no adversarial review question and no command checks other hosts' installation or sign-in

#### Scenario: The user asks for a review by another provider
- **WHEN** the user asks for an adversarial review of the candidate and picks an agent
- **THEN** verify runs it with that agent on that work item's candidate within its limit and reruns each reported finding before keeping it

### Requirement: Decision log for unattended and coordinated work
Unattended work and bounded E2E coordination SHALL keep one decision log at `workspace/YYYY-MM-DD-<work>/decisions.tsv` of the main checkout's support home, a tab-separated file with the columns `ts`, `phase`, `decision`, `why`, `evidence` and `result`, one single-line row per decision and one per external effect, whose evidence names the target and the change and whose result names how to revert it. Evidence SHALL be a link or path, never a paragraph. Its format SHALL have one home, in `unattended-delegation`, which E2E coordination links; in E2E only the coordinator writes it. The log SHALL NOT be committed or published unless the user asks, and the final report SHALL cite it by path. The task record SHALL keep only progress.

#### Scenario: An unattended run takes an external effect
- **WHEN** an unattended run changes a resource outside the repository within its authorized ceiling
- **THEN** `decisions.tsv` gains one row whose evidence names the target and the change and whose result names how to revert it, and the task record gains no copy of it

### Requirement: Moving between stages shows the next-step block

Every move from one flow to the next SHALL end the report with the next-step block and its command, except the continuations its rule names. A reply to a question or the acceptance of a proposal SHALL only define or redirect scope and SHALL NOT start a flow by itself, except a fix the close footer or a findings round sends and a continuation the user's own request set. Research SHALL continue into build in the same turn only when the user's own request named the change and its target.

#### Scenario: A ticket picked for research

- **GIVEN** the user picks a ticket for research without asking for a change
- **WHEN** the agent's proposal is accepted with a reply such as "1"
- **THEN** the next message ends with the next-step block toward `flow-build` or `flow-plan`
- **AND** no build starts until the user's reply names it

### Requirement: Session questions carry fixed headers

The delivery mode, the code review mechanism and the security review questions SHALL carry a fixed header in the session language instead of a decision ID, with the rest of the decision format. A mechanical change SHALL get no code review question; every other change SHALL get a code review, so the question SHALL offer only review mechanisms, also under `Delivery: direct-base`.

#### Scenario: Delivery round of a UI fix

- **GIVEN** a work item that fixes a screen flow
- **WHEN** the agent asks the session questions
- **THEN** they read as **Delivery** and **Review** without `D<n>`, and the review options are mechanisms

### Requirement: Tracker states are discovered and cached

Before moving a ticket, the agent SHALL read that ticket's available transitions live, through the first tool that lists them in this order: the tracker's CLI, the host's authenticated MCP tool for that tracker, then the tracker's REST API with the user's global API token; it SHALL move to the state the project uses for that purpose by its exact name, and SHALL report a blocker instead of picking among states that share a category when none is recorded for that purpose. It SHALL keep the names and categories of the project's states in the Engram topic `tracker/states/<tracker>/<prefix>`, where `<prefix>` is the project's key prefix plus the ticket type when the workflow differs by type, written when it discovers them or after a failed transition, and SHALL use it to choose the state, never in place of the live read.

#### Scenario: A localized workflow

- **GIVEN** a Jira project whose in-progress state is named «En curso»
- **WHEN** a transition to "In Progress" fails
- **THEN** the agent reads the transitions, moves the ticket to «En curso» and writes the cache

### Requirement: A ticket is documented with evidence before cleanup

At a work item's close that removes material, the agent SHALL document its ticket with the evidence the work produced, such as captures and test output, before any step that removes support material, temporary files or worktrees; a close of paused or `hold` work SHALL publish none.

#### Scenario: Captures before cleanup

- **GIVEN** a closed UI fix with before and after captures in its workspace
- **WHEN** the close runs
- **THEN** the captures reach the ticket before the cleanup removes them

### Requirement: Verification floor by kind of change

One rule SHALL set, per kind of change (mechanical, defect fix, feature, UI, risky, performance), the checks that always run, linking each check's own rule. The code review SHALL run on every non-mechanical change. A security review SHALL be recommended through its own session question for each work item that touches a security boundary, and SHALL run when the user accepts it.

#### Scenario: A UI defect fix under hold

- **GIVEN** a defect fix to a screen delivered under `hold`
- **WHEN** it is verified
- **THEN** its regression test runs failing then passing, the code review and the UX review run, and before and after captures are taken

### Requirement: The final report carries evidence and a try-it block

The final report SHALL carry a defect's regression test output seen failing and then passing, and, when the change can be seen or run, a block with the URL or command, the test account's index row, the steps and the expected result, in every delivery mode.

#### Scenario: Hold delivery

- **GIVEN** a change delivered under `hold`
- **WHEN** the agent reports it
- **THEN** the report holds the failing and passing output and the try-it block

### Requirement: Tracker writes follow the tools' capabilities

The agent SHALL write to a tracker through the first tool that has the needed capability, in this order: the tracker's CLI, the host's authenticated MCP tool, then the tracker's REST API with the user's global API token. The token SHALL live in a file the user creates once under Hive's global secret location, `~/.config/hive/secrets/` (folder `0700`, file `0600`); a script SHALL read it, and the agent SHALL check only that it exists, SHALL never write it, and SHALL ask the user to create it when it is missing. The token SHALL never appear in a command's arguments, output, an error, a message, a durable place or another provider's agent. Attachments and worklogs the CLI cannot write SHALL go through the next tool in that order.

#### Scenario: Attaching captures in Jira

- **GIVEN** `acli` lists but cannot upload attachments and the host has no MCP tool that uploads
- **WHEN** the close documents the ticket with its captures
- **THEN** the captures upload through the REST API with the global token and the token appears in no output

#### Scenario: The token file is missing

- **GIVEN** no `~/.config/hive/secrets/jira-api-token`
- **WHEN** a worklog must be written and no other tool can write it
- **THEN** the agent asks the user to create the file with its permissions and does not create it

### Requirement: Tickets record the time worked

For a ticket in a tracker with native time tracking, the agent SHALL add a worklog at the work item's delivery point (integration, the `ready` stop or the `hold` stop) to each ticket the session worked, after reading the worklogs its account already has on that ticket, holding only the converted time and no description, and SHALL update that same worklog at the close with the time added after delivery. Each ticket's time SHALL be the active time from that ticket's first message in each session that worked it, minus only the waits for the user; tickets worked together in a session SHALL each receive the full time. It SHALL be converted as the active time without the remote CI waits it recognizes times the factor (2.0 by default, between 1 and 3) plus those waits at 1. It SHALL be measured with `hive time` from the host's native transcript where the manager reads one, and from the agent's own timestamps otherwise; the plan's saved state SHALL keep the host and session ID of each session that worked the ticket so the delivering session measures them again. The worklog's ID, the real time, the factor and the ticket type SHALL stay in the work's close record in persistent memory and SHALL NOT appear in the tracker.

#### Scenario: A delivered Jira ticket

- **GIVEN** a session that researched, built and integrated a Jira ticket in 40 minutes of active time, 10 of them waiting on remote CI
- **WHEN** the work reaches its integration
- **THEN** the ticket gains one worklog of 70 minutes with no description, and the close record keeps the worklog ID, 40 minutes, factor 2.0 and the ticket type

#### Scenario: A tracker without time tracking

- **GIVEN** a GitHub issue
- **WHEN** the work reaches its delivery point
- **THEN** no worklog is written and nothing about time reaches the issue

### Requirement: `hive time` reports a session's active time

`hive time` SHALL print one JSON object with `result` (`ok`, `partial` or `no_session_file`), `host`, `session`, `since`, `until`, `active_seconds`, `user_wait_seconds`, `ci_wait_seconds`, `factor` and `converted_seconds`, rounded up to the minute. It SHALL read the whole transcript; when it cannot, it SHALL report `partial`. A host without a transcript the manager reads SHALL give `no_session_file` with exit code 0. A factor below 1 or above 3 SHALL exit with code 2. Only messages a person wrote and answers to a question tool SHALL end a user wait.

#### Scenario: A background CI wait in Claude Code

- **GIVEN** a Claude Code session that started `gh pr checks --watch` in the background and resumed on its task notification 6 minutes later
- **WHEN** `hive time` reads the session
- **THEN** those 6 minutes count as CI wait, not as user wait
