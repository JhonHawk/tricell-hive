## MODIFIED Requirements

### Requirement: Screen proposals in Figma before the plan

Screen examples SHALL be drawn by a stage of their own, `create-design`, that runs between research and the plan and that only the user's reply starts, except the continuation from research for a request whose deliverable is the design. When the host has a connection that can write to Figma and the project's Figma file is known, because the user named it in the conversation, the app's design record declares it, or the project's persistent memory records it, research SHALL end with the next-step block that recommends that stage, naming that file and putting its link in the command's argument, with going straight to the plan, or to the build for a small understood change, as other replies, when what gets built includes a new screen, even when sibling screens fix its composition and the change would otherwise continue into the build, and when it changes an existing screen by adding, removing or moving a region and neither the request nor an existing screen fixes that composition; this SHALL hold when research assessed a ticket the user picked as well as when the user asked for a change. Research SHALL NOT mention the stage for a change that keeps the screen's layout, on a host with no connection that can write to Figma, when no Figma file of the project is known, or under unattended delegation, and SHALL make no offer of proposals through a question. The user's reply that starts the stage SHALL be the instruction for the writes it makes to that Figma file, limited to what the work creates, and SHALL NOT replace the question a critical effect takes. A later flow's correction of a rule card's row, as «Rule cards beside each frame» sets, SHALL need no such reply: the guidance's standing permission covers it, limited to the card nodes the work created. On a host with no connection that can write to Figma, the stage SHALL draw nothing and report what is missing. The stage SHALL load the Figma connection's own usage instructions, when it ships them, the project's design system, the sibling screens of the same module or navigation in the file and in the code, the decisions already recorded for the app in its design record, and the record of the example each sibling was approved from when a search for the file's location and its frame node in the project's change folders, archived ones included, and in the example records under the project's support workspace finds one, before it draws. It SHALL draw in the Figma file the user named in the conversation, the research block that offered the stage named, or the app's design record declares; a declared file of another design tool SHALL NOT count, and a declared file that neither the block that offered the stage nor the user named SHALL be confirmed by the user once before the first write. When the connection cannot write to that file, the stage SHALL draw nothing and report it. With neither file, it SHALL ask the user for the file once, in the text, and end with the route to the plan when the user gives none. Before the first drawing it SHALL tell the user each step they will have to take by hand in Figma. It SHALL draw one example per screen, following its sibling screens, and SHALL draw two or three examples only for a screen the user names when asking for alternatives, as frames on a page named after the work, and SHALL change or delete only the nodes the work created; it MAY create components and variables local to the file, SHALL record their identifiers with the example, and SHALL remove those only the unchosen examples used; a change to an existing variable, style or component, or to a shared library, SHALL return to the user as a decision. The main thread SHALL draw the frames; no child SHALL draw them. When the connection runs scripts in the file, each write SHALL define once the pieces its frames use, built on the tool's automatic layout and bound to the file's variables, with the app shell cloned from an existing screen, and SHALL compose the frames from them; the stage SHALL bind one Figma design library before it draws: the one the user named or shared in the conversation, otherwise the one the app's design record declares, otherwise one it detects, which is a library the file has enabled as the connection lists it, a library whose components the sibling frames instance or whose variables they bind, or a library the example record of a sibling names; a link the user gives to a Figma file other than the one the stage draws in SHALL count as the library the user named when the user presents it as a library, a kit, a design system or the source of the components, or when the connection lists it among the libraries that file has enabled, and SHALL NOT count when the user links it as the screen to replicate or as a reference; among several detected libraries the stage SHALL bind the one the sibling frames instance, and SHALL ask the user once, in the text, only when that leaves none or more than one, or when the user named more than one. Before the first drawing the stage SHALL write in the example's record an inventory, keeping every line of a record that already exists except its searches when the bound library changed: each component and variable the sibling frames instance or bind, from any library or from the file itself, and, for each piece the work's screens need that no sibling uses, the terms it searched in the bound library and what the search found. Each piece and variable the bound library holds SHALL come from it, as an instance of its component or a binding to its variable, a piece it lacks that the siblings use SHALL come from the siblings' component or variable, whether or not a library is bound, a piece or variable SHALL be defined only for what the inventory shows the library and the siblings lack, each such piece SHALL be listed in the example's record as a candidate for the library and named in the approval message, adding the library to the file SHALL be among the steps the user takes by hand, and when the file cannot use the library the stage SHALL draw with its own pieces and say so in the approval message; a correction SHALL change a piece or the data and draw the frame again, not move or resize nodes one by one. Examples of a screen of a web app SHALL be drawn at one desktop width, the one the screen's existing frames in the file use, otherwise 1440 px, and, when the screen is served on mobile, which a responsive layout such as a breakpoint or a viewport meta tag shows, also at 440 px. Examples of a screen of an app that runs only on mobile, which a native app or a phone shell with no desktop layout in the code and the sibling frames shows even when the app has a viewport meta tag or breakpoints, SHALL be drawn once, at 440 px, whatever width the file's existing frames use, and SHALL get no second frame, and its verification SHALL capture and compare the 440 px viewport alone. With neither code nor sibling frames, the platform the request or the ticket names SHALL decide; with none, the screen SHALL be drawn as a web screen with its 440 px frame, and the approval message SHALL say so. When the work has more than one screen, the stage SHALL first draw, at that width only, the screen that uses the most pieces of the inventory, have one `hive-review-ux` child review it alone, fix the findings that hold, reconcile the unsettled rules it lists as it reconciles those of the set's review and give the user its location without asking anything and without ending the turn, and only then draw the other screens and every 440 px frame from its pieces; that review SHALL get no rule cards and no follow-up, and the first review the rule cards wait for SHALL be the one of the whole set. Before the user is asked, one `hive-review-ux` child SHALL review the drawn frames, the first screen's included, and one the 440 px frame, against the request, the sibling screens, the recorded decisions, the bound design library and their alignment, without the drawer's reasoning; the reviewer SHALL read each sibling's image as well as its node tree, SHALL report as a finding a piece the inventory shows the siblings instance that is drawn other than as an instance, SHALL search the bound library for each candidate piece and report as a finding one the library holds, SHALL NOT pass the frames, in a review that is not a follow-up, while a sibling's image or a library search that the connection lets it run was not run, and SHALL report one the connection cannot run as not run, which the stage SHALL name in the approval message; without a connection that reads the file, it SHALL judge from the frames' images and report the alignment measurements as not run, and the stage SHALL say so in the approval message. The main thread SHALL fix the findings that hold, have only the corrected frames and the rule cards reviewed again once, update the record's status to pending the user's approval, and name what remains in the message to the user. A correction the user asks for SHALL NOT get another review. The stage SHALL finish drawing and reviewing the completed set, including each applicable 440 px mobile frame, before asking one decision to approve the set or request corrections naming affected screens, alongside the separate proposed rule decisions that «Business rules that appear while drawing» sets. It SHALL name the recommended example for each screen with alternatives, and the user may name a different alternative with corrections. Corrections SHALL leave only the named screens pending and approve the others; after redrawing, one approval decision SHALL cover only the affected screens and their mobile frames, preserving the others' approval. When the user's correction sets a convention that applies beyond that screen, the stage SHALL ask once, in the round that asks the approval, whether to record it among the decisions of the app's design record, and SHALL keep a yes in the example's record without editing the design record itself. The approved example SHALL be recorded in the report and in a note of the work's support folder with the file's location, each frame's node, a local image and its status; that note SHALL be created with the inventory before the first drawing, and its frame and status lines SHALL be written as soon as the frames are drawn, with a status saying they are under review, before the review and before the user is asked, and its status SHALL be updated after the review and when the user approves. The record SHALL hold one status per screen, and a plan SHALL treat only a screen whose status is approved as having an approved example; the alternatives not chosen SHALL be removed, and the approved examples of the other screens SHALL stay. When the app's design record does not declare the Figma file the stage drew in or the library it bound, or declares another Figma file or library, the stage SHALL ask once, in the round that asks the approval, whether to record them there, showing each value as it will be written beside the value it replaces; a field that names a file or a library of another design tool SHALL stay as it is and SHALL NOT be asked about. A yes SHALL write nothing in the repository and SHALL ask no delivery question: the stage SHALL keep it in the example's record as pending, with each value as it will be written, and name it as pending; a no SHALL be kept in the example's record so the work does not ask again. For a request whose deliverable is the design, research SHALL continue into the stage in the same turn when the host has a connection that can write to Figma, and SHALL report the missing connection as a blocker when it has none. Once every screen is approved, the stage SHALL end with the next-step block toward the plan, or toward the build only when the block that offered the stage listed the build, the approved examples add no region the request did not name, the record holds no yes to recording a convention and no pending yes to recording the file or the library, and no rule in it is still proposed, except when a rule changed the scope, which ends as «Business rules that appear while drawing» sets, and except for a request whose deliverable is the design, which ends as this requirement sets for it. A request whose deliverable is the design SHALL be drawn by the same stage, in the page or frames it names when it names them, changing and never deleting those frames; it SHALL NOT be reported as done before the user approves the example, and after that approval the stage SHALL name as pending each yes to recording a convention in the design record, and each yes to recording the file or the library, since no plan follows to write it, and SHALL run the close, which treats it as an operational task. Under unattended delegation the stage SHALL ask neither for the file nor for the confirmation of a declared one, SHALL bind a library without asking, binding none when the conversation gives more than one or detection leaves none or more than one and still reusing what the siblings use, SHALL neither ask to record the file or the library nor write the design record, SHALL draw nothing and report the file as unresolved when it has none it may use, the choice or the approval of an example SHALL stay unanswered and be reported as unresolved, every example SHALL be kept, and the plan SHALL settle the UI without an approved example. A plan with a change folder SHALL carry the approved example's record in its `UI` section and that record's rule lines in the rule record of its proposal, and work without one SHALL carry both in its checkpoint file. A plan for the screen SHALL use the approved example as the after sketch of the regions it shows, the plan SHALL add the task that writes each convention the record holds a yes for, and the design file and the design library when the record holds a pending yes for them, a build without a plan SHALL keep the record in its checkpoint file, and the build's first view, the verification walk and the UX review SHALL compare the rendered screen against it. A difference SHALL be a region, a component, an action's placement or an order that differs from the example, not rendering variance, and not what the plan records as decided otherwise or a correction the user asked for after approving the example, which the example's record SHALL name; it SHALL be handled as a UI defect the change introduced, and the differences that remain SHALL be listed in the validation message when the delivery stops for validation and in the final report when it does not. When a frame's image is missing where the comparison runs, the frame SHALL be read again from its recorded location and node, and a comparison that cannot run SHALL be reported as not run. The user's validation in verification SHALL be asked as before.

#### Scenario: A ticket with a new screen and a Figma connection

- **GIVEN** a host whose connection can write to Figma
- **WHEN** research assesses a ticket whose work includes a new screen, whether or not sibling screens fix its composition
- **THEN** its next-step block recommends `create-design` and lists going straight to the plan as another reply
- **AND** it asks no question that offers proposals

#### Scenario: A requested change for a new screen whose siblings fix its layout

- **GIVEN** a host whose connection can write to Figma, and a requested new list screen whose sibling screens fix its layout
- **WHEN** research finds the how leaves no decision for the user
- **THEN** it ends with the next-step block that recommends `create-design`, with the build as another reply, instead of continuing into the build

#### Scenario: No Figma connection

- **GIVEN** a host with no connection that can write to Figma
- **WHEN** research reaches a new screen
- **THEN** it does not mention the stage and routes to the plan, which draws its text sketches

#### Scenario: A Figma connection but no known file

- **GIVEN** a host whose connection can write to Figma, and no Figma file the user named in the conversation, the app's design record declares or the project's persistent memory records
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

#### Scenario: One approval for ten screens

- **GIVEN** ten completed screens, including their applicable mobile frames
- **WHEN** the stage asks for approval
- **THEN** it asks one decision to approve or correct the set
- **AND** each proposed business rule is a separate decision in the same round

#### Scenario: Corrections to two screens

- **GIVEN** the user requests corrections to two screens in a set of ten
- **WHEN** the stage redraws them and asks for approval again
- **THEN** the decision covers only those two screens and their applicable mobile frames
- **AND** the other eight retain their approval

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
- **WHEN** the user approves the set of examples, with the recommended alternatives unless they name others
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

#### Scenario: A library shared in the conversation and not declared

- **GIVEN** an app whose design record declares no design library, and a user who gives a link to a second Figma file as the library that holds the app's components
- **WHEN** the stage draws a screen
- **THEN** each piece that file's library holds is drawn as an instance of its component
- **AND** the stage asks once, in the round that asks the approval, whether to record the file and the library in the design record

#### Scenario: A linked file to replicate is not a library

- **GIVEN** a user who links a second Figma file as the screen to replicate, and a design record that declares a design library
- **WHEN** the stage binds its library
- **THEN** it binds the declared library, not the linked file

#### Scenario: A library only the sibling frames reveal

- **GIVEN** an app whose design record declares no design library, a user who has named none, and sibling frames that instance components of one published library
- **WHEN** the stage reads the siblings
- **THEN** it binds that library without asking and says so before the first drawing
- **AND** it draws with instances of that library's components

#### Scenario: A table the library already holds

- **GIVEN** a bound library that holds a table header and a table cell, and a new screen with a table
- **WHEN** the stage writes the inventory before drawing
- **THEN** the inventory lists the header and the cell the siblings instance
- **AND** the screen's table is drawn with instances of them, with no candidate line for a table row

#### Scenario: The first screen before the set

- **GIVEN** a work of two screens and their mobile frames
- **WHEN** the stage draws
- **THEN** it draws one screen, has it reviewed, fixes the findings that hold and gives the user its location without asking anything
- **AND** only then does it draw the other frames, and it asks one approval for the completed set

#### Scenario: A design-only work records the file and the library

- **GIVEN** a request whose deliverable is the design, drawn in a file and with a library the design record does not declare
- **WHEN** the user answers yes to recording them
- **THEN** the stage writes nothing in the repository, asks no delivery question and keeps the yes in the example's record as pending
- **AND** the close names it as pending, since no plan follows to write it

#### Scenario: A yes to recording with no design record

- **GIVEN** an app with no design record
- **WHEN** the user answers yes to recording the file and the library
- **THEN** the stage writes nothing, keeps the yes in the example's record and names it as pending

#### Scenario: A screen of a mobile-only app

- **GIVEN** an app that runs only on mobile, whose existing frames in the Figma file are 390 px wide
- **WHEN** the stage draws its screens
- **THEN** it draws each screen once, at 440 px, and no frame at 390 px
- **AND** each flow's section holds its screens in one row, with no second frame below each one

#### Scenario: A web screen served on mobile

- **GIVEN** a web app whose screen is served on mobile and whose existing frames are 1440 px wide
- **WHEN** the stage draws the screen
- **THEN** it draws the screen at 1440 px and its 440 px frame below it

### Requirement: Business rules that appear while drawing

A rule SHALL be a business rule, as «Business rules a work item adopts are recorded and reach its ticket» defines it, that a frame shows. A rule SHALL be settled when the request, the ticket's criteria, research's decisions as the conversation, the ticket or the work's record holds them, the decisions recorded for the app, or the behavior its sibling screens or their code already implement settle it. While `create-design` draws, it SHALL list each rule its frames show that none of these settle, draw the recommended behavior, and record each in the example's record as one line naming the rule, the screen it shapes, its status, proposed, assumption, scope, or approved with the date and the answer, and whether the drawer or the reviewer found it. A rule for which none of the criteria of «Clear-winner decisions are taken and listed» holds SHALL be stated as an assumption in the approval message, listed as that requirement sets, and recorded with that status, not asked. A rule that adds a capability the request did not name, changes data or settings that other screens or modules use, or changes a permission SHALL be recorded as a scope rule and SHALL NOT be asked in the approval round: the stage SHALL record each screen it shapes as stopped by that rule, delete what it drew of them, finish the record and the review for the other screens, ask their approval, and end with the next-step block toward research naming those rules instead of the block toward the plan; when the stage runs again for that work, it SHALL resume from the existing record before drawing, record each scope rule research settled as approved with its answer, draw only the screens recorded as stopped, and keep the status of the others. The route to the build SHALL require that no rule is still proposed. For a request whose deliverable is the design, such a rule SHALL leave the work not done and end the same way. In the round that asks the approval, each proposed rule SHALL be its own decision whose answer is recorded in the turn it is given, and the set's approve option SHALL be worded to hold under every answer to it, naming that decision; approving the set SHALL NOT approve its rules, and an answer that differs from the drawn behavior SHALL redraw that part and keep the screen pending, even when the set's approve option was chosen, until the user approves it as redrawn. The UX reviewer of the frames SHALL receive among its requirements the request, the ticket's criteria, research's decisions and the app's recorded decisions, not the drawer's list of proposed rules, and SHALL report each behavior that none of them, the app's recorded decisions or the sibling screens settle apart from its findings, without failing its verdict; the stage SHALL check each one its record lacks against its own sources and the siblings' code, drop those they settle, and record the rest as it records its own rules. A plan SHALL take each rule the record marks as approved as one of its decisions and each assumption as a decision it lists as taken, written in the rule record of its proposal, SHALL ask each rule still proposed before the plan is approved, and SHALL NOT decide a scope rule research has not settled: it SHALL name it and end with the next-step block toward research; a scope rule research settled SHALL be one of its decisions, and the plan SHALL sketch the screens it stopped. Under unattended delegation the rules SHALL stay as recorded and be reported as unresolved, a scope rule SHALL still leave the screens it shapes undrawn, the stage SHALL NOT end toward research, and the stopped screens SHALL stay unresolved while the plan covers the others.

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

### Requirement: Rule cards beside each frame

`create-design` SHALL draw, beside each frame it draws and outside the frame's bounds, an annotation card marked as not being interface, with one row per rule of that screen whose condition or consequence the frame's visible text does not state, each row naming its origin: stated by a source, decided, assumed or still to decide. A frame with no such rule SHALL get no card. The stage SHALL draw the cards after it reconciles the first review, so that the reviewer first lists unsettled rules without them, and SHALL have them reviewed in the one follow-up review. It SHALL name each card's node in the example's record and SHALL update a card's row in the same turn its rule is answered. When a later flow changes a rule a card shows, the main thread SHALL correct the row, keeping the earlier rule as cancelled with its reason, when the host has a connection that can write to Figma, and otherwise SHALL record the card as stale. A frame a later decision contradicts SHALL be recorded as stale and listed among what the work leaves open at its close; it SHALL NOT be redrawn outside the design stage.

#### Scenario: A frame whose rule the screen does not state

- **GIVEN** a list frame whose rows hide the delete action for records in use
- **WHEN** the stage finishes its first review
- **THEN** a card beside the frame carries that rule with its origin
- **AND** the record names the card's node

#### Scenario: Alternatives reuse a stated rule

- **GIVEN** two alternatives for one screen whose specification states a rule neither frame says
- **WHEN** the first card has already recorded that rule as stated
- **THEN** both alternatives carry the rule in their cards and the record holds it once for that screen

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

### Requirement: Canvas sections hold what the design stage draws

`create-design` SHALL place every node it creates on a page inside a section it created, so that no frame, rule card, reference image or component of the work stays loose, except a rule card for a frame the work did not create, which stays beside that frame; when the tool cannot create a section, it SHALL draw without one and say so when it asks for approval. By default it SHALL group what it creates, from left to right, in a section for the pieces the work defines, a section for the reference material the work brings, and one section per flow, a flow being the screens a user goes through to complete one task in one app; a grouping the user asked for SHALL replace that list. It SHALL create a section when its first node is placed and never an empty one, SHALL name each section with a two-digit number and what it holds, SHALL tell the user which sections it expects to create before the first drawing without waiting for an answer, SHALL regroup what is drawn when a later reply changes the grouping, and SHALL record each section's node with the example. It SHALL NOT move or rename a node the work did not create, nor move one into a section, and SHALL ask as a decision before grouping existing frames a request names as its target; unattended work SHALL NOT ask it and SHALL report those frames as unresolved.

#### Scenario: A work with two flows and reference captures

- **GIVEN** a work that draws the screens of two tasks of one app, defines a component and brings captures of another product
- **WHEN** the stage draws its frames
- **THEN** the page holds a section for the work's pieces, one for the captures and one per flow, each numbered and named, and no node of the work sits outside them

#### Scenario: The user asked for another grouping

- **GIVEN** a request that asks for one section per screen
- **WHEN** the stage draws its frames
- **THEN** it groups them as asked, and the other section rules still apply

#### Scenario: Rules added to frames that already existed

- **GIVEN** an attended request to add rule cards to existing frames that sit loose on the page
- **WHEN** the stage draws the cards
- **THEN** each card stays beside its frame, no existing frame is moved, no empty section is created, and the approval round asks whether to group those frames

### Requirement: Entry router with playbooks
Hive SHALL provide an entry skill, `hive-mode`, that replaces `flow-research`. It SHALL read before asking, play the request back, ask what is missing and settle what the user needs, and SHALL NOT edit repository files itself; it MAY hand the work to `flow-build` in the same turn only for a small understood change that the user's own request named with its target, never after a reply to its question or proposal. It SHALL classify the work and open at most one playbook under `playbooks/<kind>.md`, record its name in the change folder's `proposal.md` (written by `flow-plan`, or by `flow-build` for a change without a plan that alters a behavior an existing spec describes) or, for work without a change folder, in a `- Playbook:` line of the checkpoint file, and, when no playbook fits or the work is large and cross-cutting, follow the general route from research to plan or build. A playbook names the order of the stages, what is specific to that kind and the leaves it uses; the stages (`create-design`, `flow-plan`, `flow-build`, `flow-verify`, `flow-close`, `flow-deploy`) SHALL remain commands that read the playbook's specific steps.

#### Scenario: A reported defect with a fix request
- **WHEN** the user reports a defect and asks for its fix
- **THEN** `hive-mode` opens the `bugfix` playbook, and build and verify fix the defect at its cause as «Defects are fixed at their cause» sets, with its regression evidence.

#### Scenario: An investigation confirms a defect
- **GIVEN** work running the `investigate` playbook
- **WHEN** the findings confirm a defect and the user asks for its fix
- **THEN** the work hands off explicitly to the `bugfix` playbook as a scope change, and the new playbook is recorded

## ADDED Requirements

### Requirement: Review of a design that already exists

`create-design` SHALL have an entry that reviews a design someone already drew, given as a link to Figma frames the connection can read or as an image, without drawing, moving or changing anything in the design file. One `hive-review-ux` child SHALL review it in a thread of its own, without the reasoning of whoever drew it, against the request it serves when one is given, the app's existing screens and recorded decisions, and the UI criteria of `create-design`. The entry SHALL return the findings by priority, the states the design lacks, such as empty, loading, error and permission states, each business rule the design shows that the request, the ticket and the recorded decisions leave unsettled, asked as a question with a recommended answer, and a verdict. The findings SHALL stay in the conversation unless the user asks for a saved file, and the entry SHALL end with a block of replies and no command. Reviewing an image SHALL NOT need a connection to Figma, and the entry router SHALL send it to this entry without the connection check; without a connection that reads Figma, a link SHALL be reported as unreadable and an image SHALL still be reviewed, with the measurements that need the file's nodes reported as not run.

#### Scenario: A Figma link to frames another person drew

- **GIVEN** a host whose connection reads Figma and a link to frames another person drew
- **WHEN** the user asks to review that design
- **THEN** one `hive-review-ux` child reviews them in a thread of its own and the entry returns findings by priority, missing states, unsettled rules as questions with a recommended answer and a verdict
- **AND** no node of the file is created, moved or changed

#### Scenario: An image of a mock with no Figma connection

- **GIVEN** a host with no connection that reads Figma and an image of a mock
- **WHEN** the user asks to review it
- **THEN** the entry reviews the image and reports the node measurements as not run
