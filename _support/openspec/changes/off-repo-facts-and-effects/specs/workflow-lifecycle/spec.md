## ADDED Requirements

### Requirement: A plan measures the external facts it depends on

When a plan's chosen approach reads or changes an external system, the plan SHALL list each external fact the approach depends on and measure it with a read-only read before its tasks are written. An external system SHALL be one provider account or service reached by its own credential or CLI; what the delivery already uses of the repository's remote, CI and tracker SHALL NOT count as one. With two or more external systems, one read-only child per external system SHALL be launched in the same step, each a state-reading child briefed to return, per fact, the exact read-only command and a verbatim excerpt of its output. The main thread SHALL accept a fact only with both, and SHALL compare a capability, limit or permission fact with the provider's current documentation. A fact that observation and documentation disagree on, or that no read-only read can show, SHALL be written in the plan as not measured, with the task that meets it first, and SHALL appear in the plan summary.

#### Scenario: Plan across two external systems

- **GIVEN** a plan whose approach depends on another branch's contents in a hosting organization and on a cloud account's permissions
- **WHEN** the approach is chosen
- **THEN** one read-only child per system is launched in the same step
- **AND** the plan records each fact with its command or source, its result and its date

#### Scenario: Fact that cannot be measured

- **GIVEN** an external fact no read-only read can show
- **WHEN** the plan is written
- **THEN** the plan names it as not measured with the task that meets it first
- **AND** the plan summary shows it to the user

#### Scenario: Plan that only delivers through its own repository

- **GIVEN** a plan whose approach pushes a branch, opens a pull request, waits for its checks and moves its ticket, and reaches nothing else
- **WHEN** the plan is written
- **THEN** no external fact is listed and no child is launched for this purpose

### Requirement: A plan authorizes foreseen effects by class

Each plan SHALL ask one decision per class of effect its tasks foresee on an external system: in the delivery and review round when that round is asked, and in a round of its own when those answers are already in force. A class SHALL name one kind of action on one named target, its spending ceiling and the tasks that use it, and SHALL leave out the effects the delivery answer or a rule already authorizes. A class the user approved SHALL be the user's instruction for each effect of that class that the named tasks of that work item take within the ceiling, and for no other plan or work item. A class the user refused SHALL authorize nothing and SHALL NOT be asked again: its tasks are revised with the user, or skipped and reported in unattended work. A class left unanswered SHALL authorize nothing. A class whose target is a production one SHALL say so in its question. No class SHALL cover a production deployment, an effect that stops or interrupts a component serving users, an irreversible or hard-to-reverse effect, destroying or replacing infrastructure, changing or deleting data the task did not create in a non-local database, or creating accounts or records in a shared environment; each of those SHALL keep its own question. Each covered effect SHALL be recorded with its target and how to revert it before it is taken, and listed in the final report. An effect no class was asked for SHALL be asked when it arrives in attended work. An operational task SHALL have no such round.

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

#### Scenario: Deploy run that does not come from the plan

- **GIVEN** a deploy run that carries tickets of several work items
- **WHEN** it reaches a non-production environment
- **THEN** it asks the authorization for that environment, whatever a plan's classes answered

#### Scenario: Effect outside every class

- **GIVEN** an effect the plan did not foresee
- **WHEN** a task needs it in attended work
- **THEN** the agent asks it as its own question

### Requirement: Unattended work may spend an amount for unforeseen effects

At every activation of unattended delegation, the activation message SHALL state the amount for unforeseen spend the user gave, or ask it. That question SHALL NOT count among the open answers that keep the work on hold, and no answer SHALL mean zero. The answers in force SHALL include the classes of effect the user approved in the plan, each ceiling counting as authorized spend for its target; the effects of a refused class SHALL be skipped and reported; an unanswered class SHALL add no authorization and SHALL NOT keep the work on hold, and each of its effects SHALL be decided as any other effect of unattended work. Unattended work SHALL take a spending effect no answer covers only when an amount is in force, the resource or setting change can be undone by a command recorded before it is taken, and the total of such spend, counted at the provider's price until the delegation's expiry or at the effect's maximum, stays within the amount; when neither the price nor a maximum can be read, or the effect keeps charging after the expiry without a known maximum that fits, it SHALL skip the effect; money already spent SHALL count against the amount and SHALL NOT make the effect irreversible. It SHALL record the effect before taking it and list it in the end report with how to revert it. Otherwise it SHALL skip the effect and report it. The forbidden actions of unattended delegation SHALL NOT change.

#### Scenario: Small unforeseen spend within the amount

- **GIVEN** unattended work with an amount in force and a fix whose change a recorded command undoes and whose cost until expiry is less than what remains of the amount
- **WHEN** the work needs that fix
- **THEN** it records the target, the change and how to revert it, takes the effect and continues
- **AND** the end report lists it

#### Scenario: No amount stated

- **GIVEN** unattended work whose user stated no amount and did not answer the question
- **WHEN** the work needs an unforeseen spend
- **THEN** it keeps working on what does not depend on that effect, skips it and reports it

#### Scenario: Activation by silence

- **GIVEN** a user who falls silent after answering the route, the delivery and the review
- **WHEN** unattended delegation activates
- **THEN** the activation message asks the amount and the work continues with an amount of zero

#### Scenario: Change that cannot be undone

- **GIVEN** unattended work with an amount in force
- **WHEN** it needs an effect no recorded command can undo
- **THEN** it skips the effect and reports it, whatever its cost
