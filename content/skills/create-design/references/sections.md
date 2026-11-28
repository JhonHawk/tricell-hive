# Canvas sections

Read before the first drawing, at [create-design](../SKILL.md) step 5; again before placing a rule card, a reference image or a piece this work defines on the canvas; and when a step removes frames this work drew or the user asks for another grouping. A section is the design tool's own container for organizing a page, with a name shown on the canvas; it is not a frame and not a group.

## Nothing stays loose

Every node this work creates on a page sits inside a section this work created: each frame, rule card, reference image and component it defines. The one exception is a rule card for a frame this work did not create, as «What this work did not create» sets. Before each review of step 5, read the page's top-level nodes and move any this work created, other than its sections and those cards, inside its section.

## Which sections

Group what the work creates in these sections, in this order along the line [Where sections go](#where-sections-go) sets:

1. **Assets:** the components and other pieces this work defines in the file, the ones its record lists as `candidate:`.
2. **Moodboard:** the reference material the work brings, such as captures of other products or of the screen it replaces.
3. **One section per flow:** the screens a user goes through to complete one task, in one app; screens whose app shell differs, such as a mobile app and a web portal, are never the same flow. A work with one screen has one flow section.

Create a section when its first node is placed, never an empty one. A section created after the others, such as Assets when its first piece is defined after the review, goes at the end of the line and takes the next number; move or renumber no other for it.

A grouping the user asked for, in the request or this conversation, replaces this list (the other rules here still apply; a node it gives no place, such as a piece this work defines, goes in its section of this list). Before the first drawing, say in one line which sections you expect to create, and draw without waiting. When a later reply changes the grouping, regroup what is drawn: move this work's nodes, rename and renumber its sections, rewrite their record lines. When the stage runs again for a work whose record lists sections, reuse them and create only a section the record lacks.

## Names

Name each section `NN · <what it holds>`, such as `03 · Check-in flow`: a two-digit number in creation order, a middle dot between spaces, and what it holds in the session language. When the page already has numbered sections, continue their numbering.

## Inside a section

- Place a flow's screens in one row, in the order the user meets them, tops aligned, with the same gap between them. The gap holds a rule card with 40 px on each side: 480 px while no card exists, and the widest card plus 80 px once the cards are drawn, moving this work's frames right when that is wider.
- Place below a screen, aligned to its left edge, each other frame of that screen: a state, an alternative example and, for a web screen served on mobile, its 440 px frame. A screen of a mobile-only app is drawn once, at 440 px.
- A rule card sits where [rule cards](rule-cards.md) places it beside its frame, inside the frame's section.
- Resize the section to fit its content with 80 px of padding on every side, after each node added, moved, redrawn or removed.

## Where sections go

Place this work's sections in one line, 200 px apart, in the direction the page's numbered sections already run. Decide the direction once, when this work places its first section on the page, and keep it for every later section of this work, a rerun of the stage included:

- **A column,** when the two highest-numbered sections on the page that this work did not create share their left edge: the first of this work's sections goes 200 px below the page's lowest node, at that left edge, and each next one below it, left edges aligned.
- **A row,** otherwise, tops aligned: on an empty page the first starts at x 0, y 0; on any other page the row starts 200 px right of the page's existing content, at that content's top.

After any node is added, moved or resized, in this work's sections or beside them (such as a card beside an existing frame), move this work's sections along their line, down in a column and right in a row, until each is 200 px from the section before it and at least 200 px from every other node. A section never overlaps another node.

## What this work did not create

Never move into a section, move or rename a node this work did not create. The frames a request names as its target change only as [A request for the design itself](special-cases.md#a-request-for-the-design-itself) sets and stay where they are; what this work creates beside them goes in its own sections, and a rule card for one of them stays beside it, outside this work's sections. When those frames sit loose on the page, ask once, as a decision in step 6's round, whether to group them too; under [unattended delegation](../../unattended-delegation/SKILL.md), leave them as they are and name them among what you report unresolved.

## Record and removal

Add one `canvas section: <name> · <node>` line per section to the example's record, next to the frames' nodes. When a step removes frames (the alternatives not chosen at step 7, the frames of a screen stopped by a `scope` rule), resize their section; delete a section this work created only when it is left empty, and its record line with it. When the connection errors creating a section or has none, draw the frames in the same rows and order without one and say so in step 6's message.
