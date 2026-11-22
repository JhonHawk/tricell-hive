# Write business rules to the ticket

Read before writing rules to a ticket. Each write is a work item's business rules to its ticket, made when the user approves a design example, when the build starts and at the close.

The main thread writes the record's rules to the work's ticket at three moments, without asking the user:

1. when the user approves a design example, at step 7 of [create-design](../../create-design/SKILL.md);
2. when the build starts, with the ticket's move to in progress, at step 1 of [flow-build](../../flow-build/SKILL.md);
3. at the work item's completion point, at step 3 of [flow-close](../../flow-close/SKILL.md), at the point [ticket state](../tracker/ticket-state.md) defines.

Write nothing when the work names no ticket or was abandoned; the close then lists the record's rules in its final report. An operational task writes as the global «Scope and authorization» sets. Make each write through the tracker's own CLI or tool, in the tracker that declares the ticket's key, and in each ticket when the work names more than one. A write has two parts.

**The section.** Read the description right before the write. The section runs from one heading, `Business rules · <the work's name>`, written in the ticket's language, to its closing line, the same text followed by ` · end`. Leave every other character of the description as it was: never edit, reorder or delete its author's text, not even a sentence a rule cancels. With `python3`, save the description and the rendered section body to local files and run [business_rules.py](../../flow-plan/scripts/business_rules.py) `splice-description` (`python3 -I`; `--help` lists its arguments): write the output on `ok` and `changed`, also record the heading in a `ticket section:` line on `first_write`, skip the write when unchanged, and put the section in the comment and report it on `fallback_to_comment`. Without Python, use the comment fallback.

The section opens with the date of the write and, when a design file holds the work's frames, its location. Then it holds these groups, in this order, each only when it has rules: **Decided** (each `approved` rule with its date), **Assumptions** (each `assumption`, as what applies unless someone says otherwise), **To decide** (each `proposed` or `scope` rule) and **Cancelled** (each `cancelled` rule, quoted, with where it was stated, the reason and the rule that replaces it). Write the group names in the ticket's language and each rule in plain words: no decision ID alone, no repository path or identifier, never a secret value. Leave out the `stated` rules.

**The comment.** One comment per write: one line per rule added or decided since the section the ticket held and, for each cancellation, what was cancelled, why and what replaces it. It is how the author of a cancelled rule learns of it.

After both parts, check that the text outside the section did not change, and add one `ticket write:` line per ticket written. When the tool cannot edit the description, put the whole section in the comment and say so. Name each write in that stage's message: the ticket, and how many rules it added and cancelled. Under [unattended delegation](../../unattended-delegation/SKILL.md) the writes run the same way, and the rules still `proposed` go under **To decide**.
