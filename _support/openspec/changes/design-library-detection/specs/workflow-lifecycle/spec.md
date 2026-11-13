## MODIFIED Requirements

### Requirement: Screen proposals in Figma before the plan

Screen examples SHALL be drawn by a stage of their own, `flow-figma-design`, that runs between research and the plan and that only the user's reply starts, except the continuation from research for a request whose deliverable is the design. When the host has a connection that can write to Figma, research SHALL end with the next-step block that recommends that stage, with going straight to the plan, or to the build for a small understood change, as other replies, when what gets built includes a new screen, even when sibling screens fix its composition and the change would otherwise continue into the build, and when it changes an existing screen by adding, removing or moving a region and neither the request nor an existing screen fixes that composition; this SHALL hold when research assessed a ticket the user picked as well as when the user asked for a change. Research SHALL NOT mention the stage for a change that keeps the screen's layout, on a host with no connection that can write to Figma, or under unattended delegation, and SHALL make no offer of proposals through a question. The user's reply that starts the stage SHALL be the instruction for the writes it makes to that Figma file, limited to what the work creates, and SHALL NOT replace the question a critical effect takes. A later flow's correction of a rule card's row, as «Rule cards beside each frame» sets, SHALL need no such reply: the guidance's standing permission covers it, limited to the card nodes the work created. On a host with no connection that can write to Figma, the stage SHALL draw nothing and report what is missing. The stage SHALL load the Figma connection's own usage instructions, when it ships them, the project's design system, the sibling screens of the same module or navigation in the file and in the code, the decisions already recorded for the app in its design record, and the record of the example each sibling was approved from when a search for the file's location and its frame node in the project's change folders, archived ones included, and in the example records under the project's support workspace finds one, before it draws. It SHALL draw in the Figma file the user named in the conversation or the app's design record declares; a declared file of another design tool SHALL NOT count, and a declared file that neither the block that offered the stage nor the user named SHALL be confirmed by the user once before the first write. When the connection cannot write to that file, the stage SHALL draw nothing and report it. With neither file, it SHALL ask the user for the file once, in the text, and end with the route to the plan when the user gives none. Before the first drawing it SHALL tell the user each step they will have to take by hand in Figma. It SHALL draw one example per screen, following its sibling screens, and SHALL draw two or three examples only for a screen the user names when asking for alternatives, as frames on a page named after the work, and SHALL change or delete only the nodes the work created; it MAY create components and variables local to the file, SHALL record their identifiers with the example, and SHALL remove those only the unchosen examples used; a change to an existing variable, style or component, or to a shared library, SHALL return to the user as a decision. The main thread SHALL draw the frames; no child SHALL draw them. When the connection runs scripts in the file, each write SHALL define once the pieces its frames use, built on the tool's automatic layout and bound to the file's variables, with the app shell cloned from an existing screen, and SHALL compose the frames from them; the stage SHALL bind one Figma design library before it draws: the one the user named or shared in the conversation, otherwise the one the app's design record declares, otherwise one it detects, which is a library the file has enabled as the connection lists it, a library whose components the sibling frames instance or whose variables they bind, or a library the example record of a sibling names; a link the user gives to a Figma file other than the one the stage draws in SHALL count as the library the user named when the user presents it as a library, a kit, a design system or the source of the components, or when the connection lists it among the libraries that file has enabled, and SHALL NOT count when the user links it as the screen to replicate or as a reference; among several detected libraries the stage SHALL bind the one the sibling frames instance, and SHALL ask the user once, in the text, only when that leaves none or more than one, or when the user named more than one. Before the first drawing the stage SHALL write in the example's record an inventory, keeping every line of a record that already exists: each component and variable the sibling frames instance or bind, from any library or from the file itself, and, for each piece the work's screens need that no sibling uses, the terms it searched in the bound library and what the search found. Each piece the siblings use SHALL come from that same component or variable, whether or not a library is bound, each piece and variable the bound library holds SHALL come from it, as an instance of its component or a binding to its variable, a piece or variable SHALL be defined only for what the inventory shows the siblings and the library lack, each such piece SHALL be listed in the example's record as a candidate for the library and named in the approval message, adding the library to the file SHALL be among the steps the user takes by hand, and when the file cannot use the library the stage SHALL draw with its own pieces and say so in the approval message; a correction SHALL change a piece or the data and draw the frame again, not move or resize nodes one by one. Examples SHALL be drawn at one desktop width, the one the screen's existing frames in the file use, otherwise 1440 px. When the work has more than one screen, the stage SHALL first draw, at that desktop width only, the screen that uses the most pieces of the inventory, have one `hive-review-ux` child review it alone, fix the findings that hold, record the unsettled rules it lists and give the user its location without asking anything and without ending the turn, and only then draw the other screens and every 440 px frame from its pieces; that review SHALL get no rule cards and no follow-up, and the first review the rule cards wait for SHALL be the one of the whole set. Before the user is asked, one `hive-review-ux` child SHALL review the drawn frames, the first screen's included, and one the 440 px frame, against the request, the sibling screens, the recorded decisions, the bound design library and their alignment, without the drawer's reasoning; the reviewer SHALL read each sibling's image as well as its node tree, SHALL report as a finding a piece the inventory shows the siblings instance that is drawn other than as an instance, SHALL search the bound library for each candidate piece and report as a finding one the library holds, SHALL NOT pass the frames while a sibling's image or a library search that the connection lets it run was not run, and SHALL report one the connection cannot run as not run, which the stage SHALL name in the approval message; without a connection that reads the file, it SHALL judge from the frames' images and report the alignment measurements as not run, and the stage SHALL say so in the approval message. The main thread SHALL fix the findings that hold, have only the corrected frames and the rule cards reviewed again once, update the record's status to pending the user's approval, and name what remains in the message to the user. A correction the user asks for SHALL NOT get another review. The stage SHALL finish drawing and reviewing the completed set, including each applicable 440 px mobile frame, before asking one decision to approve the set or request corrections naming affected screens, alongside the separate proposed rule decisions that «Business rules that appear while drawing» sets. It SHALL name the recommended example for each screen with alternatives, and the user may name a different alternative with corrections. Corrections SHALL leave only the named screens pending and approve the others; after redrawing, one approval decision SHALL cover only the affected screens and their mobile frames, preserving the others' approval. When the user's correction sets a convention that applies beyond that screen, the stage SHALL ask once, in the round that asks the approval, whether to record it among the decisions of the app's design record, and SHALL keep a yes in the example's record without editing the design record itself. The approved example SHALL be recorded in the report and in a note of the work's support folder with the file's location, each frame's node, a local image and its status; that note SHALL be created with the inventory before the first drawing, and its frame and status lines SHALL be written as soon as the frames are drawn, with a status saying they are under review, before the review and before the user is asked, and its status SHALL be updated after the review and when the user approves. The record SHALL hold one status per screen, and a plan SHALL treat only a screen whose status is approved as having an approved example; the alternatives not chosen SHALL be removed, and the approved examples of the other screens SHALL stay. When the app's design record does not declare the Figma file the stage drew in or the library it bound, or declares another Figma file or library, the stage SHALL ask once, in the round that asks the approval, whether to record them there, showing each value as it will be written beside the value it replaces; a field that names a file or a library of another design tool SHALL stay as it is and SHALL NOT be asked about. On a yes it SHALL write the design file and the design library fields of that record in the same turn, adding a missing field, replacing only a value the question showed as replaced, changing nothing else in it and showing the change; when the record exists, the stage SHALL ask the delivery question in that same round when no delivery answer is in force for that repository, and SHALL deliver that edit at once, as the guidance sets for a mechanical change, apart from any plan or build that follows, and that delivery SHALL change neither how the design work closes nor the state of its ticket. When the app has no design record, a yes SHALL write nothing in the repository and SHALL be kept in the example's record and named as pending; a no SHALL be kept in the example's record so the work does not ask again. For a request whose deliverable is the design, research SHALL continue into the stage in the same turn when the host has a connection that can write to Figma, and SHALL report the missing connection as a blocker when it has none. Once every screen is approved, the stage SHALL end with the next-step block toward the plan, or toward the build only when the block that offered the stage listed the build, the approved examples add no region the request did not name, the record holds no yes to recording a convention and no pending yes to recording the file or the library, and no rule in it is still proposed, except when a rule changed the scope, which ends as «Business rules that appear while drawing» sets, and except for a request whose deliverable is the design, which ends as this requirement sets for it. A request whose deliverable is the design SHALL be drawn by the same stage, in the page or frames it names when it names them, changing and never deleting those frames; it SHALL NOT be reported as done before the user approves the example, and after that approval the stage SHALL name as pending each yes to recording a convention in the design record, and each yes to recording the file or the library that it could not write, since no plan follows to write it, and SHALL run the close, which treats it as an operational task. Under unattended delegation the stage SHALL ask neither for the file nor for the confirmation of a declared one, SHALL bind a library without asking and draw with its own pieces when detection leaves none or more than one, SHALL neither ask to record the file or the library nor write the design record, SHALL draw nothing and report the file as unresolved when it has none it may use, the choice or the approval of an example SHALL stay unanswered and be reported as unresolved, every example SHALL be kept, and the plan SHALL settle the UI without an approved example. A plan with a change folder SHALL carry the approved example's record in its `UI` section and that record's rule lines in the rule record of its proposal, and work without one SHALL carry both in its checkpoint file. A plan for the screen SHALL use the approved example as the after sketch of the regions it shows, the plan SHALL add the task that writes each convention the record holds a yes for, and the design file and the design library when the record holds a pending yes for them, a build without a plan SHALL keep the record in its checkpoint file, and the build's first view, the verification walk and the UX review SHALL compare the rendered screen against it. A difference SHALL be a region, a component, an action's placement or an order that differs from the example, not rendering variance, and not what the plan records as decided otherwise or a correction the user asked for after approving the example, which the example's record SHALL name; it SHALL be handled as a UI defect the change introduced, and the differences that remain SHALL be listed in the validation message when the delivery stops for validation and in the final report when it does not. When a frame's image is missing where the comparison runs, the frame SHALL be read again from its recorded location and node, and a comparison that cannot run SHALL be reported as not run. The user's validation in verification SHALL be asked as before.

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
- **THEN** the stage writes the design file and the design library fields in the design record in that turn and shows the change
- **AND** that edit is delivered at once, under the delivery answer in force or the one asked in that same round, before the close

#### Scenario: A yes to recording with no design record

- **GIVEN** an app with no design record
- **WHEN** the user answers yes to recording the file and the library
- **THEN** the stage writes nothing, keeps the yes in the example's record and names it as pending
