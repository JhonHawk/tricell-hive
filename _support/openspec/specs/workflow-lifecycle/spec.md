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
When a repository's `## Hive` declares `Other trackers`, the workflow guidance SHALL read and search tickets in every declared tracker, SHALL move and comment each ticket in the tracker that declares its key prefix, using the primary `Tracker` for a key no tracker declares, and SHALL create new tickets only in the primary tracker. Declaring a tracker in `## Hive` SHALL authorize those moves and comments as it does for the primary tracker.

#### Scenario: Ticket of an additional tracker reaches a later environment
- **WHEN** a deployment is verified and `Completion` holds a value for the prefix `XYZ`
- **THEN** the guidance moves each `XYZ-` ticket whose change that ref holds to the declared state in the tracker that declares `XYZ`.

#### Scenario: Mirror tickets
- **WHEN** one work item names a ticket in the primary tracker and another in an additional tracker
- **THEN** the guidance moves and comments both, each in its own tracker.

#### Scenario: New ticket
- **WHEN** a flow proposes or creates a ticket in a repository with additional trackers
- **THEN** it searches every declared tracker for an existing one and creates the new ticket in the primary tracker.

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
The close's branch cleanup SHALL run in each repository where the work created a branch, a worktree or a commit or wrote its change folder, plus the repository that holds `Specs`; a hygiene request SHALL cover the current repository, its `Specs` repository and the repositories that the pending-close records in persistent memory name. Every worktree removal SHALL keep the worktree and both copies of its branch when it has uncommitted or untracked files, or an ignored entry other than these three: a dependency or build folder or an entry in one; an entry whose content the main checkout of the worktree's own repository already holds at the same path; and, at any depth, `*.tsbuildinfo`, `next-env.d.ts`, `*.pyc`, `.turbo/`, `__pycache__/`, `.pytest_cache/`, `.mypy_cache/`, `.ruff_cache/` or `.DS_Store`, or an entry inside one of those folders; it SHALL first stop the processes the work started inside its own worktree, by process ID, and SHALL keep any worktree where a running process has its working directory or where that cannot be checked. A squash- or rebase-merged branch SHALL count as merged only when its pull request's `headRefOid` equals the head of each local and remote copy to be deleted; a remote copy that moved SHALL be reported with the commits after that head. After the work's own branch, the sweep SHALL check the other branches whose pull requests the user authored and merged into the base, the local branches `git branch --merged` lists that persistent memory records for the user's earlier work in a repository without pull requests, and the temporary or preview branches the work created. It SHALL NOT switch any checkout's branch, and SHALL keep, reporting the reason, a branch checked out in another checkout and a branch that persistent memory records for work whose close is not recorded as done, unless a pending-close record for that work is older than 24 hours, in which case it SHALL resume that work's pending close steps in order instead of removing its branch or worktree. The global guidance SHALL authorize deleting merged work branches from this work or the user's earlier work in those repositories, with their clean worktrees, and SHALL keep forbidding the deletion of unmerged branches and of worktrees with uncommitted or untracked files. The deferral of cleanup to the close footer's reply SHALL be unchanged.

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
A revision of a written plan that has a change folder SHALL edit in the same pass every artifact that states what changed, mark a replaced decision or criterion with `**Replaces:** <old IDs>` and a dropped criterion with `**Retired:** <reason>` without deleting the old line, reopen each completed task the revision invalidates, and run the coherence check until it exits 0 before work continues; a revision that changes a criterion or the delta SHALL get one `hive-review-plan` round.

#### Scenario: A task still names a replaced decision
- **GIVEN** a change folder whose revision replaced a decision and marked the new one with `**Replaces:**`
- **WHEN** the coherence check runs and a task still names the replaced decision
- **THEN** the check prints that task's file and line and exits 1, and the revision is not complete until it exits 0

#### Scenario: A revision changes a criterion
- **GIVEN** a change folder whose plan was already reviewed
- **WHEN** a revision changes a criterion or the delta
- **THEN** one `hive-review-plan` child reviews the revised folder, and a gap that survives the round goes to the user as a decision

### Requirement: Archive confirms unchecked tasks
Before merging a change folder's deltas into the specs and moving it to the archive, the close SHALL run the coherence check with `--archive` and fix each coherence line it prints; when the check lists unchecked task boxes, the close SHALL show their count and ask whether to archive anyway before moving the folder, and a close under unattended delegation SHALL leave the folder unarchived and record it as pending.

#### Scenario: Unchecked boxes remain at archive
- **GIVEN** an integrated change folder whose `tasks.md` holds unchecked boxes
- **WHEN** the close runs the check with `--archive`
- **THEN** the close shows the number of unchecked boxes and asks before archiving, and the folder moves only on the user's answer

#### Scenario: An unattended close with unchecked boxes
- **GIVEN** a close running under unattended delegation and a folder with unchecked boxes
- **WHEN** the check lists them
- **THEN** the folder stays unarchived and is recorded as pending

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

### Requirement: Decisions carry labelled context

A decision `D<n>` shown to the user SHALL carry, under its title line, a `Situation:` line that says what happened or what was found, what was verified and what to check, with its target, in words the user has used or that the same message explains; under each option, an `If you choose it:` line that says what the user gets and what it costs them in time, money, risk or what stops working; and an `I recommend <letter>:` line whose reason is a consequence for the user and never a quotation of a rule or a skill. The three labels SHALL be written in the session language, while the IDs, letters and the `(★)` mark stay as they are. The round's waiting line SHALL also say how many known decisions are still to come. In a native question card, the question SHALL carry the situation, each option's description the text of its `If you choose it:` line, or the option's effect in one line when it has none, and the starred option's description SHALL end with the reason of the recommendation. The "Open the viewer" option that a validation question carries SHALL have no `If you choose it:` line. A finding `H<n>` asked in a round carries its title and option lines without the three labelled lines, and the replies of a next-step block keep their own format.

#### Scenario: A rule requires the question

- **GIVEN** a rule requires the agent to ask before an effect
- **WHEN** the agent asks it as a decision
- **THEN** the recommendation line states the effect that needs the answer, such as its cost or what it changes
- **AND** it quotes no rule or skill as its reason

#### Scenario: A name the user was not given

- **WHEN** the situation would name a task ID, a file, a branch or an internal name the user has not been given
- **THEN** the same message explains it, or the situation leaves it out

### Requirement: Reasons state effects, not rules

Every sentence shown to the user that gives the reason for an action taken, a step skipped, a question or a stop, in a decision and the text around it, a report, a stop, a blocker or a progress line, SHALL state that reason as its effect on the user or their work: what the step protects or what skipping it would cost. It SHALL NOT give as the reason that a rule, a skill, a flow or the guidance requires, allows or forbids it. A rule, a skill or the guidance SHALL be named only when the sentence is about it: the user asked about it, the work reads or changes it, or it is an instruction of the user's project that the user can change, named with its effect.

#### Scenario: A stop before a destructive step

- **GIVEN** a worktree holds uncommitted changes and a rule makes the agent ask before forcing its removal
- **WHEN** the agent stops to ask
- **THEN** the message says that forcing the removal would lose those changes
- **AND** it does not say that a rule or a skill makes it ask

#### Scenario: A report explains a step

- **WHEN** a report says why a review ran, a check was skipped or a fix was left for the close
- **THEN** the reason is what that step protects or what it would cost
- **AND** no sentence gives a rule, a skill, a flow or the guidance as the reason

#### Scenario: The rule is the subject

- **WHEN** the user asks which rule governs a behavior, or the work reads or changes that rule
- **THEN** the message names the rule

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

### Requirement: Screen examples in research

When the user asked for a change and what gets built includes a new screen, or one that adds, removes or moves a region, and neither the request nor an existing screen of the project fixes its composition, research SHALL ask the user once per work, as a decision, whether to design and present proposals for the screen; a change that keeps the screen's layout SHALL NOT get that offer, and neither SHALL work that runs under unattended delegation. The offer SHALL give text sketches as an option always, and frames in the project's design file as an option only when the app's design record declares that file or the user named it in the conversation, and the host has a connection that can write to it. The design-file option SHALL name the write to that file in its label or description, and the user's choice of it SHALL be the instruction for those writes; the offer SHALL NOT replace the question a critical effect takes. Research SHALL show two or three examples when the composition is open and one when it is fixed. In a design file the work SHALL create frames on a page named after the work and SHALL change or delete only the nodes it created; a request for the design that names a page or frames as its target SHALL be drawn there, and the work SHALL change, and never delete, the frames it names. The user SHALL choose or correct an example as a decision, approving or correcting it when there is only one, and SHALL approve the 440 px frame of a screen served on mobile before it is recorded; the approved example SHALL be recorded in the report and in a note of the work's support folder, for a frame with its location, its nodes and a local image and for a text sketch with the sketch itself, and the examples not chosen SHALL be removed. A plan with a change folder SHALL carry that record in its `UI` section, and work without one in its checkpoint file. A request whose deliverable is the design SHALL be drawn the same way without the offer, SHALL NOT be reported as done before the user approves the example, and after that approval research SHALL run the close, which treats it as an operational task. Under unattended delegation the choice or the approval of an example SHALL stay unanswered and be reported as unresolved, every example SHALL be kept, and the plan SHALL settle the UI without an approved example. A plan for the screen SHALL use the approved example as the after sketch of the regions it shows, and the build's first view, the verification walk and the UX review SHALL compare the rendered screen against it. A difference SHALL be a region, a component, an action's placement or an order that differs from the example, not rendering variance, and not what the plan records as decided otherwise or a correction the user asked for after approving the example, which the example's record SHALL name; it SHALL be handled as a UI defect the change introduced, and the differences that remain SHALL be listed in the validation message when the delivery stops for validation and in the final report when it does not. When a frame's image is missing where the comparison runs, the frame SHALL be read again from its recorded location and node, and a comparison that cannot run SHALL be reported as not run. The user's validation in verification SHALL be asked as before.

#### Scenario: Declared design file and a writing connection

- **GIVEN** an app whose design record declares a design file, and a host whose connection can write to it
- **WHEN** research reaches a new screen whose composition is open
- **THEN** it asks whether to design proposals, with the design file, text sketches and no proposals as options
- **AND** on the design-file answer it shows two or three examples as frames on a page of that file named after the work

#### Scenario: No design tool

- **GIVEN** an app with no declared design file, or a host with no connection that can write to it
- **WHEN** research reaches a new screen whose composition is open
- **THEN** the offer gives text sketches and no proposals as its options

#### Scenario: A change that keeps the layout

- **GIVEN** a request to change a button, a copy, a field or a column of an existing screen
- **WHEN** research settles what the work is
- **THEN** it makes no offer of proposals

#### Scenario: The request is the design

- **GIVEN** a request to draw or replicate a screen in a design file
- **WHEN** research draws the example
- **THEN** it makes no offer and asks the user to approve or correct the example before reporting the work as done
- **AND** after the approval it runs the close

#### Scenario: The built screen differs from the approved example

- **GIVEN** work whose record holds an approved example
- **WHEN** the verification walk captures a region the example shows
- **THEN** each difference in a region, a component, an action's placement or an order is handled as a UI defect the change introduced
- **AND** the differences that remain are listed in the validation message, or in the final report when the delivery does not stop for validation

#### Scenario: Unattended delegation is active

- **GIVEN** work that runs under unattended delegation
- **WHEN** research reaches a new screen whose composition is open
- **THEN** it makes no offer of proposals and the plan settles the UI
