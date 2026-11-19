# Example record and review

Read at step 5 of [flow-figma-design](../SKILL.md) as soon as the frames are drawn, before the review and before asking anything: it sets the rules to list, the record to write and the review that follows.

## Rules the frames show

While drawing, list each business rule the frames show that step 2's sources do not settle and that the siblings or their code do not already implement. [Business rules](../../hive-rule-business-rules/SKILL.md) defines a rule and the line that records it. Draw the behavior you recommend and record the rule as `proposed`, so approving the screen does not decide it unseen. A rule that passes that leaf's test of `assumption` is one: state it in step 6's message instead of asking it. A rule that adds a capability the request did not name, changes data or settings that other screens or modules use, or changes a permission is a `scope` rule: mark each screen it shapes `stopped by <the rule>`, delete any part of their frames you drew, and finish step 5 for the other screens; stopped screens are not reviewed, step 6 does not ask the rule, and step 8 ends toward research.

## The record

Write in `examples.md`, created with the inventory in the work's folder under `_support/workspace/`, placed as [layout](../../project-layout/references/layout.md) sets:

- the tool, the file's location, each frame's node, and the identifiers of the components and variables this work created;
- the path of each frame's local image in that folder's `images/`;
- one status line per screen of the work, `drawn, under review` or `stopped by <the rule>`;
- apart from those, one `rule:` line per rule, `proposed`, `assumption` or `scope`, found by the drawer or the reviewer;
- one `candidate: <piece> · <node>` line per piece defined outside the bound library, which step 6's message lists.

The location and the nodes are the record: the image is a local copy, read again from them where it is missing.

## The review

Have one `hive-review-ux` child review the frames against the request, the siblings, the recorded decisions and their alignment, briefed as [Reviewer brief](../../hive-rule-reviewer-brief/SKILL.md) sets for frames in a design file. The reviewer lists apart the rules it found unsettled: check each one the record lacks against step 2's sources and the siblings and their code, drop those they settle, and record the rest, found by the reviewer, as the drawer's rules above, a `scope` one stopping its screen. Fix the findings that hold as step 5 sets for a correction, then draw the rule cards beside the frames as [rule cards](rule-cards.md) sets, have only the corrected frames and the cards reviewed again once, and set each status line that reads `drawn, under review` to `drawn, pending the user's approval`. Name in step 6's message the findings that remain, the checks the review reports as not run and a verdict still not passed.
