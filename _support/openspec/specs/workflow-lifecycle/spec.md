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
The close's branch cleanup SHALL run in each repository where the work created a branch, a worktree or a commit or wrote its change folder, plus the repository that holds `Specs`; a hygiene request SHALL cover the current repository, its `Specs` repository and the repositories that the pending-close records in persistent memory name. Every worktree removal SHALL keep the worktree and both copies of its branch when it has uncommitted or untracked files, or an ignored entry that is not a dependency or build folder or in one, whose content the main checkout of the worktree's own repository does not already hold at the same path, and that is not a TypeScript or Next.js generated file (`*.tsbuildinfo`, `next-env.d.ts`); it SHALL first stop the processes the work started inside its own worktree, by process ID, and SHALL keep any worktree where a running process has its working directory or where that cannot be checked. A squash- or rebase-merged branch SHALL count as merged only when its pull request's `headRefOid` equals the head of each local and remote copy to be deleted; a remote copy that moved SHALL be reported with the commits after that head. After the work's own branch, the sweep SHALL check the other branches whose pull requests the user authored and merged into the base, the local branches `git branch --merged` lists that persistent memory records for the user's earlier work in a repository without pull requests, and the temporary or preview branches the work created. It SHALL NOT switch any checkout's branch, and SHALL keep, reporting the reason, a branch checked out in another checkout and a branch that persistent memory records for work whose close is not recorded as done, unless a pending-close record for that work is older than 24 hours, in which case it SHALL resume that work's pending close steps in order instead of removing its branch or worktree. The global guidance SHALL authorize deleting merged work branches from this work or the user's earlier work in those repositories, with their clean worktrees, and SHALL keep forbidding the deletion of unmerged branches and of worktrees with uncommitted or untracked files. The deferral of cleanup to the close footer's reply SHALL be unchanged.

#### Scenario: A leftover worktree of earlier work
- **WHEN** a close or hygiene request runs and the specs repository holds a clean worktree whose branch the user merged into the base in earlier work whose pending close is older than 24 hours
- **THEN** that work's pending close steps run in order, including its support decision, and its worktree and branch are removed by its own cleanup

#### Scenario: A worktree still in use
- **WHEN** a merged branch's worktree is the working directory of a running process, holds an ignored entry that is not in a dependency or build folder, not already held by the main checkout at the same path and not a TypeScript or Next.js generated file, or belongs to work in progress or whose close has been pending for less than 24 hours
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
