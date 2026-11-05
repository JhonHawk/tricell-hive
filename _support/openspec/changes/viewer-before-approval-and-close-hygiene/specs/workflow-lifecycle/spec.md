## ADDED Requirements
### Requirement: Close cleanup sweeps the repositories the work touched
The close's branch cleanup SHALL run in each repository where the work created a branch, a worktree or a commit or wrote its change folder, plus the repository that holds `Specs`; a hygiene request SHALL cover the current repository, its `Specs` repository and the repositories that the pending-close records in persistent memory name. In each, after the work's own branch, it SHALL apply the same merge checks to the other branches merged into the base whose pull requests the authenticated user authored with a `headRefOid` equal to the local head, to local branches `git branch --merged` lists that hold at least one commit of their own now in the base in a repository without pull requests, and to the temporary or preview branches the work created, removing their clean worktrees before deleting them. The sweep SHALL NOT switch any checkout's branch. It SHALL keep, and report with its reason, a branch with unmerged commits or with no commits of its own beyond the base, a branch checked out in a checkout other than a clean worktree it removes, a worktree with uncommitted, untracked, or ignored files outside dependency and build folders, a worktree where a running process has its working directory or where that cannot be checked, and a branch or worktree that persistent memory records for work whose close has been pending for less than 24 hours. The global guidance SHALL authorize deleting merged work branches from this work or the user's earlier work in those repositories, with their clean worktrees, and SHALL keep forbidding the deletion of unmerged branches and of worktrees with uncommitted or untracked files. The deferral of cleanup to the close footer's reply SHALL be unchanged.

#### Scenario: A leftover worktree of earlier work
- **WHEN** a close or hygiene request runs and the specs repository holds a clean worktree whose branch the user merged into the base in earlier work whose pending close is older than 24 hours
- **THEN** the cleanup removes that worktree and its branch, updates that work's pending record and reports them

#### Scenario: A worktree still in use
- **WHEN** a merged branch's clean worktree is the working directory of a running process, holds ignored files outside dependency and build folders, or belongs to work whose close has been pending for less than 24 hours
- **THEN** the cleanup keeps it and reports the reason

#### Scenario: A freshly created branch
- **WHEN** a branch has no commits of its own beyond the base
- **THEN** the sweep keeps it even though `git branch --merged` lists it

### Requirement: Specs in a direct-base repository
When `Specs` lives in another repository that declares `Delivery: direct-base`, the change folder SHALL be edited in that repository's main checkout on its base, with no branch or worktree, even when another session writes there; only the change folder's paths SHALL be committed, the commits on the base that are not this work's SHALL be named and asked about before the push, and a rejected push SHALL stop for the user's decision without rebasing over uncommitted changes.

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
A promotion that has not been made SHALL NOT be a reason to keep a branch, worktree, process or support material. Only a task of the plan that promotes to an environment where the ticket or the request asks for the change SHALL be recorded as pending work. When the ticket or the request asks for the change in a named environment, the plan SHALL include that promotion as a task run through `flow-deploy`, and that request SHALL be the deploy instruction for that environment only, with production keeping its own question; otherwise promotions SHALL be left to be grouped and run when the user asks. Each project's `Completion` value SHALL keep its meaning.

#### Scenario: Work integrated before its promotion
- **WHEN** work is merged into the base, its promotion to a later environment has not run, and no plan task promotes to it
- **THEN** its cleanup keeps nothing for the promotion and the promotion is not listed as pending

#### Scenario: A ticket that asks for an environment
- **WHEN** the ticket asks for the change in a named environment
- **THEN** the plan includes the promotion to that environment as a task
