---
name: hive-rule-business-record
description: "Hive rule leaf: read when writing or changing a line of a work item's business-rule record: the line format, the statuses, never deleting a line, and where the record lives at each stage."
disable-model-invocation: true
---

# Business-rule record

Read when you write or change a line of a work item's business-rule record: its format, statuses and where it lives. [Business rules](../hive-rule-business-rules/SKILL.md) says when a rule enters it.

Each work item has one rule record, a list of lines in this format:

`rule: <the rule> · screen <name, or none> · <status> · found by <drawer | reviewer | research | plan | build | verify>`

Write the rule in plain words a reader outside the conversation understands, never as a decision ID alone. The status is one of:

- `proposed`: the user has not decided it yet.
- `assumption`: adopted without asking, which is allowed only when none of the criteria a–f of [questions](../hive-rule-questions/SKILL.md) «What to ask» holds; a reply that accepts the list leaves it `assumption`.
- `scope`: it changes the scope, as step 5 of [flow-figma-design](../flow-figma-design/SKILL.md) sets.
- `approved <date>: <answer>`: the user decided it.
- `stated: <source>`: a source already says it. Only an example's record holds this status, for its rule card, as [rule cards](../flow-figma-design/references/rule-cards.md) sets; the plan does not copy these lines.
- `cancelled <date>: <reason> · replaced by <the rule that replaces it>`: a later decision replaced it.

Never delete a rule's line or change the rule it states. When a decision replaces a rule the record holds, change that line's status to `cancelled` and add the new rule as its own line. When it replaces a rule that only the ticket or the documentation states, add that rule as a line, quoted, with where it was stated, already `cancelled`, and add the new rule. The reason says why the earlier rule no longer holds, in words its author can read.

The record lives in one place at a time:

- During the design stage, the `examples.md` that [flow-figma-design](../flow-figma-design/SKILL.md) writes.
- Once the plan writes a change folder, the `## Business rules` section of its `proposal.md`, which governs from then on. The plan copies there the `rule:` lines of `examples.md`, except the `stated` ones, and the ticket's control lines. When a design stage runs again it writes its lines in `examples.md`, and the plan's revision copies the changed ones and removes the `card stale:` and `frame stale:` lines that stage corrected.
- For work without a change folder, once its plan or its build starts, the work's checkpoint file: the file named after the work in the host's temporary location that [flow-build](../flow-build/SKILL.md) keeps, created for the record when the work has none. It receives the same lines and governs from then on.

Besides the rules, the record holds the ticket's control lines, `ticket section: <key> · <heading>` and `ticket write: <date> · <moment> · <key>`, and the `card stale:` and `frame stale:` lines that [rule cards](../flow-figma-design/references/rule-cards.md) sets. The `card:` lines stay in the example's record.
