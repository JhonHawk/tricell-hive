## ADDED Requirements

### Requirement: Tickets are created through one procedure

The workflow guidance SHALL keep the procedure for creating, breaking down and validating a tracker ticket in the `create-issue` skill, and every other flow SHALL point to it instead of stating its steps. The global guidance SHALL keep when a ticket is proposed and that creating one needs the user's answer. The skill SHALL create a ticket of one of four types (story, bug, task, spike) with that type's sections, SHALL ask at most three questions per round, each with a recommended answer, SHALL write what it does not ask as an assumption in the ticket, and SHALL check the ticket against its fixed readiness list in the same thread, without a reviewer child. Run on its own, the skill SHALL end at the ticket and only offer the next step; followed from a flow, it SHALL return to that flow's next step. It SHALL NOT add a statement that an agent wrote the ticket unless the target repository's contribution rules ask for one.

#### Scenario: A flow needs a ticket
- **WHEN** a flow's finding or a separate topic gets the user's answer to create a ticket
- **THEN** the flow follows the `create-issue` skill in the same turn, which searches every declared tracker for an existing one, relates the new ticket to its origin and creates it in the primary tracker without another confirmation, shows the created ticket with its link, and returns to the flow's next step.

#### Scenario: The skill is used on its own
- **WHEN** the user asks for a ticket outside a flow
- **THEN** the skill shows the draft, creates the ticket after the user confirms, and ends by offering the next step.

#### Scenario: An undecided business rule
- **WHEN** a ticket is created while one of its business rules is undecided
- **THEN** the ticket is created marked as blocked, through the tracker's native blocked-by relation, a blocked label or state the project already has, or else the first line of the ticket with the question and who it waits on, and no label or state is created.

#### Scenario: A ticket chosen for research
- **WHEN** the user picks a ticket for research
- **THEN** `flow-research` reports its impact, complexity and dependencies, and judges its readiness with the skill's readiness list.

#### Scenario: A session without a repository
- **WHEN** no `## Hive` section declares a tracker for the session
- **THEN** the skill asks which tracker and repository to create the ticket in before it searches or writes, and omits the solution hint, saying so in the ticket.

### Requirement: A ticket states what is needed, not how it is built

A ticket the skill creates SHALL state the outcome, how it will be verified and what is out of scope, and SHALL carry at most one solution hint, in a separate section marked as non-binding, with at most five locations and one sentence of direction. It SHALL NOT carry a list of implementation steps, a choice between approaches, a contract definition or a file-by-file change list. `flow-plan` SHALL treat a ticket's hint as a claim to verify.

#### Scenario: The hint would need a decision
- **WHEN** writing the hint would require choosing between approaches or defining a contract
- **THEN** the ticket says that it requires a plan and carries no hint.

### Requirement: Work is broken down into a parent ticket with children on one approval

The skill SHALL propose a breakdown when the work names more than one change that integrates separately, touches more than one repository, or has more than 7 acceptance criteria. It SHALL show the whole set, with the parent, each child and their dependencies, and SHALL ask one approval for the whole set before the first write. It SHALL write the tickets one at a time, all in the primary tracker and each child naming its repository, SHALL record each one as it is created and read each one back, and SHALL stop and report what was created when a write fails. A parent ticket with children SHALL be created as Jira's epic type, as a parent issue with sub-issues in GitHub, and as a parent issue with sub-issues in Linear, without creating a project.

#### Scenario: Unattended work reaches a breakdown
- **WHEN** unattended delegation is active and the work would need a parent ticket with children
- **THEN** the skill creates none of them and reports the proposed set as unresolved.

#### Scenario: Unattended work opens a ticket for a finding
- **WHEN** unattended delegation is active and a finding needs a single ticket
- **THEN** the skill creates it without asking for a confirmation.

#### Scenario: A child is not separately integrable
- **WHEN** a part of the work cannot be integrated and accepted on its own
- **THEN** it is not proposed as a child ticket.

### Requirement: An existing ticket is validated without rewriting another person's content

When the skill validates a ticket that already exists, it SHALL first check whether the authenticated account is the ticket's author and what the account may write. When the account is not the author, it SHALL NOT edit the ticket: it SHALL show a comment with what is missing and the open questions, each with a recommended answer, and post it after the user confirms. When the account is the author, it SHALL show the change and edit after the user confirms. When the account cannot write to the tracker, it SHALL hand the report to the user. An undecided business rule SHALL always go to the ticket's owner as a question.

#### Scenario: A developer validates a product owner's ticket
- **WHEN** the authenticated account did not write the ticket and its criteria are missing
- **THEN** the skill shows a comment naming the missing criteria and its questions, and posts it only after the user confirms.

## MODIFIED Requirements

### Requirement: Tickets follow the tracker their key declares
When a repository's `## Hive` declares `Other trackers`, the workflow guidance SHALL read and search tickets in every declared tracker, SHALL move and comment each ticket, and write its business rules, in the tracker that declares its key prefix, using the primary `Tracker` for a key no tracker declares, and SHALL create new tickets only in the primary tracker. Declaring a tracker in `## Hive` SHALL authorize those moves, comments and rule writes as it does for the primary tracker.

#### Scenario: Ticket of an additional tracker reaches a later environment
- **WHEN** a deployment is verified and `Completion` holds a value for the prefix `XYZ`
- **THEN** the guidance moves each `XYZ-` ticket whose change that ref holds to the declared state in the tracker that declares `XYZ`.

#### Scenario: Mirror tickets
- **WHEN** one work item names a ticket in the primary tracker and another in an additional tracker
- **THEN** the guidance moves and comments both, and writes the work's business rules to both, each in its own tracker.

#### Scenario: New ticket
- **WHEN** a flow proposes or creates a ticket in a repository with additional trackers
- **THEN** it searches every declared tracker for an existing one, as the `create-issue` skill sets, and the new ticket is created in the primary tracker.
