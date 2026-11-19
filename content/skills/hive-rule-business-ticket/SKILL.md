---
name: hive-rule-business-ticket
description: "Hive rule leaf: read before writing a work item's business rules to its ticket: the description section, the comment, the readback, and the three moments of the write."
disable-model-invocation: true
---

# Write business rules to the ticket

Read before each write of a work item's business rules to its ticket: when the user approves a design example, when the build starts and at the close.

The main thread writes the record's rules to the work's ticket at three moments, without asking the user:

1. when the user approves a design example, at step 7 of [flow-figma-design](../flow-figma-design/SKILL.md);
2. when the build starts, with the ticket's move to in progress, at step 1 of [flow-build](../flow-build/SKILL.md);
3. at the work item's completion point, at step 3 of [flow-close](../flow-close/SKILL.md), at the point [ticket state](../hive-rule-ticket-state/SKILL.md) defines.

Write nothing to a tracker when the work names no ticket or was abandoned; the close then lists the record's rules in its final report. An operational task writes only when its request's deliverable is the design and names its ticket, or when the user asks.

Make each write through the tracker's own CLI or tool, in the tracker that declares the ticket's key, and in each ticket when the work names more than one. A write has two parts.

**The section.** Read the description right before the write. The section runs from one heading, `Business rules · <the work's name>`, written in the ticket's language, to its closing line, the same text followed by ` · end`. Leave every other character of the description as it was: never edit, reorder or delete its author's text, not even a sentence a rule cancels. When `python3` runs it, save the description and rendered section body to local files and run `python3 -I <flow-plan-dir>/scripts/business_rules.py splice-description --description-file <description-file> --heading <heading> --body-file <body-file> --out <output-file>` ([splicer](../flow-plan/scripts/business_rules.py)). On `ok` and `changed`, write the output; on `first_write`, record the heading in a `ticket section:` line. On unchanged, skip the write; on `fallback_to_comment`, put the section in the comment and report it. Without Python, preserve the same marker boundaries, author text and unchanged comparison ignoring the generated date; if those cannot be verified, use the comment fallback.

The section opens with the date of the write and, when a design file holds the work's frames, its location. Then it holds these groups, in this order, each only when it has rules:

- **Decided:** each `approved` rule with its date.
- **Assumptions:** each `assumption`, as what applies unless someone says otherwise.
- **To decide:** each `proposed` or `scope` rule.
- **Cancelled:** each `cancelled` rule, quoted, with where it was stated, the reason and the rule that replaces it.

Write the group names in the ticket's language and each rule in plain words: no decision ID alone, no path or identifier of the repository, never a secret value. Leave out the `stated` rules, which the ticket or its sources already say.

**The comment.** Add one comment per write: one line per rule added or decided since the section the ticket held and, for each cancellation, what was cancelled, why and what replaces it. The comment is how the author of a cancelled rule learns of it.

After both parts, read the description back, check that the section is there and that the text outside its two lines did not change, and add one `ticket write:` line per ticket written. When the tool cannot edit the description, put the whole section in the comment and say so in the report. Name each write in that stage's message: the ticket, and how many rules it added and cancelled. Under [unattended delegation](../unattended-delegation/SKILL.md) the writes run the same way, and the rules still `proposed` go under **To decide**.
