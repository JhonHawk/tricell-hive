# audit-issues

## ADDED Requirements

### Requirement: Audit of open tickets

Hive SHALL ship an `audit-issues` skill that counts a tracker's candidate tickets and suggests the scope of a pass, checks each ticket in the pass against the current code, and gives each a verdict with its evidence: still valid, close as a false positive, already fixed, duplicate of a named ticket or covered by a named change, not confirmed, or not ready.

#### Scenario: A duplicate without firm evidence
- **GIVEN** two tickets whose text is similar but whose error, reproduction steps or root cause in code differ
- **WHEN** the pass runs
- **THEN** the skill proposes a related link between them, not a duplicate closure

#### Scenario: A reproduction that fails
- **GIVEN** a Bug ticket whose reproduction fails and whose code does not show the failure cannot happen
- **WHEN** the pass runs
- **THEN** its verdict is not confirmed and the skill does not propose closing it

### Requirement: Groups under a parent ticket

The skill SHALL group the valid tickets that share a root cause, a domain concept or a module or view under a parent ticket, naming which of those tiers formed each group; a root-cause group is a parent whose children one change closes, and a domain or module group is a parent whose children are delivered separately. Tickets that share only a theme get a related link instead of a group. A ticket in progress, with an open branch or pull request, or that already has a parent is reported but not moved: its group reuses its existing parent, or it gets a related link.

#### Scenario: Tickets about the same concept
- **GIVEN** several valid tickets about timezone handling in different modules
- **WHEN** the pass forms groups
- **THEN** they are proposed under one parent with the tier named
- **AND** each member shows the evidence that places it in the group

### Requirement: One approval before tracker writes

The skill SHALL show the whole proposed set, with each ticket's verdict and evidence and whether closing a parent closes its children in that tracker, and write nothing to the tracker before one approval of the set. After approval it writes one change at a time, reads each ticket back, and stops at the first failed write. A ticket another account wrote may be moved under a parent but is never closed: it gets a comment for its author instead. While unattended delegation is active, the skill writes nothing and reports the proposed set as unresolved.

#### Scenario: Closing another author's ticket
- **GIVEN** an approved set where a ticket written by another account has a closing verdict
- **WHEN** the skill writes the set
- **THEN** that ticket gets a comment with the verdict and its evidence and stays open
