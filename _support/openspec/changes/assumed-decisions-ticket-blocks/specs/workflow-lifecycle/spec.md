## ADDED Requirements

### Requirement: Clear-winner decisions are taken and listed

An open decision of what a work includes or how it is done, in research, in a plan or as a late fact of a build, SHALL be taken without asking when none of the criteria below holds and the reason for its recommended option is a consequence the other options lack; it SHALL be listed in the next message under a heading such as «What was not defined and I decided», one line each with the decision, the alternative discarded and how it is reversed, at most five lines, the rest approved in one question of the same round. It SHALL be asked, through the question channel the guidance sets, when at least one holds: it adds or removes behavior, touches another module or splits the work against what the ticket or the request says; the recommended option costs less now while another avoids rework or migration later; it sets a business rule or a preference that no behavior the ticket, the specs or the user's words name requires; it is hard to reverse, such as a public contract, stored data, permissions or security, money or production; the recommendation rests on an unverified fact, or no consequence separates the options; or it contradicts the user, the ticket or the specs. The user's reply to that message's card or next-step block SHALL accept each listed decision it does not correct, the message SHALL say so in one line, and an accepted listed decision SHALL authorize no effect and SHALL NOT turn an assumed business rule into an approved one. The rounds other rules require SHALL keep their questions: effects and critical effects, the session questions, the validation stop, findings rounds, the approval and proposed rules of a design in Figma, a plan's choice of approach, the rounds of ticket and requirement-document creation, an unanswered card, the question about an external project and the test-infrastructure decision of a plan.

#### Scenario: A clear winner is listed, a scope question is asked

- **GIVEN** a plan with an open choice of where a rule lives, whose recommended place has a consequence the others lack and meets no criterion, and an open choice of adding a rate limit the ticket does not name
- **WHEN** the agent reports the plan's decisions
- **THEN** the first appears in the list with the place discarded and how it is reversed
- **AND** only the second goes through the card

#### Scenario: The user corrects one listed decision

- **GIVEN** a message that lists three decisions and ends with the next-step block
- **WHEN** the user replies with the flow command and a correction to the second
- **THEN** the first and the third stand, and the second takes the user's answer

### Requirement: A chosen ticket's assessment is written in blocks

When the user picks a ticket for research, the assessment SHALL be written in this order: a summary of impact, complexity and dependencies; a verdict that says whether the ticket is ready and names each check of the create-issue ready list it fails with what is missing; the proposed acceptance criteria when the ticket lacks them; the decisions taken and listed as «Clear-winner decisions are taken and listed» sets; the questions that requirement sends to the user, through the card; and the ending. A ticket SHALL be not ready while it fails any check of that ready list or the specs contradict it, and its scope SHALL be settled in research, never left for the plan to decide what enters. When the findings confirm work the ticket asks for, the ending SHALL be the next-step block, closing the same message when nothing is asked and the message that follows the answers otherwise; when they do not, the ending SHALL be the block of replies the report endings set.

#### Scenario: A defect ticket without criteria

- **GIVEN** a ticket that reports a defect with a recording and no acceptance criteria, and leaves a rate limit open
- **WHEN** the agent assesses it
- **THEN** the verdict reads not ready and names the failed checks for criteria, out of scope and the open question
- **AND** the criteria that follow from the defect are proposed and listed, and the rate limit is asked through the card in the same message

#### Scenario: A ready ticket

- **GIVEN** a ticket whose ready list passes and whose findings confirm the work
- **WHEN** the agent assesses it
- **THEN** the verdict reads ready and the message ends with the next-step block

## MODIFIED Requirements

### Requirement: Business rules that appear while drawing

A rule SHALL be a business rule, as «Business rules a work item adopts are recorded and reach its ticket» defines it, that a frame shows. A rule SHALL be settled when the request, the ticket's criteria, research's decisions as the conversation, the ticket or the work's record holds them, the decisions recorded for the app, or the behavior its sibling screens or their code already implement settle it. While `flow-figma-design` draws, it SHALL list each rule its frames show that none of these settle, draw the recommended behavior, and record each in the example's record as one line naming the rule, the screen it shapes, its status, proposed, assumption, scope, or approved with the date and the answer, and whether the drawer or the reviewer found it. A rule for which none of the criteria of «Clear-winner decisions are taken and listed» holds SHALL be stated as an assumption in the approval message, listed as that requirement sets, and recorded with that status, not asked. A rule that adds a capability the request did not name, changes data or settings that other screens or modules use, or changes a permission SHALL be recorded as a scope rule and SHALL NOT be asked in the approval round: the stage SHALL record each screen it shapes as stopped by that rule, delete what it drew of them, finish the record and the review for the other screens, ask their approval, and end with the next-step block toward research naming those rules instead of the block toward the plan; when the stage runs again for that work, it SHALL resume from the existing record before drawing, record each scope rule research settled as approved with its answer, draw only the screens recorded as stopped, and keep the status of the others. The route to the build SHALL require that no rule is still proposed. For a request whose deliverable is the design, such a rule SHALL leave the work not done and end the same way. In the round that asks the approval, each proposed rule SHALL be its own decision whose answer is recorded in the turn it is given, and the set's approve option SHALL be worded to hold under every answer to it, naming that decision; approving the set SHALL NOT approve its rules, and an answer that differs from the drawn behavior SHALL redraw that part and keep the screen pending, even when the set's approve option was chosen, until the user approves it as redrawn. The UX reviewer of the frames SHALL receive among its requirements the request, the ticket's criteria, research's decisions and the app's recorded decisions, not the drawer's list of proposed rules, and SHALL report each behavior that none of them, the app's recorded decisions or the sibling screens settle apart from its findings, without failing its verdict; the stage SHALL check each one its record lacks against its own sources and the siblings' code, drop those they settle, and record the rest as it records its own rules. A plan SHALL take each rule the record marks as approved as one of its decisions and each assumption as a decision it lists as taken, written in the rule record of its proposal, SHALL ask each rule still proposed before the plan is approved, and SHALL NOT decide a scope rule research has not settled: it SHALL name it and end with the next-step block toward research; a scope rule research settled SHALL be one of its decisions, and the plan SHALL sketch the screens it stopped. Under unattended delegation the rules SHALL stay as recorded and be reported as unresolved, a scope rule SHALL still leave the screens it shapes undrawn, the stage SHALL NOT end toward research, and the stopped screens SHALL stay unresolved while the plan covers the others.

#### Scenario: A rule the screen decides

- **GIVEN** a rejection dialog whose comment field nothing settles as required or optional
- **WHEN** the stage asks approval of the set containing that screen
- **THEN** the same round asks whether the comment is required as a decision of its own
- **AND** the set's approve option names that decision and holds under either answer

#### Scenario: A rule the app already settles

- **GIVEN** a list screen whose sibling screens already sort by the last change, newest first
- **WHEN** the stage draws the new list with the same order
- **THEN** it asks no rule decision for that order

#### Scenario: A rule that changes the scope

- **GIVEN** a workflow wizard whose review step would hold a company-wide approval policy
- **WHEN** the stage finds, while drawing or in the reviewer's report, that this rule changes a setting other workflows use
- **THEN** it records the rule as a scope rule, leaves the screens it shapes undrawn and asks the approval of the others, without asking that rule
- **AND** it ends with the next-step block toward research naming that rule

#### Scenario: The reviewer finds a rule the drawer missed

- **GIVEN** frames whose review brief lists the request, the ticket's criteria, research's decisions and the app's recorded decisions
- **WHEN** a frame shows a behavior none of them or the sibling screens settle
- **THEN** the reviewer reports it apart from its findings without failing its verdict
- **AND** the stage records it as proposed and asks it in the approval round

#### Scenario: The plan reads the rules

- **GIVEN** an example's record with one rule approved and one still proposed
- **WHEN** the plan settles the screen's UI
- **THEN** the approved rule is a decision of the plan and the proposed one is asked before the plan is approved
