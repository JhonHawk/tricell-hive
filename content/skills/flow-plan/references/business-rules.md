# Business rules

Read when a work item decides, assumes, replaces or cancels a business rule, in any flow, and before each write of its rules to its ticket: when the user approves a design example, when the build starts and at the close. This is the one home of the rule record and of that tracker write, which the global guidance's «Scope and authorization» authorizes.

## What a business rule is

A business rule is a behavior of the product in a situation: what an action requires, who may take it, which state follows it, what it blocks or keeps, or where a setting that applies beyond one screen lives. A choice of layout, wording, component or code structure that changes none of those is not one.

A rule is **stated** when the request, the ticket, in its description, its criteria or its comments, or the documentation the work read says it. The section and the comments that the writes below add to a ticket are copies of the record, never a source: a rule found only there is not stated. This reference governs what a work item adopts beyond the stated rules: each rule the user decides, each rule the agent assumes, and each stated rule a decision replaces.

## The record

Each work item has one rule record, a list of lines in this format:

`rule: <the rule> · screen <name, or none> · <status> · found by <drawer | reviewer | research | plan | build | verify>`

Write the rule in plain words a reader outside the conversation understands, never as a decision ID alone. The status is one of:

- `proposed`: the user has not decided it yet.
- `assumption`: adopted without asking, which is allowed only when every option leaves the scope, a contract, stored data, the cost, reversibility and who may act unchanged.
- `scope`: it changes the scope, as step 5 of [flow-figma-design](../../flow-figma-design/SKILL.md) sets.
- `approved <date>: <answer>`: the user decided it.
- `stated: <source>`: a source already says it. Only an example's record holds this status, for its rule card, as [rule cards](../../flow-figma-design/references/rule-cards.md) sets; the plan does not copy these lines.
- `cancelled <date>: <reason> · replaced by <the rule that replaces it>`: a later decision replaced it.

Never delete a rule's line or change the rule it states. When a decision replaces a rule the record holds, change that line's status to `cancelled` and add the new rule as its own line. When it replaces a rule that only the ticket or the documentation states, add that rule as a line, quoted, with where it was stated, already `cancelled`, and add the new rule. The reason says why the earlier rule no longer holds, in words its author can read.

The record lives in one place at a time:

- During the design stage, the `examples.md` that [flow-figma-design](../../flow-figma-design/SKILL.md) writes.
- Once the plan writes a change folder, the `## Business rules` section of its `proposal.md`, which governs from then on. The plan copies there the `rule:` lines of `examples.md`, except the `stated` ones, and the ticket's control lines. When a design stage runs again it writes its lines in `examples.md`, and the plan's revision copies the changed ones and removes the `card stale:` and `frame stale:` lines that stage corrected.
- For work without a change folder, once its plan or its build starts, the work's checkpoint file: the file named after the work in the host's temporary location that [flow-build](../../flow-build/SKILL.md) keeps, created for the record when the work has none. It receives the same lines and governs from then on.

Besides the rules, the record holds the ticket's control lines, `ticket section: <key> · <heading>` and `ticket write: <date> · <moment> · <key>`, and the `card stale:` and `frame stale:` lines that [rule cards](../../flow-figma-design/references/rule-cards.md) sets. The `card:` lines stay in the example's record.

## When a rule enters the record

Write the record in the same turn, in any flow, when one of these happens:

- The user decides a rule no source stated, or one that replaces a stated or recorded rule: write it `approved`, and the replaced rule `cancelled`.
- The main thread adopts a rule without asking: write it `assumption` and state it in the next message to the user, where a reply can turn it into a decision. A rule that fails the test of `assumption` above is asked as a decision, as the global guidance's «Questions and routes» sets, never assumed.
- A child reports a rule it had to assume to finish its task: check it against the sources, record it the same way and state it. A child never writes the record, the ticket or the design file; the brief of a child that builds or draws asks it to report each such rule.
- A flow starts after research: write as `approved` each rule the user decided in research that no source states. Research itself writes no record and no ticket.

When the rule that changes is on a rule card, or a decision contradicts an approved frame, also follow «After the design stage» of [rule cards](../../flow-figma-design/references/rule-cards.md).

## The ticket

The main thread writes the record's rules to the work's ticket at three moments, without asking the user:

1. when the user approves a design example, at step 7 of [flow-figma-design](../../flow-figma-design/SKILL.md);
2. when the build starts, with the ticket's move to in progress, at step 1 of [flow-build](../../flow-build/SKILL.md);
3. at the work item's completion point, at step 3 of [flow-close](../../flow-close/SKILL.md), which defines that point.

Write nothing to a tracker when the work names no ticket or was abandoned; the close then lists the record's rules in its final report. An operational task writes only when its request's deliverable is the design and names its ticket, or when the user asks.

Make each write through the tracker's own CLI or tool, in the tracker that declares the ticket's key, and in each ticket when the work names more than one. A write has two parts.

**The section.** Read the description right before the write. The section runs from one heading, `Business rules · <the work's name>`, written in the ticket's language, to its closing line, the same text followed by ` · end`. On the first write, add both at the end of the description, with the section between them, and record the heading in a `ticket section:` line. On a later write, replace only what lies between those two lines. When either line is missing or appears more than once, do not edit the description: put the section in the comment instead and say so in the report. Leave every other character of the description as it was: never edit, reorder or delete its author's text, not even a sentence a rule cancels.

The section opens with the date of the write and, when a design file holds the work's frames, its location. Then it holds these groups, in this order, each only when it has rules:

- **Decided:** each `approved` rule with its date.
- **Assumptions:** each `assumption`, as what applies unless someone says otherwise.
- **To decide:** each `proposed` or `scope` rule.
- **Cancelled:** each `cancelled` rule, quoted, with where it was stated, the reason and the rule that replaces it.

Write the group names in the ticket's language and each rule in plain words: no decision ID alone, no path or identifier of the repository, never a secret value. Leave out the `stated` rules, which the ticket or its sources already say.

Skip the write when the groups you would write equal the ones the ticket's section already holds.

**The comment.** Add one comment per write: one line per rule added or decided since the section the ticket held and, for each cancellation, what was cancelled, why and what replaces it. The comment is how the author of a cancelled rule learns of it.

After both parts, read the description back, check that the section is there and that the text outside its two lines did not change, and add one `ticket write:` line per ticket written. When the tool cannot edit the description, put the whole section in the comment and say so in the report. Name each write in that stage's message: the ticket, and how many rules it added and cancelled. Under [unattended delegation](../../unattended-delegation/SKILL.md) the writes run the same way, and the rules still `proposed` go under **To decide**.

**At the close.** Closing a work item's business rules means making write 3 when it applies and otherwise listing the record's rules in the final report. Either way, name in that report each `card stale:` and `frame stale:` line of the record, as «After the design stage» of [rule cards](../../flow-figma-design/references/rule-cards.md) sets.

## Documentation in the repository

A rule that changes an observable behavior goes in the plan's spec delta, as [flow-plan](../SKILL.md#change-folder) sets. When a file of the repository states a rule the work cancels or replaces, such as a requirements document, the plan has a task that updates it. Documentation outside the repository, such as a file a person sent, is never edited: the ticket's section records the change.
