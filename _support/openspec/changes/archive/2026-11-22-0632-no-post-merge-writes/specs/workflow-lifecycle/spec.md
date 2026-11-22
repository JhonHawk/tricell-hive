## ADDED Requirements

### Requirement: Nothing is written to an integrated change folder

Once a change folder is on `origin/<base>` with every task done, nothing SHALL be written to it until the hygiene batch archives it. `flow-plan/scripts/tasks.py` SHALL refuse `checkpoint` and `done` in that state, from any checkout, writing nothing, and SHALL name `--out`; `checkpoint --out <file>` SHALL write only the checkpoint block to that file, outside the folder's repository. A folder with an open task, or a state Git cannot answer, SHALL be written as before. The close SHALL keep its checkpoint after the merge in a file in the host's temporary location, and `tasks.py versioned` SHALL tell it to restore a `tasks.md` whose only uncommitted change is its checkpoint.

#### Scenario: The close updates its checkpoint after the merge

- **GIVEN** a change folder merged into `development` with every task done, and a checkout still on the merged branch
- **WHEN** the close runs `tasks.py <folder> checkpoint` without `--out`
- **THEN** the script exits 1 naming `--out`, and `tasks.md` is unchanged

#### Scenario: A later delivery unit of the same plan

- **GIVEN** a change folder already on `development` whose task T3 is still open
- **WHEN** the agent records T3 with `tasks.py <folder> done T3`
- **THEN** the task is recorded in `tasks.md` as before
