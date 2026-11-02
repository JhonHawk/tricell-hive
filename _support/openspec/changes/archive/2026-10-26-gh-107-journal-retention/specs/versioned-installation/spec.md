## ADDED Requirements

### Requirement: Transaction journal retention
After an operation commits, the manager SHALL keep at most the 10 most recent transaction journals in the `committed` or `recovered` phase and delete older ones, both for a plain apply and for an installation with optional steps. It SHALL NOT delete the journal named by a pending operation, the core journal referenced by a pending onboarding journal, or a journal in any other phase. When the manager cannot determine which journals are protected, it SHALL delete none. A pruning failure SHALL be reported as a warning and SHALL NOT fail the committed operation. The manager SHALL describe the transactions directory as recovery journals, not as backups.

#### Scenario: Old committed journals are pruned
- **WHEN** a twelfth operation commits while eleven earlier journals are committed
- **THEN** the ten most recent committed journals remain, including the new one, and the others are deleted.

#### Scenario: A journal needed for recovery is kept
- **WHEN** an operation commits while `pending.json` or `onboarding-pending.json` references an older journal
- **THEN** that journal remains regardless of its age.
