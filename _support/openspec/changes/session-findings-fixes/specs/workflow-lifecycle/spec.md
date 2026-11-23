## ADDED Requirements

### Requirement: Checks are scoped by the project's path map

When the project guidance assigns a check to each path a change touches, that assignment SHALL be the known scope of the change's checks. A round of fixes whose paths are all assigned SHALL run only the assigned checks and SHALL reuse the candidate's passing full-suite run; the full suite SHALL run again only for a path with no assigned check, or when the project guidance requires it for that kind of change. A plan task's `Verification:` line SHALL name the check assigned to the task's paths when there is one, not the full suite.

#### Scenario: A Markdown-only fix after the build's suite

- **GIVEN** a candidate whose full suite passed at the end of the build, and project guidance that assigns `go test ./tests/content/...` to text under `content/`
- **WHEN** a review fix edits only a leaf under `content/`
- **THEN** the agent runs the assigned check and does not run the full suite again

#### Scenario: A plan task on mapped paths

- **GIVEN** a plan task whose files are all assigned a check
- **WHEN** the plan writes its `Verification:` line
- **THEN** the line names the assigned check, not the full suite

### Requirement: A second review runs only for fixes that change code or a contract

After the round of fixes, a second code review SHALL run only when a hunk of the fixes' diff lies outside a `.md` or test file and changes more than comment lines, or when the fixes change a contract. A fix whose every hunk lies in a `.md` or test file or changes only comment lines, and that changes no contract, SHALL NOT start a review and SHALL carry an `Unreviewed: <reason>` line in its commit message. The artifact of an agent code review SHALL exclude the files `git diff -M --name-status` reports as `R100`, and its brief SHALL name them in one line; a native review mechanism that accepts paths SHALL get every changed path except those, added files and files renamed with changes included.

#### Scenario: A wording fix

- **GIVEN** a blocking finding fixed by editing only a Markdown leaf
- **WHEN** the round of fixes ends
- **THEN** no second review runs, and the fix's commit carries an `Unreviewed:` line

#### Scenario: A refactor that moves files

- **GIVEN** a diff that moves 180 files unchanged and edits 6 files
- **WHEN** an agent code review is launched
- **THEN** its artifact holds the 6 edited files, and the brief names the moved ones in one line

### Requirement: The session audit flags full-suite reruns after text fixes and extra review rounds

`go run ./tests/pilot --audit-session` SHALL recognize an unfiltered suite run through the session's shell segments, not text inside arguments or heredocs: `go test ./...` and `go -C <dir> test ./...` with `-count` or `-race`, `npm test`, `pnpm test` and `pytest` without arguments, while `-run` or a sub-package pattern filters it. It SHALL report `full_suite_after_text_fix`, for two runs in the same working directory with no `git push`, `gh pr merge`, new branch or new worktree between them, as `fail` when the second follows a first run shown to pass, by an echo of that run's own `$?` or `${PIPESTATUS[0]}` showing 0 in its output, or by the success of a command run without a pipe or background, a background launch or a run with no result never counting as shown, and everything changed between them is `.md`, `.mdx` or `.txt` files, by edit tools or by shell writes whose targets are of those types; as `pass` when another edit, or a rebase, merge, pull, checkout, switch, cherry-pick, am, apply, reset or restore, lies between them; as `review` when a shell write has a target that cannot be shown to be text or the first run cannot be shown to pass; and as `not_observed` without two runs, never counting `git add`, `git commit` or `git push` as a change. It SHALL report `review_rounds_per_candidate` for main Claude Code sessions, counting code review rounds (the `code-review` skill and `hive-review-code` children, the launches of one message being one round) and resetting at `git push`, `gh pr merge`, a new load of `flow-build` or `flow-plan` or a branch switch: `fail` with more than two rounds, `review` when the third follows a `/code-review` the user typed, `pass` with two or fewer, and `not_observed` without reviews, in child sessions and on other hosts.

#### Scenario: Two suites around a text fix

- **GIVEN** a session that runs `go test ./...` with a passing result, edits one `.md` file and runs `go test ./...` again in the same directory
- **WHEN** the audit reads it
- **THEN** `full_suite_after_text_fix` is `fail` and names the second run

#### Scenario: A shell write between the runs

- **GIVEN** a session that runs `go test ./...`, changes a Go test through a shell command and runs `go test ./...` again
- **WHEN** the audit reads it
- **THEN** `full_suite_after_text_fix` is `review` or `pass`, never `fail`

#### Scenario: A third code review round

- **GIVEN** a session whose agent loads the `code-review` skill in three messages with no push, merge, flow load or branch switch between them
- **WHEN** the audit reads it
- **THEN** `review_rounds_per_candidate` is `fail` and names the third launch

## MODIFIED Requirements

### Requirement: Session questions carry fixed headers

The delivery mode, the code review mechanism and the security review questions SHALL carry a fixed header in the session language instead of a decision ID, with the rest of the decision format. A mechanical change SHALL get no code review question; every other change SHALL get a code review, so the question SHALL offer only review mechanisms, also under `Delivery: direct-base`. When the project guidance, apart from a retired `Review:` key of its `## Hive` section, names a review mechanism that runs on the current host, the agent SHALL state it in one line that says how to change it and does not end in a question, instead of asking the code review question, and that mechanism SHALL be the review answer in force. When the delivery in force is `hold` and the working tree holds uncommitted changes of an earlier work item, the agent SHALL ask the delivery question again before the next work item's first edit, in the same round as the where-to-work question of `git-workflow` step 3 when both apply.

#### Scenario: Delivery round of a UI fix

- **GIVEN** a work item that fixes a screen flow
- **WHEN** the agent asks the session questions
- **THEN** they read as **Delivery** and **Review** without `D<n>`, and the review options are mechanisms

#### Scenario: The project names its review

- **GIVEN** a repository whose `AGENTS.md` prefers `/code-review` and a Claude Code session
- **WHEN** the agent reaches the delivery round
- **THEN** it asks the delivery question and states the review in one line, without asking it

#### Scenario: Hold left changes behind

- **GIVEN** a session whose delivery is `hold` and whose working tree holds the uncommitted changes of a finished work item
- **WHEN** the next work item reaches its first edit
- **THEN** the agent asks the delivery question again before editing

### Requirement: Plan revisions keep the change folder coherent
A revision of a written plan that has a change folder SHALL edit in the same pass every artifact that states what changed, mark a replaced decision or criterion with `**Replaces:** <old IDs>` and a dropped criterion with `**Retired:** <reason>` without deleting the old line, reopen each completed task the revision invalidates, and run the coherence check (`flow-plan/scripts/check_change.py`, which also validates the canonical grammar of lines under `## Business rules` in `proposal.md` and of the delivery-unit boxes under `## Delivery units` in `tasks.md` when run without `--archive`) until it exits 0 before work continues. A revision that changes a contract, an interface, a criterion, a decision, an assumption, an external fact, a verification line or the delta SHALL be reread by the plan's reviewers, each through one follow-up message to the same child, or, when its ID is gone, the host has no tool to message it or the message returns an error, by a new child of its role given the earlier findings; a reviewer SHALL NOT reread a revision that applies its own proposal as proposed. During planning, the plan SHALL get at most one reread round, and a gap that survives it, or one a later revision brings, SHALL go to the user as a decision in the plan summary; during the build, each such revision SHALL get one reread round.

#### Scenario: A task still names a replaced decision
- **GIVEN** a change folder whose revision replaced a decision and marked the new one with `**Replaces:**`
- **WHEN** the coherence check runs and a task still names the replaced decision
- **THEN** the check prints that task's file and line and exits 1, and the revision is not complete until it exits 0

#### Scenario: A revision changes a criterion
- **GIVEN** a change folder whose plan was already reviewed
- **WHEN** a revision changes a criterion or the delta
- **THEN** the plan's reviewer rereads the revised folder in one round, through a follow-up message to the same child or a new child given its earlier findings, and a gap that survives the round goes to the user as a decision

#### Scenario: A second revision during planning
- **GIVEN** a plan whose reviewers already reread one revision
- **WHEN** a later revision before the plan summary changes a criterion
- **THEN** no reviewer rereads it, and the summary lists what it changed for the user to decide

#### Scenario: Malformed business rule line in proposal
- **GIVEN** a change folder whose `proposal.md` holds a line under `## Business rules` that violates the canonical prefix or field grammar
- **WHEN** `check_change.py` runs without `--archive`
- **THEN** the check prints the file and line of the malformed entry and exits 1

#### Scenario: A live box without its predicate
- **GIVEN** a `tasks.md` whose `## Delivery units` section holds `- [ ] live: <scenario>` without `· pass: <predicate>` and without `n/a:`
- **WHEN** `check_change.py` runs without `--archive`
- **THEN** the check prints the file and line and exits 1

### Requirement: A plan that crosses domains gets one reviewer per domain

When a plan with a change folder has tasks that change files of two or more of these domains: server, client, guidance or infrastructure, or a data-entry screen where the user enters or chooses data, which also counts as client, the plan SHALL be reviewed by one `hive-review-plan` child per counted domain, launched together, with at most three children, the data-entry screen's included. Tests and documentation SHALL count with the domain they check or describe; repository-wide files, such as manifests, lockfiles, build configuration, the changelog and the change folder, SHALL count with the other files of the task that changes them, or with server when that task changes no other file; any other file no domain names SHALL count with server; and a task that changes files of two domains SHALL count in both. With four or more domains, the assignment with the fewest tasks SHALL be merged into the one with the next fewest, a merged assignment counting the sum of its tasks and taking its earliest domain's place, until at most three remain, a tie taking the later domain in that list, and the data-entry screen assignment SHALL never be merged. Each brief SHALL name its domain, the questions its domain examines and the contracts it shares with the other counted domains, and only the first assignment SHALL run the whole-plan checks of criteria, external facts and verification lines, the others being told to skip them. A finding several reviewers report SHALL be merged into one that names each reviewer. Any other plan with a change folder SHALL keep one reviewer. When a revision is reread, as the change-folder requirement sets, the reviewer of each domain whose tasks or criteria it changed SHALL reread it, the first assignment SHALL also reread a change to an external fact or a verification line, and every reviewer SHALL reread a change that belongs to no single domain.

#### Scenario: Server and client tasks

- **GIVEN** a plan whose tasks change a server endpoint and a web client's dialog where the user picks a time
- **WHEN** the plan is written
- **THEN** a server reviewer, a client reviewer and a data-entry screen reviewer are launched together
- **AND** each brief names the payload the server and the client share

#### Scenario: Four domains

- **GIVEN** a plan whose tasks change server, client, guidance and infrastructure files, with three server tasks, three client tasks, two guidance tasks and one infrastructure task
- **WHEN** the plan is written
- **THEN** three reviewers are launched, and infrastructure shares one assignment with guidance

#### Scenario: A plan in one domain

- **GIVEN** a plan whose tasks change only agent guidance files
- **WHEN** the plan is written
- **THEN** one `hive-review-plan` child reviews it

#### Scenario: A revision that touches one domain

- **GIVEN** a plan reviewed by a server and a client reviewer
- **WHEN** a revision changes a criterion that only a server task closes
- **THEN** the server reviewer rereads the revised folder, and the client reviewer does not
