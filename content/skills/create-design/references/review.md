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
