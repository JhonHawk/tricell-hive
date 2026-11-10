## MODIFIED Requirements

### Requirement: A plan authorizes foreseen effects by class

Each plan SHALL ask one decision per class of effect its tasks foresee on an external system: in the delivery and review round when that round is asked, and in a round of its own when those answers are already in force. A class SHALL name one kind of action on one named target, its spending ceiling and the tasks that use it, and SHALL leave out the effects the delivery answer or a rule already authorizes. A class the user approved SHALL be the user's instruction for each effect of that class that the named tasks of that work item take within the ceiling, and for no other plan or work item. A class the user refused SHALL authorize nothing and SHALL NOT be asked again: its tasks are revised with the user, or skipped and reported in unattended work. A class left unanswered SHALL authorize nothing. When a revision of the plan changes an approved class's kind of action, target, ceiling or tasks, its answer SHALL return to unanswered and the class SHALL be asked again before any further effect of it. A class whose target is a production one SHALL say so in its question. No class SHALL cover a production deployment, an effect that stops or interrupts a component serving users, an irreversible or hard-to-reverse effect, destroying or replacing infrastructure, changing or deleting data the task did not create in a non-local database, or creating accounts or records in a shared environment; each of those SHALL keep its own question. Each covered effect SHALL be recorded with its target and how to revert it before it is taken, and listed in the final report. An effect no class was asked for SHALL be asked when it arrives in attended work. An operational task SHALL have no such round.

#### Scenario: Several effects of one class

- **GIVEN** a plan whose tasks change access policies in one cloud account several times
- **WHEN** the user approves that class in the plan's round
- **THEN** each of those changes is taken without a new question, within the ceiling
- **AND** the final report lists each one with how to revert it

#### Scenario: Delivery answers already in force

- **GIVEN** a second plan in a session whose delivery and review answers are reused
- **WHEN** that plan foresees effects on an external system
- **THEN** it asks their classes in a round of its own

#### Scenario: Production deployment in the same plan

- **GIVEN** a plan with an approved class and a promotion task that reaches production
- **WHEN** the deploy stage reaches production
- **THEN** it asks the production question, whatever the classes approved

#### Scenario: Refused class

- **GIVEN** a class the user refused in the plan's round
- **WHEN** a task of that plan needs an effect of that class
- **THEN** the effect is not taken and the class is not asked again
- **AND** the task is revised with the user, or skipped and reported in unattended work

#### Scenario: Revision changes an approved class

- **GIVEN** a class the user approved in the plan's round
- **WHEN** a revision of the plan changes its target, its ceiling, its kind of action or its tasks
- **THEN** its answer returns to unanswered
- **AND** the class is asked again before any further effect of it

#### Scenario: Deploy run that does not come from the plan

- **GIVEN** a deploy run that carries tickets of several work items
- **WHEN** it reaches a non-production environment
- **THEN** it asks the authorization for that environment, whatever a plan's classes answered

#### Scenario: Effect outside every class

- **GIVEN** an effect the plan did not foresee
- **WHEN** a task needs it in attended work
- **THEN** the agent asks it as its own question

