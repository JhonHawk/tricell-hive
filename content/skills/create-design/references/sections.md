# Canvas sections

Read before the first drawing, at [create-design](../SKILL.md) step 5, and again before the [mobile phase](approval.md#mobile-phase) draws its first frame; before placing a rule card, a reference image or a piece this work defines on the canvas; and when a step removes frames this work drew or the user asks for another grouping. A section is the design tool's own container for organizing a page, with a name shown on the canvas; it is not a frame and not a group.

## Nothing stays loose

Every node this work creates on a page sits inside a section this work created: each frame, rule card, label, reference image and component it defines. The one exception is a rule card for a frame this work did not create, as «What this work did not create» sets. Before each review of step 5 and of the mobile phase, read the page's top-level nodes and move any this work created, other than its sections and those cards, inside its section.

## Which sections

Group what the work creates in these sections, in this order along the line [Where sections go](#where-sections-go) sets:

1. **Assets:** the components and other pieces this work defines in the file, the ones its record lists as `candidate:`.
2. **Moodboard:** the reference material the work brings, such as captures of other products or of the screen it replaces.
3. **One section per app:** the screens the work draws for one app, with one band per flow, as [Inside a section](#inside-a-section) sets. A flow is the screens a user goes through to complete one task. Screens whose app shell differs, such as a mobile app and a web portal, belong to different apps, each on its app's page as [draw](draw.md#what-the-stage-may-change-in-the-file) sets.
4. **The mobile twin:** for a web app served on mobile, its 440 px frames, which the [mobile phase](approval.md#mobile-phase) draws after the web frames are approved, in a section of their own beside the web section. An app that runs only on mobile has its one section, drawn at 440 px, and no twin.

Create a section when its first node is placed, never an empty one. A section created after the others, such as Assets when its first piece is defined after the review, goes at the end of the line and takes the next number; move or renumber no other for it. A mobile twin is the exception: it takes its web section's number and goes beside it, as «Where sections go» sets.

A grouping the user asked for, in the request or this conversation, replaces this list (the other rules here still apply; a node it gives no place, such as a piece this work defines, goes in its section of this list). Before the first drawing, say in one line which sections you expect to create, and draw without waiting. When a later reply changes the grouping, regroup what is drawn: move this work's nodes, rename and renumber its sections, rewrite their record lines. When the stage runs again for a work whose record lists sections, reuse them and create only a section the record lacks.

## Names

Name each section `NN · <what it holds>`: a two-digit number in creation order, a middle dot between spaces, and what it holds in the session language. An app's section also carries the work's ticket key after the number when the work has a ticket, such as `03 · ABC-123 · Check-in`. When the page already has numbered sections, continue their numbering. A mobile twin takes its web section's number and name followed by ` (mobile)`, such as `03 · ABC-123 · Check-in (mobile)`.

Label each band `N · <flow>`: a number from 1 in the order of the bands and the flow's name in the session language, such as `2 · Requests`. Name the label's text node `Row label · N · <flow>`, with `Row label` written in the session language. A mobile twin repeats its web section's labels with their numbers and leaves out the band of a flow it has no frame for, without renumbering the others, so a band keeps its number on both sides.

## Inside a section

- **Bands.** Place one band per flow, top to bottom in the order the user meets the flows, with 200 px between the lowest node of a band and the top of the next.
- **Label.** Place each band's label 80 px from the section's left edge, 560 px wide, its top aligned with the band's top; the band's frames start 120 px to its right. Every label of the section uses the same text style.
- **Order.** Place a band's screens in one row, tops aligned, in the order the user meets them on the flow's main path. After them, in the same row, place the states and variants of those screens, such as an error, an empty list or a newer version, introduced by a text node reading «States and variants» in the session language, 40 px above the top of the first of them and aligned to its left edge. A band without states has no such text.
- **Alternatives.** Place below a screen, aligned to its left edge, each alternative example of it that the user asked for.
- **Step with cards.** In a section whose frames carry rule cards, a web section or the section of an app that runs only on mobile, each frame has a card's slot: the 400 px card 40 px to its right, as [rule cards](rule-cards.md) places it, then 200 px to the next frame. The step from one frame's left edge to the next is the frame's width plus 640 px: 2080 px for a 1440 px frame and 1080 px for a 440 px one. A frame without a card keeps its slot empty, so every frame of the section keeps the same step.
- **Mobile step.** In a mobile twin, frames carry no card, since each shares its web screen's card as [rule cards](rule-cards.md) sets, and the step is 540 px: 440 px frames 100 px apart. A band holds the mobile frames of the screens and states of its web band, in the same order, with the «States and variants» text before the first state.
- **Section bounds.** Resize the section to fit its content with 80 px of padding on every side, after each node added, moved, redrawn or removed.

## Where sections go

Place this work's sections in one line, 200 px apart, in the direction the page's numbered sections already run, leaving mobile twins out of that line. Decide the direction once, when this work places its first section on the page, and keep it for every later section of this work, a rerun of the stage included:

- **A column,** when the two highest-numbered sections on the page that this work did not create, mobile twins left out, share their left edge: the first of this work's sections goes 200 px below the page's lowest node, at that left edge, and each next one below it, left edges aligned.
- **A row,** otherwise, tops aligned: on an empty page the first starts at x 0, y 0; on any other page the row starts 200 px right of the page's existing content, at that content's top.

Place a mobile twin 200 px right of its web section, tops aligned. The two count as one unit of the line: in a column, the next section goes 200 px below the lower of the two; in a row, it starts 200 px right of the twin.

After any node is added, moved or resized, in this work's sections or beside them (such as a card beside an existing frame), move this work's sections along their line, down in a column and right in a row, until each is 200 px from the section or unit before it and at least 200 px from every other node, and keep each twin 200 px right of its web section. A section never overlaps another node.

## What this work did not create

Never move into a section, move or rename a node this work did not create. The frames a request names as its target change only as [A request for the design itself](special-cases.md#a-request-for-the-design-itself) sets and stay where they are; what this work creates beside them goes in its own sections, and a rule card for one of them stays beside it, outside this work's sections. When those frames sit loose on the page, ask once, as a decision in step 6's round, whether to group them too; under [unattended delegation](../../unattended-delegation/SKILL.md), leave them as they are and name them among what you report unresolved.

## Record and removal

Add one `canvas section: <name> · <node>` line per section, the mobile twin included, to the example's record, next to the frames' nodes. When a step removes frames (the alternatives not chosen at step 7, the frames of a screen stopped by a `scope` rule), resize their section; delete a section this work created only when it is left empty, and its record line with it. When the connection errors creating a section or has none, draw the frames in the same bands and order without one and say so in step 6's message.
