## MODIFIED Requirements

### Requirement: Hook registration per host
The manager SHALL install, update and remove only its own hook entries: entries whose command runs `hive hook` in Claude Code's user settings and Codex's and Cursor's hooks files, its own native hook file in Grok, its own event plugin in OpenCode's plugin directory, its own extension in Pi, and its own top-level named hook `hive` in Antigravity's `hooks.json`, leaving every other hook and setting byte-identical and keeping a backup of each edited JSON file. It SHALL NOT write Codex hook trust; its report SHALL tell the user to approve the hook in Codex. Hook processes SHALL print nothing, except on Antigravity, where every run, including one that fails, times out or receives unknown arguments, SHALL print the smallest JSON object that Antigravity's contract for that event accepts and that changes nothing in the run; they SHALL finish within the host's time limit and exit successfully on any error. Host support not verified against current documentation and the installed version SHALL be reported as unverified.

#### Scenario: Existing third-party hooks
- WHEN Claude Code's settings already hold other tools' hooks and Hive installs and later removes its hooks
- THEN the other hooks and settings stay byte-identical after both operations.

#### Scenario: Antigravity hook with a foreign named hook
- WHEN Antigravity's `hooks.json` already holds another tool's named hook and Hive installs, updates or removes the Antigravity host
- THEN the other named hook keeps the same bytes and only the `hive` key is added, changed or removed.

#### Scenario: Antigravity hook run that fails
- WHEN `hive hook` runs for Antigravity and cannot deliver the event
- THEN it still prints the accepted JSON object for that event and exits successfully.
