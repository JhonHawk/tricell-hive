# Rule cards

Read at step 5 of [flow-figma-design](../SKILL.md), once the first review is reconciled, to draw the rule cards; at its step 7, when the user answers a rule; and from any later flow when a rule a card shows changes or a decision contradicts an approved frame. [Business rules](../../flow-plan/references/business-rules.md) defines a business rule, the line that records it and its statuses.

A rule card is an annotation beside a frame that lists the business rules shaping that screen which the frame does not say by itself, so a reader of the design file finds them without the ticket or the conversation. It is not part of the interface.

## Which rules go on a card

For each frame the work drew, take:

- each `rule:` line of the example's record whose screen is that frame's, with status `proposed`, `assumption`, `approved` or `stated`;
- each rule a source states, in the request, the ticket or the documentation the work read, that shapes that screen and that the record does not already hold.

Keep a rule only when its condition or its consequence is not written in the frame's visible text: leave out a rule that a label, a helper text or a message of the frame already says in full. Add each stated rule you keep to the record only when that screen has no line for it, as `rule: <the rule> · screen <name> · stated: <source> · found by drawer`.

A frame with no rule to keep gets no card. A `scope` rule has stopped its screens, so it has no frame and no card. The 440 px frame of step 6 shares its screen's card: draw no second one. A screen drawn with alternatives gets one card per alternative.

## Draw the cards

Draw them after the first review's findings are fixed and its unsettled rules recorded, never before: the first reviewer must list the unsettled rules without seeing yours.

- Use the declared design library's annotation component when it holds one; otherwise define one card component in the file, on the tool's automatic layout and bound to the file's variables, as a piece of this work with its `candidate:` line, and place one instance per frame.
- Place each card outside its frame's bounds, to the right, aligned to the frame's top and 40 px away, so an export of the frame leaves it out. Space the frames this work draws so a card fits between two of them. Never move a frame this work did not create to make room: ask it as a decision first.
- Open the card with a heading that says it is a design annotation and not interface, in the session language.
- Write one row per rule: its origin, then the rule in one or two sentences of plain words, never a decision ID alone. The origin is one of four, named in the session language: the specification, for a `stated` rule, with the source's own identifier when it has one; a decision, for an `approved` rule, with its date; an assumption; or still to decide, for a `proposed` rule.
- Put on a card no row that repeats the frame's text and no rule of another screen.
- Add one `card: <screen> · <node>` line per card to the example's record, where those lines stay.

Then the one follow-up review of step 5 covers the cards together with the corrected frames, and runs for the cards alone when no frame was corrected; its brief gives each card's node, as [Reviewer brief](../../flow-build/references/delegation.md#reviewer-brief) sets. Fix what holds. The cards get no further review.

## When the user answers a rule

At step 7, in the same turn the record's line becomes `approved`, rewrite that rule's row: its origin becomes a decision with the date, and its text the answer. When the answer replaces a rule a card shows as stated, correct that card as «After the design stage» sets: the stated row stays, marked as cancelled, beside the new one. When step 7 removes the alternatives not chosen, remove their cards and `card:` lines with them.

## After the design stage

When a later flow records a change to a rule a card shows, as [business rules](../../flow-plan/references/business-rules.md) sets, the main thread also does one of these in the same turn:

- With a connection that can write to that Figma file, correct the card: mark the earlier row as cancelled, with the date and the reason, and add a row for the rule that replaces it; never delete the earlier row. Change only the card nodes the example's record lists in its `card:` lines, which the global guidance's «Scope and authorization» authorizes without a question. Never change a frame or any other node from outside the design stage.
- Without that connection, without those `card:` lines where the flow runs, or when the write fails, add `card stale: <screen> · <rule>` to the work's rule record.

When a decision contradicts what an approved frame itself shows, such as a region, an action, a state or a text of the frame, add `frame stale: <node> · <rule>` to the work's rule record and say so in that turn's message. The build follows the decision, and no flow redraws the frame outside the design stage. The close names each `card stale:` and `frame stale:` line in its final report, as left open.

When [flow-figma-design](../SKILL.md) runs again for that work, read those lines and the record's `cancelled` rules, wherever the rule record now lives, before drawing; correct the cards and the frames they name; and name each line you corrected in the stage's report. The plan's revision then removes those lines, as [business rules](../../flow-plan/references/business-rules.md) sets; for work without a change folder, remove them from the checkpoint file yourself.

## Special cases

- **A request for the design itself.** When the request names existing frames as its target, draw the cards beside those frames, and move none of them without asking.
- **Unattended delegation.** Draw the cards with the rules as the record holds them; a `proposed` rule stays as still to decide.
