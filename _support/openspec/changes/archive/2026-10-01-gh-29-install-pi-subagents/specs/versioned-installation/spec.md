## ADDED Requirements

### Requirement: Pi user-scope install declares pi-subagents
When a user-scope plan includes the Pi host, the manager SHALL leave `pi-subagents` selectable by Pi's native package mechanism without rewriting unrelated `settings.json` entries. If the user settings do not declare the package by npm name, apply SHALL run `pi install npm:pi-subagents@0.74.0` with `PI_CODING_AGENT_DIR` set to the plan's Pi home. If the package is already declared and the extensions filter is omitted, the plan SHALL omit install and SHALL NOT change the existing pin. A declaration with `extensions: []`, or a non-empty filter that does not demonstrably include the extension, SHALL be a conflict or an unverified-availability report, not success, and SHALL NOT change settings. Missing files or a pin/version mismatch on a pre-existing declaration SHALL be reported without reinstalling. Recover SHALL remove the package only when this operation added the exact source string and that string is still present. Tests SHALL use a fake `pi` and a synthetic home and SHALL NOT call npm or write the real user settings.

#### Scenario: Package absent
- **WHEN** a user-scope install or update plan includes `pi` and the planned Pi settings do not declare `pi-subagents`
- **THEN** the plan shows install of `npm:pi-subagents@0.74.0`, apply adds that source, and other package entries remain.

#### Scenario: Package already declared
- **WHEN** the planned Pi settings already declare `pi-subagents` with a loadable extension
- **THEN** the plan omits install, apply does not invoke `pi install`, and recover does not remove the entry.

#### Scenario: Empty extensions filter
- **WHEN** the planned Pi settings declare `pi-subagents` with `extensions: []`
- **THEN** the plan reports a conflict and apply does not change settings.

## MODIFIED Requirements

### Requirement: Optional operation capabilities
The optional capability catalog SHALL be limited to Engram, Context7, and pi-subagents. Engram and Context7 SHALL remain manual until each recipe passes a native gate for the host and platform; a selected capability without a verified recipe SHALL be reported as manual, with its official source and reason, and SHALL NOT be presented as installed or verified. pi-subagents on a user-scope Pi host follows the Pi user-scope install declares pi-subagents requirement instead of that manual path. Development-of-Hive requirements SHALL remain separate from installation and operation requirements. Hosts, runtimes and other tools SHALL NOT be installed implicitly.

#### Scenario: Capability without a verified recipe
- **WHEN** the user selects Engram or Context7, whose recipes have not passed their native gate
- **THEN** the core installs, no provider process runs, and the result lists the capability as manual with its official instructions and a non-zero exit status.
