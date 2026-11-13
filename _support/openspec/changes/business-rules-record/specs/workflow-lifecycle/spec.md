# workflow-lifecycle

## ADDED Requirements

### Requirement: Business rules a work item adopts are recorded and reach its ticket

A business rule SHALL be a behavior of the product in a situation: what an action requires, who may take it, which state follows it, what it blocks or keeps, or where a setting lives. Each rule a work item decides or assumes that its ticket and the documentation the work read do not state, in any flow, SHALL be recorded in the work's rule record in the same turn: one line naming the rule, the screen it shapes when it has one, its status, and the flow that found it. A rule adopted without the user's decision, by the main thread or by a child that reports it, SHALL be recorded as an assumption and stated in the main thread's next message. A rule that replaces one the ticket, the documentation or the record states SHALL keep the earlier rule's line, marked cancelled with the date, the reason and the rule that replaces it; no line SHALL be deleted. The record SHALL live in the example's record of the design stage (`examples.md`) while the work has no change folder, in the proposal of its change folder once it has one, which then governs, and otherwise in the work's checkpoint file. The main thread SHALL write the record's rules to the work's ticket when the user approves a design example, when the build starts and at the close of a work item that was integrated: a section of the description that it alone replaces, leaving the rest of the description unchanged, holding the rules decided, the assumptions, the rules still to decide and the rules cancelled with their reasons, and a comment naming what that write added and what it cancelled, with each reason. It SHALL skip a write when the record did not change since the last one, and SHALL write nothing to a tracker when the work has no ticket or was abandoned; a close that writes nothing SHALL list the rules in its report. These writes SHALL take no question, as an exception the guidance states to its rule for effects on data the task did not create. Research SHALL read a ticket with its comments and SHALL NOT write the ticket.

#### Scenario: A rule decided during the plan

- **GIVEN** a ticket that does not say whether a search matches by plate or by any field
- **WHEN** the user decides it in the plan and later starts the build
- **THEN** the rule is a line of the proposal's rule record in the turn the user answered
- **AND** the build's start writes it in the ticket's rules section and adds a comment naming it

#### Scenario: A decision replaces a rule the ticket states

- **GIVEN** a ticket whose description says an evidence can satisfy a required document
- **WHEN** the user decides during the plan that it cannot
- **THEN** the record keeps the ticket's rule as a cancelled line with the date, the reason and the rule that replaces it
- **AND** the next write lists it among the cancelled rules of the section, leaves the original sentence of the description unchanged, and its comment says what was cancelled and why

#### Scenario: A child assumes a rule during the build

- **GIVEN** a build child that had to choose which stage a new record starts in
- **WHEN** it reports that choice
- **THEN** the main thread records it as an assumption and states it in its next message
- **AND** the child writes neither the record nor the ticket

#### Scenario: Work without a ticket

- **GIVEN** a work item that names no ticket
- **WHEN** it reaches its close with rules in its record
- **THEN** no tracker is written and the close's report lists the rules

#### Scenario: Nothing changed since the last write

- **GIVEN** a record whose last ticket write was at the build's start
- **WHEN** the close finds no line added or changed since then
- **THEN** it writes neither the section nor a comment

### Requirement: Rule cards beside each frame

`flow-figma-design` SHALL draw, beside each frame it draws and outside the frame's bounds, an annotation card marked as not being interface, with one row per rule of that screen whose condition or consequence the frame's visible text does not state, each row naming its origin: stated by a source, decided, assumed or still to decide. A frame with no such rule SHALL get no card. The stage SHALL draw the cards after it reconciles the first review, so that the reviewer first lists unsettled rules without them, and SHALL have them reviewed in the one follow-up review. It SHALL name each card's node in the example's record and SHALL update a card's row in the same turn its rule is answered. When a later flow changes a rule a card shows, the main thread SHALL correct the row, keeping the earlier rule as cancelled with its reason, when the host has a connection that can write to Figma, and otherwise SHALL record the card as stale. A frame a later decision contradicts SHALL be recorded as stale and listed among what the work leaves open at its close; it SHALL NOT be redrawn outside the design stage.

#### Scenario: A frame whose rule the screen does not state

- **GIVEN** a list frame whose rows hide the delete action for records in use
- **WHEN** the stage finishes its first review
- **THEN** a card beside the frame carries that rule with its origin
- **AND** the record names the card's node

#### Scenario: A frame with nothing to annotate

- **GIVEN** a frame whose every rule is written in its visible text
- **WHEN** the stage draws the cards
- **THEN** that frame gets no card

#### Scenario: The reviewer reads the cards

- **GIVEN** cards drawn after the first review
- **WHEN** the follow-up review runs
- **THEN** it reports a row that repeats the frame's text, a row the frame contradicts and a recorded rule missing from its card

#### Scenario: The build changes a rule a card shows

- **GIVEN** an approved example whose card says a claim starts in the first stage as open
- **WHEN** the user decides during the build that it starts in process
- **THEN** on a host connected to Figma the card keeps the earlier row as cancelled with its reason and adds the new one
- **AND** on a host without that connection the record marks the card stale

#### Scenario: A decision contradicts an approved frame

- **GIVEN** an approved frame that shows a link between an evidence and a document
- **WHEN** the plan decides that link does not exist
- **THEN** the record marks the frame stale, no flow outside the design stage redraws it, and the close lists it among what the work leaves open

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
- **THEN** it searches every declared tracker for an existing one and creates the new ticket in the primary tracker.

### Requirement: Screen proposals in Figma before the plan

Screen examples SHALL be drawn by a stage of their own, `flow-figma-design`, that runs between research and the plan and that only the user's reply starts, except the continuation from research for a request whose deliverable is the design. When the host has a connection that can write to Figma, research SHALL end with the next-step block that recommends that stage, with going straight to the plan, or to the build for a small understood change, as other replies, when what gets built includes a new screen, even when sibling screens fix its composition and the change would otherwise continue into the build, and when it changes an existing screen by adding, removing or moving a region and neither the request nor an existing screen fixes that composition; this SHALL hold when research assessed a ticket the user picked as well as when the user asked for a change. Research SHALL NOT mention the stage for a change that keeps the screen's layout, on a host with no connection that can write to Figma, or under unattended delegation, and SHALL make no offer of proposals through a question. The user's reply that starts the stage SHALL be the instruction for the writes it makes to that Figma file, limited to what the work creates, and SHALL NOT replace the question a critical effect takes. A later flow's correction of a rule card's row, as «Rule cards beside each frame» sets, SHALL need no such reply: the guidance's standing permission covers it, limited to the card nodes the work created. On a host with no connection that can write to Figma, the stage SHALL draw nothing and report what is missing. The stage SHALL load the Figma connection's own usage instructions, when it ships them, the project's design system, the sibling screens of the same module or navigation in the file and in the code, the decisions already recorded for the app in its design record, and the record of the example each sibling was approved from when a search for the file's location and its frame node in the project's change folders, archived ones included, and in the example records under the project's support workspace finds one, before it draws. It SHALL draw in the Figma file the user named in the conversation or the app's design record declares; a declared file of another design tool SHALL NOT count, and a declared file that neither the block that offered the stage nor the user named SHALL be confirmed by the user once before the first write. When the connection cannot write to that file, the stage SHALL draw nothing and report it. With neither file, it SHALL ask the user for the file once, in the text, and end with the route to the plan when the user gives none. Before the first drawing it SHALL tell the user each step they will have to take by hand in Figma. It SHALL draw one example per screen, following its sibling screens, and SHALL draw two or three examples only for a screen the user names when asking for alternatives, as frames on a page named after the work, and SHALL change or delete only the nodes the work created; it MAY create components and variables local to the file, SHALL record their identifiers with the example, and SHALL remove those only the unchosen examples used; a change to an existing variable, style or component, or to a shared library, SHALL return to the user as a decision. The main thread SHALL draw the frames; no child SHALL draw them. When the connection runs scripts in the file, each write SHALL define once the pieces its frames use, built on the tool's automatic layout and bound to the file's variables, with the app shell cloned from an existing screen, and SHALL compose the frames from them; when the app's design record declares a Figma design library, each piece and variable that library holds SHALL come from it, as an instance of its component or a binding to its variable, a piece or variable SHALL be defined only for what the library lacks, each such piece SHALL be listed in the example's record as a candidate for the library and named in the approval message, adding the library to the file SHALL be among the steps the user takes by hand, and when the file cannot use the library the stage SHALL draw with its own pieces and say so in the approval message; a correction SHALL change a piece or the data and draw the frame again, not move or resize nodes one by one. Examples SHALL be drawn at one desktop width, the one the screen's existing frames in the file use, otherwise 1440 px. Before the user is asked, one `hive-review-ux` child SHALL review the first drawing, and one the 440 px frame, against the request, the sibling screens, the recorded decisions, the declared design library and their alignment, without the drawer's reasoning; without a connection that reads the file, it SHALL judge from the frames' images and report the alignment measurements as not run, and the stage SHALL say so in the approval message. The main thread SHALL fix the findings that hold, have only the corrected frames and the rule cards reviewed again once, update the record's status to pending the user's approval, and name what remains in the message to the user. A correction the user asks for SHALL NOT get another review. The user SHALL approve or correct each screen's example as one decision per screen, in one round, beside the decisions on that screen's rules that «Business rules that appear while drawing» sets, choosing among the alternatives of a screen drawn with them, and SHALL approve the 440 px frame of a screen served on mobile before it is recorded. When the user's correction sets a convention that applies beyond that screen, the stage SHALL ask once, in the round that asks the approval, whether to record it among the decisions of the app's design record, and SHALL keep a yes in the example's record without editing the design record itself. The approved example SHALL be recorded in the report and in a note of the work's support folder with the file's location, each frame's node, a local image and its status; that note SHALL be written as soon as the frames are drawn, with a status saying they are under review, before the review and before the user is asked, and its status SHALL be updated after the review and when the user approves. The record SHALL hold one status per screen, and a plan SHALL treat only a screen whose status is approved as having an approved example; the alternatives not chosen SHALL be removed, and the approved examples of the other screens SHALL stay. When the file is not declared in the app's design record, the stage SHALL ask once, in the round that asks the approval, whether to record it there, and SHALL keep a yes in the example's record without editing the design record itself. For a request whose deliverable is the design, research SHALL continue into the stage in the same turn when the host has a connection that can write to Figma, and SHALL report the missing connection as a blocker when it has none. Once every screen is approved, the stage SHALL end with the next-step block toward the plan, or toward the build only when the block that offered the stage listed the build, the approved examples add no region the request did not name, the record holds no yes to recording the file or a convention, and no rule in it is still proposed, except when a rule changed the scope, which ends as «Business rules that appear while drawing» sets, and except for a request whose deliverable is the design, which ends as this requirement sets for it. A request whose deliverable is the design SHALL be drawn by the same stage, in the page or frames it names when it names them, changing and never deleting those frames; it SHALL NOT be reported as done before the user approves the example, and after that approval the stage SHALL name as pending each yes to recording the file or a convention in the design record, since no plan follows to write it, and SHALL run the close, which treats it as an operational task. Under unattended delegation the stage SHALL ask neither for the file nor for the confirmation of a declared one, SHALL draw nothing and report the file as unresolved when it has none it may use, the choice or the approval of an example SHALL stay unanswered and be reported as unresolved, every example SHALL be kept, and the plan SHALL settle the UI without an approved example. A plan with a change folder SHALL carry the approved example's record in its `UI` section and that record's rule lines in the rule record of its proposal, and work without one SHALL carry both in its checkpoint file. A plan for the screen SHALL use the approved example as the after sketch of the regions it shows, the plan SHALL add the task that writes each convention the record holds a yes for, a build without a plan SHALL keep the record in its checkpoint file, and the build's first view, the verification walk and the UX review SHALL compare the rendered screen against it. A difference SHALL be a region, a component, an action's placement or an order that differs from the example, not rendering variance, and not what the plan records as decided otherwise or a correction the user asked for after approving the example, which the example's record SHALL name; it SHALL be handled as a UI defect the change introduced, and the differences that remain SHALL be listed in the validation message when the delivery stops for validation and in the final report when it does not. When a frame's image is missing where the comparison runs, the frame SHALL be read again from its recorded location and node, and a comparison that cannot run SHALL be reported as not run. The user's validation in verification SHALL be asked as before.

#### Scenario: A ticket with a new screen and a Figma connection

- **GIVEN** a host whose connection can write to Figma
- **WHEN** research assesses a ticket whose work includes a new screen, whether or not sibling screens fix its composition
- **THEN** its next-step block recommends `flow-figma-design` and lists going straight to the plan as another reply
- **AND** it asks no question that offers proposals

#### Scenario: A requested change for a new screen whose siblings fix its layout

- **GIVEN** a host whose connection can write to Figma, and a requested new list screen whose sibling screens fix its layout
- **WHEN** research finds the how leaves no decision for the user
- **THEN** it ends with the next-step block that recommends `flow-figma-design`, with the build as another reply, instead of continuing into the build

#### Scenario: No Figma connection

- **GIVEN** a host with no connection that can write to Figma
- **WHEN** research reaches a new screen
- **THEN** it does not mention the stage and routes to the plan, which draws its text sketches

#### Scenario: A change that keeps the layout

- **GIVEN** a request to change a button, a copy, a field or a column of an existing screen
- **WHEN** research settles what the work is
- **THEN** it does not mention the stage

#### Scenario: One example that follows its siblings

- **GIVEN** a new list screen in a module whose other screens already exist in the Figma file
- **WHEN** the stage draws it without the user asking for alternatives
- **THEN** the main thread draws one example whose shell, header, toolbar, table and row actions follow those sibling screens
- **AND** no child draws it

#### Scenario: Alternatives the user asks for

- **GIVEN** a flow of several screens
- **WHEN** the user asks for alternatives for one screen they name
- **THEN** the stage draws two or three examples for that screen and one for each of the others

#### Scenario: Review before the user sees the frames

- **GIVEN** frames the main thread has drawn
- **WHEN** the stage is about to ask the user
- **THEN** the record already lists the frames, written when they were drawn
- **AND** one `hive-review-ux` child has reviewed them against the request, the sibling screens, the recorded decisions and their alignment
- **AND** the findings that hold are fixed, and the ones that remain are named in the message to the user

#### Scenario: A convention the user corrects

- **GIVEN** a user who corrects the row actions of an example to icon buttons with a tooltip, for every list of the app
- **WHEN** the stage asks the approval
- **THEN** it asks in the same round whether to record that convention among the decisions of the app's design record
- **AND** on a yes the plan for that screen includes the task that writes it

#### Scenario: The design file is not declared

- **GIVEN** an app whose design record declares no design file, and a user who has named none
- **WHEN** the stage starts
- **THEN** it asks the user once for the file
- **AND** when the user gives none, it ends with the route to the plan

#### Scenario: The request is the design

- **GIVEN** a request to draw or replicate a screen in Figma
- **WHEN** the stage draws the example
- **THEN** it asks the user to approve or correct the example before reporting the work as done
- **AND** after the approval it runs the close

#### Scenario: The user approves an example

- **GIVEN** examples drawn as frames on the work's page
- **WHEN** the user approves the example of each screen, choosing one alternative for a screen drawn with them
- **THEN** its record holds the file's location, each frame's node, a local image and the approved status
- **AND** only the alternatives not chosen are removed

#### Scenario: The built screen differs from the approved example

- **GIVEN** work whose record holds an approved example
- **WHEN** the verification walk captures a region the example shows
- **THEN** each difference in a region, a component, an action's placement or an order is handled as a UI defect the change introduced
- **AND** the differences that remain are listed in the validation message, or in the final report when the delivery does not stop for validation

#### Scenario: Unattended delegation is active

- **GIVEN** work that runs under unattended delegation
- **WHEN** research reaches a new screen
- **THEN** it does not recommend the stage and the plan settles the UI

### Requirement: Business rules that appear while drawing

A rule SHALL be a business rule, as «Business rules a work item adopts are recorded and reach its ticket» defines it, that a frame shows. A rule SHALL be settled when the request, the ticket's criteria, research's decisions as the conversation, the ticket or the work's record holds them, the decisions recorded for the app, or the behavior its sibling screens or their code already implement settle it. While `flow-figma-design` draws, it SHALL list each rule its frames show that none of these settle, draw the recommended behavior, and record each in the example's record as one line naming the rule, the screen it shapes, its status, proposed, assumption, scope, or approved with the date and the answer, and whether the drawer or the reviewer found it. A rule whose options all leave the scope, a contract, stored data, the cost, reversibility and who may act unchanged SHALL be stated as an assumption in the approval message and recorded with that status, not asked. A rule that adds a capability the request did not name, changes data or settings that other screens or modules use, or changes a permission SHALL be recorded as a scope rule and SHALL NOT be asked in the approval round: the stage SHALL record each screen it shapes as stopped by that rule, delete what it drew of them, finish the record and the review for the other screens, ask their approval, and end with the next-step block toward research naming those rules instead of the block toward the plan; when the stage runs again for that work, it SHALL resume from the existing record before drawing, record each scope rule research settled as approved with its answer, draw only the screens recorded as stopped, and keep the status of the others. The route to the build SHALL require that no rule is still proposed. For a request whose deliverable is the design, such a rule SHALL leave the work not done and end the same way. In the round that asks the approval, each proposed rule SHALL be its own decision whose answer is recorded in the turn it is given, and the screen's approve option SHALL be worded to hold under every answer to it, naming that decision; approving a screen SHALL NOT approve its rules, and an answer that differs from the drawn behavior SHALL redraw that part and keep the screen pending, even when its approve option was chosen, until the user approves it as redrawn. The UX reviewer of the frames SHALL receive among its requirements the request, the ticket's criteria, research's decisions and the app's recorded decisions, not the drawer's list of proposed rules, and SHALL report each behavior that none of them, the app's recorded decisions or the sibling screens settle apart from its findings, without failing its verdict; the stage SHALL check each one its record lacks against its own sources and the siblings' code, drop those they settle, and record the rest as it records its own rules. A plan SHALL take each rule the record marks as approved as one of its decisions and each assumption as a minor call it states, written in the rule record of its proposal, SHALL ask each rule still proposed before the plan is approved, and SHALL NOT decide a scope rule research has not settled: it SHALL name it and end with the next-step block toward research; a scope rule research settled SHALL be one of its decisions, and the plan SHALL sketch the screens it stopped. Under unattended delegation the rules SHALL stay as recorded and be reported as unresolved, a scope rule SHALL still leave the screens it shapes undrawn, the stage SHALL NOT end toward research, and the stopped screens SHALL stay unresolved while the plan covers the others.

#### Scenario: A rule the screen decides

- **GIVEN** a rejection dialog whose comment field nothing settles as required or optional
- **WHEN** the stage asks the approval of that screen
- **THEN** the same round asks whether the comment is required as a decision of its own
- **AND** the screen's approve option names that decision and holds under either answer

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
