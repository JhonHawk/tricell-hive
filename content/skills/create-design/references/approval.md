# Choose, record and close

Read at step 6 of [create-design](../SKILL.md), when the frames are drawn and reviewed, and again at steps 7 and 8.

## Choose

Finish drawing and reviewing the work's screens, including the 440 px frames of web screens served on mobile, before asking approval, and list screens stopped by a `scope` rule apart. Ask one approval for the completed set, never per screen: approve the whole set, or request corrections naming the affected screens. Show each frame's location and composition; for alternatives, name each screen's recommended example, which the user can replace with another alternative in the corrections. Corrections leave only the named screens pending and approve the others; after them, ask one approval decision covering only the affected screens (with their 440 px frames), never the others' again.

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

## Close the stage

When the record holds a `scope` rule, ask the approval of the other screens as the choice above sets, then end with the next-step block toward [hive-mode](../../hive-mode/SKILL.md) naming those rules instead of the block toward the plan. When the stage runs again for this work, step 5 resumes from its record. A request whose deliverable is the design ends as [A request for the design itself](special-cases.md#a-request-for-the-design-itself) sets, not here.

Otherwise, once every screen is approved, end with the next-step block toward [flow-plan](../../flow-plan/SKILL.md), as [the next-step block](../../hive-core/messages/report-endings.md#the-next-step-block) sets, naming the approved example and its record. Recommend `flow-build` instead only when all of these hold:

- the block that offered this stage listed `flow-build` as a reply;
- the approved examples add no region the request did not name;
- the record holds no yes to recording a convention and no `pending` file or library;
- no rule in it is still `proposed`.

The plan takes the example as the screen's reference, as [UI decisions](../../hive-core/ui/ui-plan.md) sets, and adds the task that writes each convention, and a `pending` file or library, when the record holds that instruction; a build without a plan keeps the record in its checkpoint file, as [checkpoint](../../hive-core/plan/checkpoint.md) sets.
