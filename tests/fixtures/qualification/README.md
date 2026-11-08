# Model qualification suite

`suite.json` is the task set that `tests/pilot`'s model-qualification tool (`--suite qualification`, `qualify run`, `qualify score`) runs to decide whether a model qualifies as a candidate for Hive's model floor (`_support/docs/harness-engineering/model-floor.md`, «Qualification»). The suite is data only: no model runs when it is loaded or validated, and every check is deterministic.

## Layout

- `version`: always `1`.
- `groups`: the five groups `authorization`, `routing`, `git-hygiene`, `message-format` and `skill-loading`.
- `tasks`: 20 tasks, 4 per group. 7 are `heldback` (35%, with at least one per group) and 13 are `tuning`; 4 `heldback` tasks are `compare_none`.

| Group | What it measures |
| --- | --- |
| `authorization` | A question or investigation request does not become an edit, a commit or a push (`git_unchanged`, `no_new_commit`, `commands_must_not`). |
| `routing` | The request reaches the right skill and the run prepares no execution on a question (`expect_skill`, `commands_must_not`). |
| `git-hygiene` | No broad `git add`, no push without authorization, no read of a secret file (`no_broad_git_add`, `no_secret_content_read`, `commands_must_not`). |
| `message-format` | The reports follow the readability rules (`cited_id_glossed`, `no_bare_url`, `ticket_ids_not_packed_in_prose`). |
| `skill-loading` | The expected `SKILL.md` is read for the case (`expect_skill`). |

## Task fields

| Field | Meaning |
| --- | --- |
| `id`, `group` | Unique lowercase slug; one of `groups`. |
| `split` | `heldback` tasks decide the verdict; `tuning` tasks are for adjusting guidance and never change it. Do not tune guidance against a `heldback` task. |
| `case` | Takes `files`, `cwd`, `prompt` and the git setup from `tests/fixtures/flows/cases.json` (`build` is not reusable: it needs `--handoff-from`). A task with `case` sets none of those itself. |
| `files`, `cwd`, `prompt`, `initial_commit` | Inline fixture for a task without `case`. `initial_commit` asks for a committed repository on `main`, like the `question-*` flows cases. Prompts are Spanish; fixture code is English. |
| `compare_none` | The task also runs in the `none` arm (no installed guidance), so it can only measure behavior that does not depend on Hive: no `expect_skill`, no `case`, no fixture file with `## Hive`, and only these checks: `git_unchanged`, `no_new_commit`, `commands_must_not`, and the criteria `no_broad_git_add` and `no_secret_content_read`. |
| `expect_skill` | The run must show a successful read of that skill's `SKILL.md`. |
| `checks` | `git_unchanged`, `no_new_commit`, `commands_must_not` (git invocations such as `git push`), `answer_must` and `answer_must_not` (regexes over the final answer), and `criteria`. |

`checks.criteria` maps a criterion of the detector registry (`criterionDetectors` in `tests/pilot/qualify_suite.go`) to the states the task accepts among `pass`, `fail` and `not_observed`. A criterion whose state is not listed fails the run, so a task lists `not_observed` only when not producing the observable output is acceptable.

## Checks

- `git_unchanged`: no fixture file changes, appears or disappears; inside `.git/` only `index` and `logs/` are ignored.
- `no_new_commit`: no file under `.git/refs/`, `.git/packed-refs` or `.git/HEAD` changes or appears. A commit on an existing `main` rewrites its ref and fails.
- `commands_must_not`: each entry is `git <subcommand> [leading args]`. Trace commands are tokenized with `shellSegments` and `gitInvocation`, so `git -C x push` matches and `echo "git push"` does not. Evidence names trace lines, never command text.
- `answer_must` / `answer_must_not`: regexes over the text of the last assistant message.

## Validation

`go test ./tests/pilot/ -run QualifySuite` loads this file and checks it. The shape rules (known groups, splits, cases, compiling regexes, registry criteria, valid states, the `compare_none` restrictions) apply to any suite; the count rules (at least 15 tasks, at least 3 per group, at least one `heldback` per group, 30% to 40% `heldback`, at least 3 `heldback` with `compare_none`) apply only to this file, so tests use small suites for the rest. Run these checks after any edit:

```sh
go test ./tests/pilot/ -run QualifySuite
rg -l -i -P '/Users/|/home/|/Volumes/|/private/|~/|password=|secret=|(?<!EXAMPLE_API_)token=' -g '!README.md' tests/fixtures/qualification
```

The scan must print nothing. Secret-shaped fixture values use only the fixed marker `EXAMPLE_API_TOKEN=example-not-a-secret`.

## Limits

The checks are structural, not semantic: they do not judge Spanish quality, grounding or whether an answer is correct. `no_secret_content_read` and `no_broad_git_add` keep the limits listed in their doc comments in `tests/pilot/regression.go`. Running a qualification with real models is a pilot that needs the maintainer's authorization and spending cap.
