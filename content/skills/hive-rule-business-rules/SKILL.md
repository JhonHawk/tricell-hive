---
name: hive-rule-business-rules
description: "Hive rule leaf: read when a decision or an assumption adds, replaces or cancels a business rule, or when a work item closes its rules. What a business rule is, when it enters the record."
disable-model-invocation: true
---

# Business rules

Read when a work item decides, assumes, replaces or cancels a business rule, in any flow. This leaf, the [record](../hive-rule-business-record/SKILL.md) and the [ticket write](../hive-rule-business-ticket/SKILL.md) are the one home of the rule record and of that tracker write, which the global guidance's «Scope and authorization» authorizes.

## What a business rule is

A business rule is a behavior of the product in a situation: what an action requires, who may take it, which state follows it, what it blocks or keeps, or where a setting that applies beyond one screen lives. A choice of layout, wording, component or code structure that changes none of those is not one.

A rule is **stated** when the request, the ticket, in its description, its criteria or its comments, or the documentation the work read says it. The section and the comments that the writes below add to a ticket are copies of the record, never a source: a rule found only there is not stated. This leaf governs what a work item adopts beyond the stated rules: each rule the user decides, each rule the agent assumes, and each stated rule a decision replaces.

## When a rule enters the record

When a decision, or a rule you or a child assumed, adds, replaces or cancels a business rule, record it in the same turn, in any flow, as the matching case sets; write the record when one of these happens:

- The user decides a rule no source stated, or one that replaces a stated or recorded rule: write it `approved`, and the replaced rule `cancelled`.
- The main thread adopts a rule without asking: write it `assumption` and state it in the next message to the user, where a reply can turn it into a decision. A rule that fails the test of `assumption` above is asked as a decision, as [decision format](../hive-rule-questions/SKILL.md) sets, never assumed.
- A child reports a rule it had to assume to finish its task: check it against the sources, record it the same way and state it. A child never writes the record, the ticket or the design file; the brief of a child that builds or draws asks it to report each such rule.
- A flow starts after research: write as `approved` each rule the user decided in research that no source states. Research itself writes no record and no ticket.

When the rule that changes is on a rule card, or a decision contradicts an approved frame, also follow «After the design stage» of [rule cards](../flow-figma-design/references/rule-cards.md).

## Documentation in the repository

A rule that changes an observable behavior goes in the plan's spec delta, as [flow-plan](../hive-rule-change-folder/SKILL.md#change-folder) sets. When a file of the repository states a rule the work cancels or replaces, such as a requirements document, the plan has a task that updates it. Documentation outside the repository, such as a file a person sent, is never edited: the ticket's section records the change.

## At the close

Closing a work item's business rules means making write 3 of the [ticket write](../hive-rule-business-ticket/SKILL.md) when it applies and otherwise listing the record's rules in the final report. Either way, name in that report each `card stale:` and `frame stale:` line of the record, as «After the design stage» of [rule cards](../flow-figma-design/references/rule-cards.md) sets.
