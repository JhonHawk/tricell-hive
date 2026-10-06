## ADDED Requirements
### Requirement: Verified release package installation
A complete package published as a GitHub Release asset SHALL be installable after the user verifies it against its published `.sha256` file and runs the package's `install.sh`. The installer SHALL validate and safely extract the package. The core installation SHALL work without Go, Git, GitHub CLI, or network downloads; only the Pi host's pi-subagents step may download, from npm. Every package SHALL include `LICENSE` and `THIRD_PARTY_NOTICES.md`.

#### Scenario: Invalid archive
- **WHEN** the package checksum, archive paths, entry types, or limits fail validation
- **THEN** installation does not begin and no entry is written outside the task-owned extraction directory.

#### Scenario: Cancellation or unavailable terminal
- **WHEN** the user cancels before application or a controlling terminal is unavailable
- **THEN** no persistent installation changes occur.

#### Scenario: Pi host selected without pi-subagents
- **WHEN** a user-scope package installation includes the Pi host and Pi's settings do not declare pi-subagents
- **THEN** the plan runs `pi install` for the pinned pi-subagents source, which downloads from npm; no other host step downloads.

## REMOVED Requirements
### Requirement: Verified online bootstrap and offline installation
The online `bootstrap.sh` entry was retired in 0.1.0 (decision D8-A); its offline part is restated in "Verified release package installation".

### Requirement: Recovery executable retention
It applied only to the retired online installation. The unused `hive bootstrap` subcommand remains in code until a follow-up ticket retires or reconnects it.
