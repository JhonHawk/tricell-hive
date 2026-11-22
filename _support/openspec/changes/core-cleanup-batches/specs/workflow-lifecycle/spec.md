## MODIFIED Requirements

### Requirement: Tickets follow the tracker their key declares
When a repository's `## Hive` declares `Other trackers`, the workflow guidance SHALL read and search tickets in every declared tracker, SHALL move and comment each ticket, and write its business rules, in the tracker that declares its key prefix, using the primary `Tracker` for a key no tracker declares, and SHALL create new tickets only in the primary tracker unless the user names another tracker or repository for that ticket. Declaring a tracker in `## Hive` SHALL authorize those moves, comments and rule writes as it does for the primary tracker.

#### Scenario: Ticket of an additional tracker reaches a later environment
- **WHEN** a deployment is verified and `Completion` holds a value for the prefix `XYZ`
- **THEN** the guidance moves each `XYZ-` ticket whose change that ref holds, to the declared state in the tracker that declares `XYZ`; when the deployment's functional walk runs, only a ticket whose acceptance criteria it observed in that environment moves, and when the walk does not run the ticket moves on the verified deployment and the report names it as not walked.

#### Scenario: Mirror tickets
- **WHEN** one work item names a ticket in the primary tracker and another in an additional tracker
- **THEN** the guidance moves and comments both, and writes the work's business rules to both, each in its own tracker.

#### Scenario: New ticket
- **WHEN** a flow proposes or creates a ticket in a repository with additional trackers
- **THEN** it searches every declared tracker for an existing one, as the `create-issue` skill sets, and the new ticket is created in the primary tracker.

#### Scenario: The user names another tracker for a new ticket
- **WHEN** the user asks to create a ticket in a declared additional tracker
- **THEN** the ticket is created there, and `ticket-state.md` is the one place that states this exception.
