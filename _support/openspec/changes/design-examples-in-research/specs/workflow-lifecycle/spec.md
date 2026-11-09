## ADDED Requirements

### Requirement: Screen examples in research

When the user asked for a change and what gets built includes a new screen, or one that adds, removes or moves a region, and neither the request nor an existing screen of the project fixes its composition, research SHALL ask the user once per work, as a decision, whether to design and present proposals for the screen; a change that keeps the screen's layout SHALL NOT get that offer, and neither SHALL work that runs under unattended delegation. The offer SHALL give text sketches as an option always, and frames in the project's design file as an option only when the app's design record declares that file or the user named it in the conversation, and the host has a connection that can write to it. The design-file option SHALL name the write to that file in its label or description, and the user's choice of it SHALL be the instruction for those writes; the offer SHALL NOT replace the question a critical effect takes. Research SHALL show two or three examples when the composition is open and one when it is fixed. In a design file the work SHALL create frames on a page named after the work and SHALL change or delete only the nodes it created. The user SHALL choose or correct an example as a decision; the approved example SHALL be recorded in the report and in a note of the work's support folder, for a frame with its location, its nodes and a local image and for a text sketch with the sketch itself, and the examples not chosen SHALL be removed. A request whose deliverable is the design SHALL be drawn the same way without the offer, SHALL NOT be reported as done before the user approves the example, and after that approval research SHALL end with the route to the close, which treats it as an operational task; under unattended delegation that approval SHALL stay unanswered and be reported as unresolved. A plan for the screen SHALL use the approved example as the after sketch of the regions it shows, and the build's first view, the verification walk and the UX review SHALL compare the rendered screen against it. A difference SHALL be a region, a component, an action's placement or an order that differs from the example, not rendering variance, and not what the plan records as decided otherwise; it SHALL be handled as a UI defect the change introduced, and the differences that remain SHALL be listed in the validation message when the delivery stops for validation and in the final report when it does not. The user's validation in verification SHALL be asked as before.

#### Scenario: Declared design file and a writing connection

- **GIVEN** an app whose design record declares a design file, and a host whose connection can write to it
- **WHEN** research reaches a new screen whose composition is open
- **THEN** it asks whether to design proposals, with the design file, text sketches and no proposals as options
- **AND** on the design-file answer it shows two or three examples as frames on a page of that file named after the work

#### Scenario: No design tool

- **GIVEN** an app with no declared design file, or a host with no connection that can write to it
- **WHEN** research reaches a new screen whose composition is open
- **THEN** the offer gives text sketches and no proposals as its options

#### Scenario: A change that keeps the layout

- **GIVEN** a request to change a button, a copy, a field or a column of an existing screen
- **WHEN** research settles what the work is
- **THEN** it makes no offer of proposals

#### Scenario: The request is the design

- **GIVEN** a request to draw or replicate a screen in a design file
- **WHEN** research draws the example
- **THEN** it makes no offer and asks the user to approve or correct the example before reporting the work as done
- **AND** after the approval it ends with the route to the close

#### Scenario: The built screen differs from the approved example

- **GIVEN** a plan whose `UI` section records an approved example
- **WHEN** the verification walk captures a region the example shows
- **THEN** each difference in a region, a component, an action's placement or an order is handled as a UI defect the change introduced
- **AND** the differences that remain are listed in the validation message, or in the final report when the delivery does not stop for validation

#### Scenario: Unattended delegation is active

- **GIVEN** work that runs under unattended delegation
- **WHEN** research reaches a new screen whose composition is open
- **THEN** it makes no offer of proposals and the plan settles the UI
