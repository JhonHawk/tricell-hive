# Business-rule record

Read before editing a rule record line. It holds the format, statuses and location of a work item's business-rule record. [Business rules](business-rules.md) says when a rule enters it.

Each work item has one rule record, a list of lines in this format:

`rule: <the rule> · screen <name, or none> · <status> · found by <drawer | reviewer | research | plan | build | verify>`

Write the rule in plain words a reader outside the conversation understands. The status is one of:

- `proposed`: the user has not decided it yet.
- `assumption`: adopted without asking, as [questions](../messages/questions.md) «Decide and list» sets.
- `scope`: it changes the scope, as step 5 of [create-design](../../create-design/SKILL.md) sets.
- `approved <date>: <answer>`: the user decided it.
- `stated: <source>`: a source already says it. Only an example's record holds this status, for its rule card, as [rule cards](../../create-design/references/rule-cards.md) sets; the plan does not copy these lines.
- `cancelled <date>: <reason> · replaced by <the rule that replaces it>`: a later decision replaced it.

Never delete a rule's line or change the rule it states. When a decision replaces a rule the record holds, change that line's status to `cancelled` and add the new rule as its own line. When it replaces a rule that only the ticket or the documentation states, add that rule as a quoted line with where it was stated, already `cancelled`, and add the new rule. The reason says why the earlier rule no longer holds, in words its author can read.

The record lives in one place at a time:

- During the design stage, the `examples.md` that [create-design](../../create-design/SKILL.md) writes.
- Once the plan writes a change folder, the `## Business rules` section of its `proposal.md`, which governs from then on. The plan copies there the `rule:` lines of `examples.md` except the `stated` ones, and the ticket's control lines; when a design stage runs again, its revision copies the changed lines and removes the `card stale:` and `frame stale:` lines that stage corrected.
- For work without a change folder, once its plan or build starts, the work's checkpoint file that [flow-build](../../flow-build/SKILL.md) keeps.

The record also holds the ticket's control lines, `ticket section: <key> · <heading>` and `ticket write: <date> · <moment> · <key>`, and the `card stale:` and `frame stale:` lines of [rule cards](../../create-design/references/rule-cards.md); the `card:` lines stay in the example's record.
