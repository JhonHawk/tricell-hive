## MODIFIED Requirements

### Requirement: The session audit flags full-suite reruns after text fixes and extra review rounds

`go run ./tests/pilot --audit-session` SHALL recognize an unfiltered suite run through the session's shell segments, not text inside arguments or heredocs: `go test ./...` and `go -C <dir> test ./...` with `-count` or `-race`, `npm test`, `pnpm test` and `pytest` without arguments, while `-run` or a sub-package pattern filters it. It SHALL report `full_suite_after_text_fix`, for two runs in the same working directory with no `git push`, `gh pr merge`, new branch or new worktree between them, as `fail` when the second follows a first run shown to pass, by an `echo` or `printf` of its own `$?` in the next command, whose whole printed line shows 0, or by the success of a command run without a pipe or background, a pipe inside a `bash -c` or `sh -c` argument counting as a pipe, a background launch or a run with no result never counting as shown, and everything changed between them is `.md`, `.mdx` or `.txt` files outside `testdata/` and `tests/fixtures/` folders, by edit tools or by shell writes, interpreter scripts included, whose targets are of those types; as `pass` when another edit, or a rebase, merge, pull, checkout, switch, cherry-pick, am, apply, reset or restore, lies between them; as `review` when a shell write has a target that cannot be shown to be text or another kind or the first run cannot be shown to pass; and as `not_observed` without two runs, never counting `git add`, `git commit` or `git push` as a change. It SHALL report `review_rounds_per_candidate` for main Claude Code sessions, counting code review rounds (the `code-review` skill and `hive-review-code` children, the launches of one message being one round, and a `/code-review` the user typed and the skill launch that carries it out being one round) and resetting at `git push`, `gh pr merge`, a new load of `flow-build` or `flow-plan` or a branch switch, a `git checkout` or `git restore` that names only paths not being one: `fail` with more than two rounds, `review` when the third follows a `/code-review` the user typed, `pass` with two or fewer, and `not_observed` without reviews, in child sessions and on other hosts.

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

#### Scenario: Test data between the runs

- **GIVEN** a session that runs `go test ./...` with a passing result, edits `pkg/testdata/expected.txt` and runs `go test ./...` again
- **WHEN** the audit reads it
- **THEN** `full_suite_after_text_fix` is `pass`

#### Scenario: A typed review

- **GIVEN** a session where the user types `/code-review` twice and the agent loads the `code-review` skill once after each
- **WHEN** the audit reads it
- **THEN** `review_rounds_per_candidate` counts two rounds and is `pass`
