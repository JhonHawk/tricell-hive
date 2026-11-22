## MODIFIED Requirements

### Requirement: Tickets follow the tracker their key declares
When a repository's `## Hive` declares `Other trackers`, the workflow guidance SHALL read and search tickets in every declared tracker, SHALL move and comment each ticket, and write its business rules, in the tracker that declares its key prefix, using the primary `Tracker` for a key no tracker declares, and SHALL create new tickets only in the primary tracker. Declaring a tracker in `## Hive` SHALL authorize those moves, comments and rule writes as it does for the primary tracker.

#### Scenario: Ticket of an additional tracker reaches a later environment
- **WHEN** a deployment is verified and `Completion` holds a value for the prefix `XYZ`
- **THEN** the guidance moves each `XYZ-` ticket whose change that ref holds, to the declared state in the tracker that declares `XYZ`; when the deployment's functional walk runs, only a ticket whose acceptance criteria it observed in that environment moves, and when the walk does not run the ticket moves on the verified deployment and the report names it as not walked.

#### Scenario: Mirror tickets
- **WHEN** one work item names a ticket in the primary tracker and another in an additional tracker
- **THEN** the guidance moves and comments both, and writes the work's business rules to both, each in its own tracker.

#### Scenario: New ticket
- **WHEN** a flow proposes or creates a ticket in a repository with additional trackers
- **THEN** it searches every declared tracker for an existing one, as the `create-issue` skill sets, and the new ticket is created in the primary tracker.

## ADDED Requirements

### Requirement: Commits that appear after the push are asked before the merge
When a commit that the work did not make, such as a review bot's autofix, appears on the pull request's branch after the work's last push, the workflow guidance SHALL name it and ask whether it goes in the same publication before the merge or the `ready` stop, as it does before a push, and SHALL NOT merge or stop with it unasked.

#### Scenario: A bot's autofix after the push
- **WHEN** a review bot pushes an autofix commit to the branch after the work's last push and the next step is the merge
- **THEN** the agent names the commit and asks before merging.
