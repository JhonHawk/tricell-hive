## ADDED Requirements

### Requirement: The close checks that the change folder is versioned

At the completion point of work not delivered under `hold`, `flow-close` SHALL run `python3 -I flow-plan/scripts/tasks.py <change-folder> versioned --work-repo <work repository>` after the change folder's commit that `git-workflow` step 4 sets and before its final report. The subcommand SHALL resolve the folder's repository from the folder itself, through `git -C <folder> rev-parse --show-toplevel`, so that it gives the same answer from a repository, a subfolder of a monorepo or the root of a workspace that is not a Git repository; it SHALL read `Base branch` and `Delivery` from that repository's `## Hive` section, taking the base from `origin/HEAD` when the section does not declare it; and it SHALL only read, with no fetch, commit or push, and SHALL print no delivery commands. It SHALL exit 0 when the folder's last commit is contained in a remote-tracking branch, in `origin/<base>` when that repository declares `Delivery: direct-base`, and, when the folder's repository is not the work repository, no file under the folder is untracked or modified; in the work repository, a file untracked or modified after that commit SHALL only be warned about, since `git-workflow` step 4 leaves later edits there to the hygiene batch's archive. It SHALL exit 1, naming each file and what is missing, when the folder is not versioned, was never committed, is in no Git repository, or declares `Delivery: direct-base` with no resolvable base; and it SHALL exit 2 on invalid arguments. It SHALL warn, without changing its exit status, when the work repository's `## Hive` declares no `Specs`, or one whose path, taken before ` · ` and resolved against that repository, does not hold the folder. On exit 1 the close SHALL commit and deliver the folder as `git-workflow` step 4 sets and run the check again; when it still exits 1, the final report SHALL list the folder as pending with its path and its question, asked after the report as the close modes set, and the ticket's move SHALL NOT change.

#### Scenario: A sibling specs repository left untracked

- **GIVEN** a work merged into its code repository whose change folder lives, untracked, in a sibling specs repository that declares `Delivery: direct-base`
- **WHEN** the close runs `versioned` on the folder from the workspace root
- **THEN** it exits 1 and names the untracked files, and the close commits and pushes the folder as `git-workflow` step 4 sets before running the check again

#### Scenario: A folder in the work repository with a later checkpoint edit

- **GIVEN** a change folder at a monorepo's root, committed on the work branch and merged, whose `tasks.md` checkpoint was edited after that commit
- **WHEN** `versioned` runs with the folder's path from a package directory and `--work-repo` naming the monorepo
- **THEN** it exits 0, names the monorepo's root as the repository and warns that the edit waits for the archive

#### Scenario: The code repository does not declare Specs

- **GIVEN** a change folder committed and pushed in a specs repository, and a work repository whose `## Hive` has no `Specs`
- **WHEN** `versioned` runs with `--work-repo` naming that work repository
- **THEN** it exits 0 and warns that the work repository should declare `Specs` with the folder's path

### Requirement: The session audit fails a close that left the change folder uncommitted

In `go run ./tests/pilot --audit-session`, `change_folder_uncommitted` SHALL read commands from their resolved form, which for Codex is the completed command item with its exit code, falling back to the shell call when its success is known and no completed item exists, SHALL take as the change folder each `changes/<id>` segment outside `archive/` of a path the session wrote through a patch, a write tool or a command of a known writing form, never through a read, and SHALL count a merge whether or not it deletes the branch. It SHALL report pass when, after the first write to the folder, a successful commit command names that segment or a `tasks.py <folder> versioned` on it exits 0; fail, naming the folder, when assistant text follows the last merge and neither happened; review when no assistant text follows the last merge, or when the session's only writes to the folder are `tasks.py` progress or checkpoint records after the last merge; and not_observed when the session wrote no change folder or merged nothing.

#### Scenario: The CAM-20 close

- **GIVEN** the real Codex session `01a1f5d4` that merged PR #384 through an interpolated command and closed with «No quedan pendientes de esta entrega» without committing `cam-specs/openspec/changes/cam-20-profile-save`
- **WHEN** the session audit reads it
- **THEN** `change_folder_uncommitted` reports fail and names `changes/cam-20-profile-save`

#### Scenario: A folder committed before the merge

- **GIVEN** a session that committed its change folder on the work branch, naming its path, and then merged the pull request
- **WHEN** the session audit reads it
- **THEN** `change_folder_uncommitted` reports pass
