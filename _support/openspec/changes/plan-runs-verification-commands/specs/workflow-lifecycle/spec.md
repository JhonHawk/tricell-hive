## ADDED Requirements

### Requirement: A plan runs the verification commands it writes

Before writing a task's verification, the plan SHALL run each command it names once on the base; an observation or a walk SHALL NOT count as a command. When the command needs the change to exist, the plan SHALL run its nearest existing form instead: the same script or runner on a test it already finds, and, when the new test will import project code, a throwaway test that only imports those modules, kept outside the repository when the runner accepts a path and otherwise untracked and deleted in the same command. A command that changes tracked files, a database that is not local, shared or remote resources or an external system, or a long check, SHALL NOT be run while planning; the runner's own caches and ignored output, and a local test database already running, SHALL NOT count as such a change. A command that is not run, and one that needs a service the build will start, SHALL be written as not measured, with the task that meets it first. A command whose last measured run took over 5 minutes SHALL run only in its shortest form with the same runner; one with no measured run SHALL be cut to that form once it passes 5 minutes. Commands that share no database, port, lock or output directory SHALL be started together and awaited together; the others SHALL run one after another. The plan SHALL record one line per command in its verified context, or in the summary of a plan kept in the conversation, with the command, its result and the date. A command that cannot run because its script, runner or a dependency is missing on the base and is not added by a task, or because a module does not load, SHALL be changed before the plan review; one that a task adds SHALL be written as not measured with that task. The plan SHALL NOT switch or stash the checkout to reach the base. When a revision of the plan changes a task's verification command or what it rests on, the plan SHALL run it again and update its line before the next review round. The plan reviewer SHALL report as a gap a task's verification command with no line, or one recorded as run although the project lacks what it needs, without running it.

#### Scenario: A test command names a runner the app does not use

- **GIVEN** a task whose verification runs a new test with Vitest in an app whose test script is `node --test` and that has no Vitest dependency
- **WHEN** the plan writes that task
- **THEN** running the command on the base fails because Vitest is missing
- **AND** since no task adds Vitest, the plan changes the task's check before the review

#### Scenario: A revision changes the command

- **GIVEN** a plan whose review found that a task's test command cannot run
- **WHEN** the revision replaces it with another runner
- **THEN** the new command, with a throwaway test that imports the same modules, runs on the base before the next review round
- **AND** its line in the verified context shows the new result and date
