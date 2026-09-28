## ADDED Requirements

### Requirement: Interactive terminal interface
Running the manager without arguments in a terminal, or `hive tui`, SHALL open a full-screen interface in the terminal's alternate screen for host status, installing and removing hosts, updating from a commit, returning to a retained release, and the voice layer. Changing views SHALL redraw the whole screen, and leaving the interface SHALL restore the terminal's previous content. Every action that writes SHALL show a preview and ask for confirmation first, and SHALL produce the same files and state as the equivalent command. Declining, going back, or an error inside an action SHALL change nothing further and return to the originating view with the message shown inside it. Without a terminal, running the manager without arguments SHALL keep its usage error, and the text commands SHALL remain the interface for scripts and screen readers. The core management and distribution packages SHALL NOT depend on third-party modules.

#### Scenario: No terminal
- **WHEN** the manager runs without arguments and without a terminal
- **THEN** it prints the usage error and changes nothing.

#### Scenario: Going back
- **WHEN** the user presses Esc in any view, or Backspace while no text field has focus
- **THEN** the previous view returns, redrawn in place.

#### Scenario: Hosts changed in one pass
- **WHEN** the user checks some hosts and unchecks others in the hosts view and applies
- **THEN** the removal is previewed, confirmed and applied first, and the installation is then planned on the resulting state, previewed and confirmed separately.

#### Scenario: Declined or cancelled action
- **WHEN** the user declines or goes back from a confirmation
- **THEN** no file or state changes and the originating view shows that nothing was applied.
