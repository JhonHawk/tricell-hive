## MODIFIED Requirements

### Requirement: Comment audit of a repository

Hive SHALL ship a `audit-comments` skill that inventories the comments of a repository or an area, classifies each one as tautology, stale, excuse, decision record, directive or constraint, and removes them in batches, filing a ticket for each excuse and writing each decision record before its comment is removed.

#### Scenario: An excuse comment
- **GIVEN** a comment that justifies a workaround
- **WHEN** its batch runs
- **THEN** a ticket for the workaround exists before the comment is removed
- **AND** the report links that ticket
