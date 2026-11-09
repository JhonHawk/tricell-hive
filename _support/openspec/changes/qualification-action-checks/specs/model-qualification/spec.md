## MODIFIED Requirements

### Requirement: Qualification suite and arms
The pilot runner SHALL run the tasks of `tests/fixtures/qualification/suite.json` with `--suite qualification` under `--arm deployed` (the user's deployed guidance) on every runner host, `--arm A` with `--guidance-source` on Codex, Grok and OpenCode, and `--arm none` on Codex and OpenCode, which prepares the same isolated home as arm A (credentials, memory isolation, protected-path fingerprints) without installing any Hive-managed file. It SHALL reject `--arm none` on Claude, Grok and Pi before launching, naming the limitation. Each qualification run SHALL record its task, split, arm, requested effort and a hash of the guidance it ran with, or `none`. A task marked `compare_none` SHALL use only checks that do not depend on Hive's own formats or files. The suite validator SHALL reject any task that a run in which the model changes no file, runs no command and only asks whether to continue would pass, unless the task is marked `no_action_expected`, which SHALL require `git_unchanged`.

#### Scenario: The none arm on an unsupported host
- **WHEN** `--suite qualification --arm none` is passed with `--host grok`
- **THEN** the runner exits with an error naming the limitation and launches nothing.

#### Scenario: A clean none home
- **WHEN** a Codex qualification run uses `--arm none`
- **THEN** no file in its isolated home carries Hive's managed block marker or a Hive skill or agent.

#### Scenario: A task that passes when nothing is done
- **WHEN** a task asks for a local commit and checks only that no push ran
- **THEN** the suite validator rejects it, naming the task.
