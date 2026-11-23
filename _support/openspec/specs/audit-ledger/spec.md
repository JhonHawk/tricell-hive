# audit-ledger

## Purpose

TBD: created by archiving ledger-writer-role; update the purpose.

## Requirements

### Requirement: A writing role keeps the audit ledgers

Hive SHALL ship a `hive-write-ledger` role that the comment and test audits brief for each lane: it writes one row per item to the ledger file the brief names, in the host's temporary location, edits no other file, and returns the ledger's path with its counts. The read-only `hive-research` role SHALL keep its read-only access.

#### Scenario: A lane of a comment audit
- **GIVEN** a comment audit whose scope exceeds the main thread's limit
- **WHEN** the skill delegates a lane
- **THEN** it briefs `hive-write-ledger`, and the child writes the lane's ledger file and returns its path and counts

#### Scenario: Ledger lanes run side by side
- **GIVEN** an audit that delegates several lanes to `hive-write-ledger`, each with its own ledger file
- **WHEN** the skill launches them
- **THEN** they run side by side, since each writes only its own file in the host's temporary location

#### Scenario: Research stays read-only
- **GIVEN** a host whose `observe` profile removes file writes
- **WHEN** `hive-research` is installed
- **THEN** its rendered role still carries that restriction
