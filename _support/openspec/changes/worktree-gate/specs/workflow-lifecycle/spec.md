## ADDED Requirements

### Requirement: The user decides whether a worktree is created

When the work is already routed to build and authorized and the user asked for a worktree, the agent SHALL create it; a question from the user about using a worktree is not such a request. When the current checkout is on the base and holds no changes the agent did not make, the agent SHALL create the work branch there. Otherwise, before creating any branch or switching, the agent SHALL ask the user as a decision whether to create a worktree, stating what it observed, such as another session's changes or another branch checked out, unless the session already works in one. An accepted worktree SHALL go in the host's native worktree location; when the host has none, or it is not verified, the agent SHALL ask for the path. An accepted worktree inside the repository SHALL be excluded through the exclude file under `git rev-parse --git-common-dir`. The work's checkpoint, in its change folder or its temporary checkpoint file, SHALL name the worktree's path and branch, and work without a checkpoint SHALL find it through `git worktree list`; the agent SHALL NOT write persistent memory because a worktree was created or removed. A delegated child isolated in its own worktree is a worktree the agent creates, and a delegated child SHALL NOT create a worktree: it compares with the base through `git diff` or `git show <base>:<path>`. Under unattended delegation the agent SHALL NOT ask: it SHALL decide at that moment whether to create the worktree, choosing the option that keeps the work going, and SHALL name its choice and reason in its report.

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

The agent SHALL give each command that acts on one repository, such as a package manager, build, test or Git command, the repository or worktree that holds the change by an absolute path, through the tool's own directory option, such as `git -C`, `pnpm --dir` or `go -C`, and only for a tool without one through `cd <absolute path> && <command>` in the same call, where the host allows it. It SHALL NOT rely on a `cd` made in an earlier call, and SHALL NOT run the command from the root of a declared workspace or after a chain of relative `cd ..`. A change that spans several repositories SHALL run each command in its own repository. Reads that span the workspace and operational tasks that change no repository are outside this requirement.

#### Scenario: A package manager command in a workspace
- **WHEN** a session started at a declared workspace root that holds several repositories runs `pnpm install` for one of them
- **THEN** it runs the command with that repository's or worktree's absolute path as working directory, and no `package.json`, lockfile or `node_modules` appears at the workspace root.

## MODIFIED Requirements

### Requirement: Close cleanup sweeps the repositories the work touched
The close's branch cleanup SHALL run in each repository where the work created a branch, a worktree or a commit or wrote its change folder, plus the repository that holds `Specs`; a hygiene request SHALL cover the current repository, its `Specs` repository and the repositories that the pending-close records in persistent memory name. Every worktree removal SHALL keep the worktree and both copies of its branch when it has uncommitted or untracked files, or an ignored entry other than these three: a dependency or build folder or an entry in one; an entry whose content the main checkout of the worktree's own repository already holds at the same path; and, at any depth, `*.tsbuildinfo`, `next-env.d.ts`, `*.pyc`, `.turbo/`, `__pycache__/`, `.pytest_cache/`, `.mypy_cache/`, `.ruff_cache/` or `.DS_Store`, or an entry inside one of those folders; it SHALL first stop the processes the work started inside its own worktree, by process ID, and SHALL keep any worktree where a running process has its working directory or where that cannot be checked. A squash- or rebase-merged branch SHALL count as merged only when its pull request's `headRefOid` equals the head of each local and remote copy to be deleted; a remote copy that moved SHALL be reported with the commits after that head. After the work's own branch, the sweep SHALL check the other branches whose pull requests the user authored and merged into the base, the local branches `git branch --merged` lists that a close record in persistent memory, or a plan or checkpoint, names for the user's earlier work in a repository without pull requests, and the temporary or preview branches the work created. It SHALL NOT switch any checkout's branch, and SHALL keep, reporting the reason, a branch checked out in another checkout and a branch that persistent memory, a plan or a checkpoint names for work whose close is not recorded as done, unless a pending-close record for that work is older than 24 hours, in which case it SHALL resume that work's pending close steps in order instead of removing its branch or worktree. The global guidance SHALL authorize deleting merged work branches from this work or the user's earlier work in those repositories, with their clean worktrees, and SHALL keep forbidding the deletion of unmerged branches and of worktrees with uncommitted or untracked files. The deferral of cleanup to the close footer's reply SHALL be unchanged.

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
