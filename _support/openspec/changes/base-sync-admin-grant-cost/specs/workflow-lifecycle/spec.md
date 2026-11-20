## ADDED Requirements

### Requirement: Close cleanup brings the local base up to date

With `--apply`, `flow-close/scripts/cleanup_branches.py` SHALL fetch the base from `origin` and, before planning the branch cleanup, SHALL fast-forward the base in the main checkout, even when `--repo` names a linked worktree, with `git merge --ff-only --no-overwrite-ignore origin/<base>` only when the main checkout is on the base, the local base has no commits missing from `origin/<base>`, and no file that changes between `HEAD` and `origin/<base>` is modified, untracked or ignored in the main checkout. Otherwise it SHALL leave the base where it was and report the reason and the files involved. It SHALL report the base's action in its plan in both modes, SHALL NOT switch any checkout's branch, and SHALL NOT fetch or move anything without `--apply`.

#### Scenario: A base left behind by the merge
- **GIVEN** a main checkout on `master`, 20 commits behind `origin/master`, with an unrelated file modified locally
- **WHEN** the close runs the cleanup with `--apply`
- **THEN** `master` moves to the tip of `origin/master`, the modified file stays as it was, and the report shows `fast_forward` with 20 commits

#### Scenario: Incoming changes touch local work
- **GIVEN** a main checkout on the base where a file that `origin/<base>` changes is modified locally
- **WHEN** the close runs the cleanup with `--apply`
- **THEN** the base does not move and the report shows `keep` naming that file

### Requirement: An admin-merge grant lasts the session

The user's explicit grant to admin-merge a pull request SHALL hold, unless the user withdraws it, for the session's later pull requests in the repositories of its work whose required checks passed or whose base requires none, a required check that never reported not being none, and SHALL be named in each report it covers. It SHALL lift the hold of a human reviewer whose latest review approves an earlier commit, naming that reviewer and commit. It SHALL NOT cover a merge into the production branch, which is the last branch of the repository's `Environments` chain when that branch is not the base, a direct push, a deviation from the promotion path, a requested reviewer even one who approved an earlier commit, or a review that requests changes. The global guidance SHALL NOT state that a repeated approval or a reviewer's pending approval forbids what such a grant covers.

#### Scenario: A second admin merge in the same session
- **GIVEN** the user granted an admin merge for a pull request of the backend repository
- **WHEN** the close needs to merge the change record's pull request in the specs repository, whose required checks passed and whose ruleset requires a review
- **THEN** the agent merges it with `--admin` without asking again and names the grant in its report

#### Scenario: A merge into production
- **GIVEN** an admin-merge grant in force in the session
- **WHEN** a pull request into the production branch needs a bypass
- **THEN** the grant does not cover it, and the merge gate's admin exception and the promotion path still govern that merge

### Requirement: The session audit reports the main thread's cost

`go run ./tests/pilot --audit-session` SHALL report, for Claude, Codex and Grok sessions, an informational `main_thread_cost` block with the main thread's model calls, processed input tokens and their cached part, compactions, tool calls, reads of installed Hive skill files, runs of Hive skill scripts and tool calls by tool, without its children's usage, and SHALL NOT change any criterion's verdict. When a host reports the parent's usage with its children's and a child's usage cannot be read, the block SHALL say so in `usage_source`.

#### Scenario: A Grok session with children
- **WHEN** the audit reads a Grok session whose turns report 599 model calls and 70,803,744 input tokens, and whose 12 children report 189 calls and 15,688,019 input tokens
- **THEN** `main_thread_cost` shows 410 model calls, 55,115,725 input tokens and `usage_source` `per_turn`
