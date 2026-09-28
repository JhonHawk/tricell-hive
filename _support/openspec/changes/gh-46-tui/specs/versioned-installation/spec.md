## ADDED Requirements

### Requirement: Interactive terminal interface
Running the manager without arguments in a terminal, or `hive tui`, SHALL open an interactive interface for status, installing and removing hosts, updating from a commit, returning to a retained release, and the voice layer. Every action that writes SHALL show a preview and ask for confirmation first, and SHALL produce the same files and state as the equivalent command. Cancelling, declining, the end of input, or an error inside an action SHALL change nothing further and return to the menu. Without a terminal and without the accessible mode, running the manager without arguments SHALL keep its usage error. An accessible mode SHALL offer the same interface as plain line-based prompts. The core management and distribution packages SHALL NOT depend on third-party modules.

#### Scenario: No terminal
- **WHEN** the manager runs without arguments, without a terminal, and `HIVE_ACCESSIBLE` is not `1`
- **THEN** it prints the usage error and changes nothing.

#### Scenario: Accessible mode
- **WHEN** `HIVE_ACCESSIBLE` is `1`
- **THEN** the same menu runs as plain prompts that read standard input, and the end of input exits.

#### Scenario: Declined or cancelled action
- **WHEN** the user declines, cancels, or reaches the end of input inside an action
- **THEN** no file or state changes and the menu returns.
