# Choose, record and close

Read at step 6 of [create-design](../SKILL.md), when the frames are drawn and reviewed, again at step 7, in the mobile phase it sets, and at step 8.

## Choose

Finish drawing and reviewing the work's screens, the set of the phase in course (the web set first, then the mobile set of the [mobile phase](#mobile-phase)), before asking approval, and list screens stopped by a `scope` rule apart. Ask one approval for the completed set, never per screen: approve the whole set, or request corrections naming the affected screens. Show each frame's location and composition; for alternatives, name each screen's recommended example, which the user can replace with another alternative in the corrections. Corrections leave only the named screens pending and approve the others; after them, ask one approval decision covering only the affected screens, never the others' again.

In the same round, ask each `proposed` rule as a decision of its own, and word the set's approve option to hold under every answer, naming that decision, such as "approve the composition, with the rule you choose in D12; a different answer is redrawn and shown to you again"; list the assumptions in the message as [questions](../../hive-core/messages/questions.md) «Decide and list» sets. When a rule's answer differs from what is drawn, redraw that part and keep the screen pending, even when its approve option was chosen, until the user approves it as redrawn. Redraw the example for each correction and update the record; a correction the user asks for gets no other review.

When a correction of the user sets a convention for screens beyond this one, such as the row actions of every list, ask in the round that asks the approval, once per convention, whether to record it among the **Decisions** of the app's `DESIGN.md`; a yes is the user's instruction, kept in the record for the plan to act on. Ask in the same round whether to record the file or the library, as [design library](design-library.md)'s «Record it» sets.

## Record the approval

In the same turn the user answers a rule, write its line as `approved <date>: <answer>` and update its card's row. In the same turn the user answers the set's approval decision:

- replace each approved screen's status line with `approved <date>`;
- leave unapproved only the screens named for correction or kept pending for a redraw by the choice above, and preserve the others' approval;
- add the user's answer about each convention;
- keep the answer about the file and the library as «Record it» sets, a yes as `pending` with nothing written in the repository;
- remove the alternatives not chosen, with the components, variables and `candidate:` lines only they used, from the file and from the record; the approved examples of the other screens stay.

Then write the work's rules to its ticket, as [business rules ticket](../../hive-core/business/business-ticket.md) sets. Write the approved example in the report. A plan with a plan file copies the record into the `UI` part of its `## Design` and its `rule:` lines into its `## Business rules`; work without one copies both into its checkpoint file, so a stage in another checkout finds them.

## Mobile phase

Read when the work has a web app served on mobile, as [draw](draw.md#width) sets, once every web screen not stopped by a `scope` rule is approved; a screen kept pending by a correction or a rule's answer holds the phase until the user approves it. An app that runs only on mobile has one phase, so the stage skips this section for it. In the turn the last of those approvals is recorded, in this order:

1. Draw the 440 px frames [draw](draw.md#width) sets, in the mobile twin section [sections](sections.md) sets, from the approved web frames and their pieces.
2. Record each mobile frame in `examples.md`, as [example record](example-record.md) sets.
3. Have the mobile frames reviewed by one `hive-review-ux` child, briefed as [frame review brief](frame-review-brief.md) sets, without the cards, so it lists the unsettled mobile rules on its own; record and fix what holds as [example record](example-record.md#the-review) sets.
4. Add the «on mobile» rows to the web cards, as [rule cards](rule-cards.md) sets, then have only the corrected frames and the changed cards reviewed again once.
5. Ask the mobile set's approval as «Choose» sets and record it as «Record the approval» sets.

A correction to a web screen that the user asks for during the mobile phase redraws that screen's mobile frames too, and both return to approval together. When the stage runs again for screens a `scope` rule stopped, their mobile frames follow the same phase once their web frames are approved.

## Close the stage

When the record holds a `scope` rule, ask the approval of the other screens as the choice above sets and, for a web app served on mobile, run the [mobile phase](#mobile-phase) for them, then end with the next-step block toward [hive-mode](../../hive-mode/SKILL.md) naming those rules instead of the block toward the plan. When the stage runs again for this work, step 5 resumes from its record. A request whose deliverable is the design ends as [A request for the design itself](special-cases.md#a-request-for-the-design-itself) sets, not here.

Otherwise, once every screen and, for a web app served on mobile, its mobile frames are approved, end with the next-step block toward [flow-plan](../../flow-plan/SKILL.md), as [the next-step block](../../hive-core/messages/report-endings.md#the-next-step-block) sets, naming the approved example and its record. Recommend `flow-build` instead only when all of these hold:

- the block that offered this stage listed `flow-build` as a reply;
- the approved examples add no region the request did not name;
- the record holds no yes to recording a convention and no `pending` file or library;
- no rule in it is still `proposed`.

The plan takes the example as the screen's reference, as [UI decisions](../../hive-core/ui/ui-plan.md) sets, and adds the task that writes each convention, and a `pending` file or library, when the record holds that instruction; a build without a plan keeps the record in its checkpoint file, as [checkpoint](../../hive-core/plan/checkpoint.md) sets.
