# Review a design that already exists

Read by [create-design](../SKILL.md) when the request is to review a design someone already drew, given as a link to Figma frames or as an image. It draws, moves and changes nothing in the design file and writes nothing in the repository. Steps 1 to 8 of the skill do not apply; the connection check of [design stage](../../hive-core/ui/design-stage.md) does not apply to an image.

## What it accepts

- **A Figma link** the connection can read: reading is enough, no write access is needed. Without a connection that reads Figma, report the link as unreadable and ask for an image of the frames.
- **An image** of the design: it needs no connection. The review judges from the image and reports each measurement that needs the file's nodes (alignment, shell size, nodes outside the automatic layout, instance use) as not run.

## How it runs

1. Gather what the review judges against: the request the design serves, when one is given; the app's existing screens, in the file and in the code; the decisions recorded for the app in its `DESIGN.md`; and the criteria of [UI criteria](ui-criteria.md) with its [precedence](ui-criteria.md#precedence). A source that does not exist is named as absent, not guessed.
2. Start one `hive-review-ux` child in a thread of its own, briefed as [reviewer brief](../../hive-core/delegation/reviewer-brief.md) sets, giving the file and nodes or the image, the request when there is one, the app's existing screens and recorded decisions, and no sibling list, library inventory, width or how the design was reached. The child never receives the reasoning of whoever drew the design. The main thread draws nothing and does not review in place of the child.
3. Return the child's result to the user as these four parts, in this order:
   - the findings, by priority, each with the frame or image region it concerns;
   - the states the design lacks, such as empty, loading, error and permission states;
   - each business rule the design shows that the request, the ticket and the recorded decisions leave unsettled, asked as a question with a recommended answer, as [questions](../../hive-core/messages/questions.md) sets;
   - a verdict, with each check the review could not run named as not run.

The findings stay in the conversation; save them to a file only when the user asks, in the destination the project's layout sets.

## How it ends

End with a block of replies and no command, as [report endings](../../hive-core/messages/report-endings.md) sets for a review.

## Judging frames

The procedure of the `hive-review-ux` child for frames in a design file or an image of a design; it replaces steps 1 to 3 of that role. Read each frame and each sibling screen the brief names through the design tool's connection, the image and the node tree of both; a sibling whose image was not read is a comparison not run, reported as such. Judge each frame, components and copy included, against the request, what the siblings fix (as the brief lists it, or measured in the named siblings), and the decisions recorded for the app, by [UI criteria](ui-criteria.md) and its precedence.

- **Library:** when the brief names a design library, declared in `DESIGN.md` or not, a piece or variable it holds that the frames draw other than as its instance or binding is a finding, unless the brief says the file could not use the library. Search the library through the connection for each `candidate:` piece in the brief and each piece drawn without an instance, and report the ones it holds. A piece the brief's inventory shows the siblings instance, drawn other than as an instance of that component or of the library's own, is a finding with or without a library.
- **Siblings' images:** compare for what a node tree does not show: the fill of table headers and containers, the page's side margins, the container around filters and toolbars, the density of rows.
- **Alignment:** measure edges and spacing shared with the siblings, a shell whose size matches theirs, nodes outside automatic layout or absolutely positioned where the siblings use neither, and text that does not fit its container.
- **Unsettled rules:** list apart each business rule a frame shows, as [business rules](../../hive-core/business/business-rules.md) defines one, that neither the brief's requirements and recorded decisions nor the siblings settle; it does not fail the verdict.

Driving a browser, loading, empty and error states, keyboard order, focus, the automated audit, the defects the base already shows and the handoff captures do not apply. Without a connection that reads the file, judge from the images in the brief and report each alignment measurement as not run.

The verdict starts as not passed and is passed when the frames show no finding against the request, the siblings, the recorded decisions or the brief's library. In a first review it stays not passed while a sibling image or a library search the connection lets you run was not run; one the connection cannot run is reported as not run without failing the verdict. A follow-up after a fix rereads only the corrected frames. When its brief gives rule cards, read each card with its frame and the brief's rule lines, as [rule cards](rule-cards.md) sets, and report as findings a row that repeats the frame's visible text, a row the frame contradicts, and a rule of those lines that its screen's card lacks although the frame's text does not say it. Report as step 4 of the role sets, with the unsettled rules after the findings, without the base's defect list and without states not covered. Never change the frames.

Given an image of a design and no file or nodes, judge from the image alone: every measurement that needs nodes is not run, and the verdict is passed when the image shows no finding against the request, the recorded decisions and the criteria.
