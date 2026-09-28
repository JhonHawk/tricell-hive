## ADDED Requirements

### Requirement: Update from a committed revision
From a Git checkout of Hive, the manager SHALL build an install plan from the content of one resolved commit, excluding uncommitted changes, for every consumer already installed in user scope. It SHALL apply that plan only after interactive confirmation, and without a terminal SHALL only preview or save the plan for a separate `apply`. It SHALL remove its temporary extraction on every return path. The offline package SHALL NOT require Git.

#### Scenario: Uncommitted changes in the checkout
- **WHEN** the checkout has uncommitted changes under `content/` and the user runs `hive update`
- **THEN** the applied release contains the committed content only.

#### Scenario: No terminal
- **WHEN** `hive update` runs without a terminal and without `--dry-run` or `--out`
- **THEN** it changes nothing and names the options that preview or save the plan.

#### Scenario: Invalid revision
- **WHEN** the revision starts with `-`, names no commit, or the source is not a Git checkout
- **THEN** the command fails before extracting or writing anything.

### Requirement: Retained release listing
The manager SHALL list every retained release snapshot with its identifier, recording time, the source commits recorded for it, and the consumers where it is installed, without modifying state. Applying a plan that carries a source commit SHALL record that commit beside the release, including when the plan changes no file, without changing the snapshot or its identifier.

#### Scenario: Release installed before commit recording
- **WHEN** a retained release has no commit record
- **THEN** it is listed with an empty commit list.

#### Scenario: Same content from two commits
- **WHEN** two commits produce the same release identifier and both are applied
- **THEN** the release lists both commits once each.
