## ADDED Requirements

### Requirement: A completion-point close cleans up before its report

At the completion point, and when an operational task ends, `flow-close` SHALL run its steps in order before its final report, the branch cleanup included, without waiting for a reply: it SHALL delete the merged work branch and its clean worktree, sweep the user's earlier merged branches as the branch cleanup sets, stop the processes the work started and remove its reproducible temporary files. Only an item that needs the user SHALL stay pending, each with its question, none of them asked before the final report: what the cleanup helper keeps for a reason the user must resolve, a dump or export the work made, and each remote resource it created; the footer's `Close` reply SHALL exist only while such an item is pending. A promotion that the retention criteria propose, and a closed `sessions/` folder, SHALL be named in the report's next steps with the `project-layout` command and SHALL NOT count as pending. A user message that ends the session without asking for more work SHALL run the cleanup again before the agent's last message, which SHALL list each item still pending with its question, each kept branch in the form ``kept `<branch>`: <reason>``, and name the change folders waiting for a hygiene batch without delivering them. A hygiene request after a work item's close SHALL run only what is still pending and the hygiene batch, SHALL report only those, and SHALL NOT write the work's final report again.

#### Scenario: A merged fix with captures already on its ticket

- **GIVEN** a work item merged into the base, with its captures attached to its ticket and read back at the close's first step, in a clean worktree on the merged branch
- **WHEN** the close runs at the completion point
- **THEN** the captures, the worktree and both copies of the branch are removed before the final report
- **AND** the footer offers no `Close` reply

#### Scenario: The user ends the session with a step pending

- **GIVEN** a close that kept a worktree because it holds material the helper cannot remove
- **WHEN** the user writes «es todo»
- **THEN** the agent runs the cleanup again before its last message, lists the kept worktree with its reason and question, and names the pending archives without opening a pull request for them

### Requirement: The cleanup helper reports clones it cannot remove

`flow-close/scripts/cleanup_branches.py` SHALL list under `stray_clones`, in its plan with and without `--apply`, each direct subfolder of the main checkout's `.worktrees/`, `.claude/worktrees/` and `.grok/worktrees/` folders and of the main checkout's parent folder's `.worktrees/` that holds its own `.git` directory and whose `origin` URL, compared by host and path without scheme, user or `.git` suffix, equals the repository's, with its path, its branch, whether it holds changed or untracked files, and how many of its commits are on no remote. It SHALL NOT remove, move or switch any of them.

#### Scenario: A clone left by an unanswered worktree question

- **GIVEN** a workspace whose `.worktrees/fsr-905-906-frontend` is a full clone of the frontend repository over SSH while the repository's `origin` is HTTPS
- **WHEN** the helper plans the frontend repository's cleanup
- **THEN** `stray_clones` lists that folder with its branch and its unpushed commit count, and the folder is still on disk after `--apply`

### Requirement: The session audit fails a close report that precedes the cleanup

In `go run ./tests/pilot --audit-session`, `close_before_branch_cleanup` SHALL report fail when an assistant close report comes after a successful merge and before the first command that deletes a branch, counting `cleanup_branches.py --apply` as such a command however the script is invoked, and SHALL report review instead only when that report holds the form ``kept `<branch>` `` naming a branch the session deleted, as `--merged-pr`, `git branch -d` or `-D`, or `git push --delete` show it.

#### Scenario: A Grok close that reported first

- **GIVEN** a real session whose close report follows the merge and whose branch deletion comes after the user's `/flow-close`
- **WHEN** the session audit reads it
- **THEN** `close_before_branch_cleanup` reports fail, even though the report says a folder «sigue en `master`»

## MODIFIED Requirements

### Requirement: Viewer finalization before deferred hygiene

The normal completion phase SHALL publish and read back the viewer's actual terminal or waiting lifecycle before the report's hygiene footer. Integrated or abandoned completed work SHALL close explicitly; a ready/human-review stop SHALL remain in_review, a real pause SHALL remain paused, and awaiting validation SHALL remain explicit. Change-record archival, and an item the close lists as pending for the user, MAY remain deferred without keeping a completed viewer record active. A later hygiene phase SHALL NOT duplicate an already confirmed closure. The procedure SHALL preserve the change-folder commit and intentional-pending-work rules of the prerequisite #311 candidate.

An unavailable viewer or missing producer authority SHALL preserve inputs, report the still-recorded state and continue independent work without starting the service, silently recovering authority or claiming a successful close. Upload cleanup SHALL retain its exact committed-readback requirement.

#### Scenario: Delivered work awaits optional cleanup
- WHEN authorized work reaches its completion point and the viewer accepts its close
- THEN the initial report observes the work closed while the hygiene footer can still wait for the user.

#### Scenario: An open pull request waits for merge
- WHEN delivery stops ready for the user's merge
- THEN the viewer remains in_review and is not made eligible for terminal autoarchive.

#### Scenario: Later hygiene
- WHEN the user requests the deferred hygiene
- THEN its archive and pending steps use the previous closure evidence without appending another closure.

### Requirement: Close cleanup sweeps the repositories the work touched
The close's branch cleanup SHALL run in each repository where the work created a branch, a worktree or a commit or wrote its change folder, plus the repository that holds `Specs`, using `flow-close/scripts/cleanup_branches.py` when `python3` runs it to compute the dry-run cleanup plan before applying deletions with `--apply`; a hygiene request SHALL cover the current repository, its `Specs` repository and the repositories that the pending-close records in persistent memory name. Every worktree removal SHALL keep the worktree and both copies of its branch when it has uncommitted or untracked files, or an ignored entry other than these three: a dependency or build folder or an entry in one; an entry whose content the main checkout of the worktree's own repository already holds at the same path; and, at any depth, `*.tsbuildinfo`, `next-env.d.ts`, `*.pyc`, `.turbo/`, `__pycache__/`, `.pytest_cache/`, `.mypy_cache/`, `.ruff_cache/` or `.DS_Store`, or an entry inside one of those folders; it SHALL first stop the processes the work started inside its own worktree, by process ID, and SHALL keep any worktree where a running process has its working directory or where that cannot be checked. A squash- or rebase-merged branch SHALL count as merged only when its pull request's `headRefOid` equals the head of each local and remote copy to be deleted; a remote copy that moved SHALL be reported with the commits after that head. After the work's own branch, the sweep SHALL check the other branches whose pull requests the user authored and merged into the base, the local branches `git branch --merged` lists that a close record in persistent memory, or a plan or checkpoint, names for the user's earlier work in a repository without pull requests, and the temporary or preview branches the work created. It SHALL NOT switch any checkout's branch, and SHALL keep, reporting the reason, a branch checked out in another checkout and a branch that persistent memory, a plan or a checkpoint names for work whose close is not recorded as done, unless a pending-close record for that work is older than 24 hours, in which case it SHALL resume that work's pending close steps in order instead of removing its branch or worktree. The global guidance SHALL authorize deleting merged work branches from this work or the user's earlier work in those repositories, with their clean worktrees, and SHALL keep forbidding the deletion of unmerged branches and of worktrees with uncommitted or untracked files. At the completion point the cleanup SHALL run before the final report, as «A completion-point close cleans up before its report» sets.

#### Scenario: A leftover worktree of earlier work
- **WHEN** a close or hygiene request runs and the specs repository holds a clean worktree whose branch the user merged into the base in earlier work whose pending close is older than 24 hours
- **THEN** that work's pending close steps run in order and its worktree and branch are removed by its own cleanup

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

### Requirement: A ticket is documented with evidence before cleanup

At a work item's close that removes material, the agent SHALL document its ticket with the evidence the work produced, such as captures and test output, before any step that removes support material, temporary files or worktrees; a close of paused or `hold` work SHALL publish none. Verification material whose attachment the ticket's attachment list shows when that step reads it back, or whose uploaded URL is in the merged pull request's body, SHALL count as replaced and SHALL be deleted without a question, and the global guidance SHALL authorize that deletion at a work item's close; material only listed in a comment for the user to attach, or that reached neither the ticket nor the body, SHALL be kept and named in the report. The close SHALL NOT commit or archive support material: what it keeps SHALL stay unversioned, moved from a worktree to the main checkout before the worktree is removed.

#### Scenario: Captures before cleanup

- **GIVEN** a closed UI fix with before and after captures in its workspace
- **WHEN** the close runs
- **THEN** the captures reach the ticket before the cleanup removes them

#### Scenario: An attachment that failed

- **GIVEN** a close whose upload of a video to the ticket failed
- **WHEN** the close reaches its cleanup
- **THEN** the video stays in the work's workspace folder and the report names it

### Requirement: Archives are delivered in a hygiene batch
When `Specs` is in the work's repository, the close of an integrated work SHALL NOT archive its change folder; it SHALL record it in persistent memory, one record per work under the topic `archive-pending/<id>`, where the id is the change-id, else the ticket key or a short slug, with the folder and the work's pull request and merge commit, SHALL discard with `git restore` the folder's uncommitted edits left after the merge, naming those that touch more than its checkpoint, and its report and footer SHALL name how many archives are pending and the hygiene reply that delivers them. The hygiene batch SHALL archive change folders only and SHALL NOT commit support material. When `Specs` is in another repository, the close SHALL keep archiving there as that repository's `Delivery` sets. A hygiene request SHALL treat as pending each change folder still under `<Specs>/changes/` on the base that has a record, or whose work the user confirms integrated, skipping one that an open hygiene pull request already archives; SHALL archive them on one branch from the base in the order each reached the base; and SHALL deliver it once under the delivery mode in force for that repository, asking only the delivery question when none is in force, leaving everything pending under `hold` and pushing to the base under `direct-base`. It SHALL remove from each record what its delivery carried once that delivery reaches the base, soft-delete a record left with nothing, and leave pending each folder whose delta changes a requirement that a folder left pending adds or changes. A folder without a record SHALL stay pending under unattended delegation.

#### Scenario: A close leaves its archive pending
- **GIVEN** a work integrated into the base with a change folder
- **WHEN** the close runs
- **THEN** no branch or pull request is opened for it, the record `archive-pending/<change-id>` lists it, and the report names the pending count and the hygiene reply

#### Scenario: A hygiene request delivers two pending archives
- **GIVEN** two change folders still under `<Specs>/changes/` on the base, the second written while the first was pending
- **WHEN** the user asks for hygiene
- **THEN** one branch archives the first and then the second, one delivery carries them, and their records are emptied once the archive reaches the base

### Requirement: The session audit fails a merged branch left only after a close

In `go run ./tests/pilot --audit-session`, `merged_branch_deleted` SHALL fail when an assistant close report follows the session's last merge and no branch deletion follows that merge, and SHALL also fail when a user message that ends the session follows that merge and no branch deletion follows the message, and SHALL report review instead when that merge is a promotion, whose head the session shows, through `gh pr create --head` or a read of `headRefName`, to be the base or a branch of the `Environments` the session's `AGENTS.md` declares, or when the last assistant message holds ``kept `<branch>` `` for that merge's branch; a merge whose head the session does not show SHALL count as a work merge. A message ends the session when it starts with one of the phrases the audit lists, such as «es todo» or "that's all", has at most 10 words and holds no question mark. When a merge has no later deletion and neither a close report nor such a message follows it, the criterion SHALL report review and SHALL name that the session may still be open or the merge may be a promotion. The pilot criterion of the same name SHALL keep its behavior.

#### Scenario: A session still promoting

- **GIVEN** a real session whose last merge is the promotion pull request from `development` to `qa`, with no close report after it
- **WHEN** the session audit reads it
- **THEN** `merged_branch_deleted` reports review, not fail

#### Scenario: A promotion before the end of the session

- **GIVEN** a session that deleted its work branch, then merged the promotion from `development` to `qa`, and whose user then wrote «es todo»
- **WHEN** the session audit reads it
- **THEN** `merged_branch_deleted` reports review, not fail

#### Scenario: A session ended after the merge

- **GIVEN** a real Codex session that merged a work branch and whose user then wrote «es todo por el momento, gracias!», with no branch deletion after it
- **WHEN** the session audit reads it
- **THEN** `merged_branch_deleted` reports fail

### Requirement: Pull requests carry before and after captures of changed screens

Before its first edit of anything an existing screen renders, a build SHALL have its main thread serve the unchanged checkout and capture each region its plan, or the request when it has no plan, changes, with the widths, themes and states the verification will check, before any child task edits that surface; those captures SHALL NOT be published to the viewer, and the checkpoint SHALL record their paths or their absence. When the verification produced captures, the remote is github.com and the installed `gh` offers `--attach`, the pull request body SHALL hold a before and after table per region, width, theme and state, with the images uploaded through `--attach`, none added to a commit, and a cell that names a new screen or a missing before capture. In a public repository or on GitHub Enterprise Server, it SHALL attach only captures that show seeded data and test accounts, without internal URLs or real email addresses, and only with the user's one-time authorization for that work. Otherwise the body SHALL keep its text-only form and name the captures as kept locally. At the close, a capture SHALL be treated as replaced when its uploaded URL was read back in the merged body or the ticket's attachment list shows it, as «A ticket is documented with evidence before cleanup» sets; other captures SHALL be kept.

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
- **THEN** the body keeps its text-only form, names the captures as kept locally, and the close keeps those its ticket does not show

#### Scenario: Public repository without authorization
- **WHEN** the repository is public and the user has not authorized the public upload for this work
- **THEN** the captures stay local, the body keeps its text-only form, and the close keeps those its ticket does not show

#### Scenario: Partial upload
- **WHEN** some attachments fail while the pull request is created
- **THEN** the agent attaches the missing ones with `gh pr edit --attach` and edits the body only from the published body

#### Scenario: Close after the merge
- **WHEN** the work is integrated and the agent read back each uploaded image URL in the merged body
- **THEN** the close deletes those local captures without a question and keeps any capture that neither a read-back URL nor the ticket's attachment list shows
