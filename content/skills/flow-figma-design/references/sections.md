# Canvas sections

Read before the first drawing, at [flow-figma-design](../SKILL.md) step 5; again before placing a rule card, a reference image or a piece this work defines on the canvas; and when a step removes frames this work drew or the user asks for another grouping. A section is the design tool's own container for organizing a page, with a name shown on the canvas; it is not a frame and not a group.

## Nothing stays loose

Every node this work creates on a page sits inside a section this work created: each frame, each rule card, each reference image and each component it defines. The one exception is a rule card for a frame this work did not create, as «What this work did not create» sets. Before each review of step 5, read the page's top-level nodes and check that the only ones this work created are its sections and those cards; move any other inside its section.

## Which sections

Group what the work creates in these sections, from left to right in this order:

1. **Assets:** the components and other pieces this work defines in the file, the ones its record lists as `candidate:`.
2. **Moodboard:** the reference material the work brings, such as captures of other products or of the screen it replaces.
3. **One section per flow:** a flow is the screens a user goes through to complete one task, in one app. Screens whose app shell differs, such as a mobile app and a web portal, are never the same flow. A work with one screen has one flow section.

Create a section when its first node is placed, never an empty one. A section created after the others, such as Assets when its first piece is defined only after the review, goes at the end of the row and takes the next number: do not move or renumber the others for it.

When the user asked for a grouping, in the request or in this conversation, that grouping replaces this list, and the other rules of this reference still apply; a node that grouping gives no place, such as a piece this work defines, goes in its section of this list. Before the first drawing, tell the user in one line which sections you expect to create, by name, and draw without waiting for an answer. When a later reply changes the grouping, regroup what is drawn: move this work's nodes, rename and renumber its sections, and rewrite their record lines.

When the stage runs again for a work whose record lists sections, reuse them: read their nodes and create only a section the record lacks.

## Names

Name each section `NN · <what it holds>`, such as `03 · Check-in flow`: a two-digit number in the order the sections are created, a middle dot between spaces, and what it holds in the session language. When the page already has numbered sections, continue their numbering instead of starting at 01.

## Inside a section

- Place a flow's screens in one row, in the order the user meets them, with their tops aligned and the same gap between them. The gap holds a rule card with 40 px on each side, which is what [rule cards](rule-cards.md) needs to place one: 480 px while no card exists, and the widest card plus 80 px once the cards are drawn, moving this work's frames to the right when that is wider.
- Place below a screen, aligned to its left edge, each other frame of that same screen: a state, an alternative example, and its 440 px frame.
- A rule card sits where [rule cards](rule-cards.md) places it beside its frame; that place is inside the frame's section.
- Resize the section to fit its content with 80 px of padding on every side, again after each node added, moved, redrawn or removed.

## Where sections go

Place this work's sections in one row with their tops aligned, 200 px apart. On a page with no other content, the first one starts at x 0, y 0; otherwise the row starts 200 px to the right of the page's existing content, at that content's top. After any node is added, moved or resized, in this work's sections or beside them, such as a card beside an existing frame, move this work's sections to the right until each is again 200 px from the section before it and at least 200 px from every other node. A section never overlaps another node.

## What this work did not create

Never move into a section, move or rename a node this work did not create. The frames a request names as its target are changed only as [A request for the design itself](../SKILL.md#a-request-for-the-design-itself) sets, and stay where they are; what this work creates beside them goes in its own sections. A rule card for one of those frames stays beside it, where [rule cards](rule-cards.md) places it, outside this work's sections. When those frames sit loose on the page, ask once, as a decision in step 6's round, whether to group them too; while [unattended delegation](../../unattended-delegation/SKILL.md) is active, do not ask it: leave them as they are and name them among what you report unresolved.

## Record and removal

Add one `canvas section: <name> · <node>` line per section to the example's record, next to the frames' nodes. When a step removes frames, such as the alternatives not chosen at step 7 or the frames of a screen stopped by a `scope` rule, resize their section; delete a section this work created only when it is left empty, and its record line with it.

## Without sections

When the connection returns an error creating a section, or the tool has none, draw the frames in the same rows and order without one and say so in step 6's message.
