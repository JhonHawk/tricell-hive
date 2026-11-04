## ADDED Requirements

### Requirement: Additional trackers in project settings
The `## Hive` section SHALL accept an optional `Other trackers` key after `Tracker`, holding one or more trackers separated by `; `, each as `<tool> · <location> · keys <PREFIX>[, <PREFIX>]`, while `Tracker` remains the required primary tracker and is not parsed. `Completion` SHALL accept one or more values separated by `; `, where at most one value without a prefix applies to the primary tracker and a value `<PREFIX>[, <PREFIX>]: <state> on <ref> · <state> on <ref>` applies to tickets with those key prefixes; each value SHALL be checked on its own. `hive doctor`, `hive project set` and the Project view SHALL validate and write both keys, and every finding on them SHALL be a warning that does not change the exit status.

#### Scenario: Valid additional tracker
- **WHEN** `hive doctor --project` runs on a repository whose section holds `Other trackers: Jira · example.atlassian.net · keys XYZ` and `Completion: XYZ: QA on qa · Done on production`
- **THEN** it reports no finding on `Other trackers` or `Completion`.

#### Scenario: Tracker without key prefixes
- **WHEN** a tracker in `Other trackers` lacks `keys <PREFIX>`, or a prefix appears in two trackers
- **THEN** `hive doctor` warns naming the tracker or the prefix, and exits with the status it would have without that finding.

#### Scenario: Malformed or undeclared Completion value
- **WHEN** a `Completion` value after `; ` does not match `<state> on <ref> · <state> on <ref>`, or names a prefix that no tracker in `Other trackers` declares
- **THEN** `hive doctor` warns naming that value or prefix.

#### Scenario: Editing the key
- **WHEN** the user sets `Other trackers` with `hive project set` or the Project view
- **THEN** the line is written in key order when the section is new, or after the last item of an existing section like any new key, and every other line of the section is unchanged.
