## MODIFIED Requirements
### Requirement: Verified release package installation
A complete package published as a GitHub Release asset SHALL be installable after the user verifies it against its published `.sha256` file and runs the package's `install.sh`. The installer SHALL validate and safely extract the package. The core installation SHALL work without Go, Git, GitHub CLI, or network downloads. Only these optional steps may download: the Pi host's pi-subagents step, from npm, and, after the operator consents to the agent-browser capability in an interactive install, its fixed recipe, which installs agent-browser with Homebrew or npm and runs `agent-browser install` to download Chrome for Testing only when no browser is found. Every package SHALL include `LICENSE` and `THIRD_PARTY_NOTICES.md`.

#### Scenario: Invalid archive
- **WHEN** the package checksum, archive paths, entry types, or limits fail validation
- **THEN** installation does not begin and no entry is written outside the task-owned extraction directory.

#### Scenario: Cancellation or unavailable terminal
- **WHEN** the user cancels before application or a controlling terminal is unavailable
- **THEN** no persistent installation changes occur.

#### Scenario: Pi host selected without pi-subagents
- **WHEN** a user-scope package installation includes the Pi host and Pi's settings do not declare pi-subagents
- **THEN** the plan runs `pi install npm:pi-subagents`, which downloads from npm; no other host step downloads.

#### Scenario: agent-browser capability consented
- **WHEN** the operator consents to the agent-browser capability in an interactive package installation
- **THEN** the plan runs only that capability's fixed recipe commands, which download from Homebrew or npm and, when no browser is found, Chrome for Testing; without that consent no capability step downloads.
