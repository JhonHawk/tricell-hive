## MODIFIED Requirements
### Requirement: Pi user-scope install declares pi-subagents
When a user-scope plan includes the Pi host, the manager SHALL leave `pi-subagents` selectable by Pi's native package mechanism without rewriting unrelated `settings.json` entries. If the user settings do not declare the package by npm name, apply SHALL run `pi install npm:pi-subagents` with `PI_CODING_AGENT_DIR` set to the plan's Pi home. If the package is already declared and the extensions filter is omitted, the plan SHALL omit install and SHALL NOT change the existing pin. A declaration with `extensions: []`, or a non-empty filter that does not demonstrably include the extension, SHALL be a conflict or an unverified-availability report, not success, and SHALL NOT change settings. Missing files or a pin/version mismatch on a pre-existing declaration SHALL be reported without reinstalling. Recover SHALL remove the package only when this operation added the exact source string and that string is still present. Tests SHALL use a fake `pi` and a synthetic home and SHALL NOT call npm or write the real user settings.

#### Scenario: Package absent
- **WHEN** a user-scope install or update plan includes `pi` and the planned Pi settings do not declare `pi-subagents`
- **THEN** the plan shows install of `npm:pi-subagents`, apply adds that source, and other package entries remain.

#### Scenario: Package already declared
- **WHEN** the planned Pi settings already declare `pi-subagents` with a loadable extension
- **THEN** the plan omits install, apply does not invoke `pi install`, and recover does not remove the entry.

#### Scenario: Empty extensions filter
- **WHEN** the planned Pi settings declare `pi-subagents` with `extensions: []`
- **THEN** the plan reports a conflict and apply does not change settings.

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
- **THEN** the plan runs `pi install npm:pi-subagents`, which downloads from npm; no other host step downloads.
