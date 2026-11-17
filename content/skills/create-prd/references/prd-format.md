# PRD format

Read before drafting a requirement document and before checking one, at [create-prd](../SKILL.md) steps 4, 5 and 7, in its [incoming document](incoming-document.md) and [technical review](technical-review.md) uses, and when the `hive-review-prd` role reviews one. Reading this reference writes no document.

## Sections

The document opens with its title, naming the module or feature, its version, its state, its date, and its owner: the person who decides its rules. Then it has these sections, in this order, with their names written in the document's language:

1. **Problem**: what happens today, to whom, and what it costs them, in two to five sentences. It names no solution.
2. **Goals**: the outcomes the change must produce, each one observable after release.
3. **Non-goals**: at least one thing a reader could expect from the title and will not get.
4. **User stories**: who does what to get what. Each story names the codes of the rules it needs.
5. **Requirements**: the rules, one entry each, as «Rule entry» sets.
6. **Acceptance criteria**: what a person other than the author can observe to check a rule or a story. Each criterion names the rule code it checks.
7. **Success metrics**: how anyone will know after release that each goal was met: the measure, its value today when known, and its target. An unknown target is an open question, never an invented number.
8. **Open questions**: each question with who answers it and the codes of the rules it holds back.
9. **Version history**: one entry per version, as «Versions» sets.

After a technical review, the document also carries, at its end, the section that [technical review](technical-review.md) sets.

The document never holds a screen layout or a component choice, which belong to the design; a technical how, such as an architecture, a schema, an endpoint or a library, which belongs to the plan; or a date, an estimate or an assignment, which belong to the schedule.

## Rule entry

Each requirement is one rule, written as an entry of its own with six parts. Never write the rules as rows of a table: a rule's text in a narrow column is hard to read. An entry has three blocks, in this order:

1. A heading, one level below the heading of the requirements section, with the code, the priority and the state: `R<n> · P0 | P1 | P2 · <state>`. A `to decide` state names who answers it; a `cancelled` state names the version that cancelled it and the code of the rule that replaces it.
2. The rule, as a paragraph of its own.
3. A list of two items: `Source: <where it comes from>` and `Verdict: <verdict against the code>`.

```text
### R8 · P1 · to decide (Ana answers)

A refund over $500 is paid only after a supervisor approves it.

- Source: Ana, 2026-11-10
- Verdict: contradicts; today no approval is asked
```

- **Code**: `R1`, `R2` and so on, in the order the rules were created. A code is stable across versions and is never reused or renumbered, since tickets, designs and plans cite a rule by its code.
- **Priority**: `P0` when the feature does not ship without the rule; `P1` when the feature works without it and this delivery still expects it; `P2` when it can wait for a later delivery.
- **The rule**: one behavior of the product in a situation: what an action requires, who may take it, which state follows it, or what it blocks or keeps. A sentence that states two behaviors is two rules.
- **State**: one of the four below.
- **Source**: who stated the rule and when, or the document and section that states it; for an `assumption`, the writer.
- **Verdict**: one of the five below; `not checked` until someone compares the rule with the code.

Write the state, the verdict and the labels `Source` and `Verdict` in the document's language, with the meaning this reference gives them.

| State | Meaning |
| --- | --- |
| `decided` | Its owner decided it; the source says who and when. |
| `assumption` | The writer completed it; it holds unless the owner says otherwise. |
| `to decide` | Nobody decided it yet; the entry's heading says who answers it. |
| `cancelled` | A later version replaced it; the entry's heading says which version cancelled it and the code of the rule that replaces it. |

| Verdict | Meaning | Effect |
| --- | --- | --- |
| `exists` | The code already does it. | None. |
| `new` | The code does not do it and nothing in the code prevents it. | None. |
| `contradicts` | The code does something else today. | A question for the owner. |
| `blocked` | The rule needs data, a relation or a permission that does not exist and that no rule of the document provides. | A question for the owner. |
| `not checked` | Nobody compared it with the code yet. | Holds back `ready for development`. |

Write a verdict in words its author can read, such as "today a supervisor is not asked", and keep code locations for the technical review section.

## Ready list

A version passes when all ten hold; report each failure. [check_prd.py](../scripts/check_prd.py) checks structure in local Markdown; judge meaning and coverage separately.

1. The nine sections are present, in order, and the non-goals name at least one thing.
2. Every rule entry has its six parts, and no rule is written as a row of a table.
3. Each rule states one behavior.
4. No vague term, such as "fast", "user-friendly", "as needed" or "etc.", and no "to be defined" marker stands in a rule or a criterion.
5. Every `P0` and `P1` rule has at least one acceptance criterion, and a person other than the author can observe each criterion.
6. No two rules contradict each other.
7. For each action a user story names, the rules say who may take it and what happens when it is refused or fails.
8. Every open question says who answers it.
9. No rule states a screen layout, a technical how or a date.
10. Every goal has a success metric, or an open question about it.

## States

The state of the document is set by what the checks support. A label in the source, such as "ready for development", is a claim to check and never sets the state. Exactly one state holds for a version. An open question that holds back no rule does not hold back an approval.

| State | It is reached when |
| --- | --- |
| `draft` | The document is still being written: check 1 or check 2 of the ready list fails. |
| `iterating` | Checks 1 and 2 pass, and the version does not meet what `approved by business` requires. |
| `approved by business` | All four hold for this version: the ready list passes, a `hive-review-prd` child reviewed it, no rule is `to decide`, and the user approved it explicitly, as [create-prd](../SKILL.md) step 7 asks. |
| `technical review` | A session with the repository took a version that is `approved by business` and is comparing it with the code. |
| `ready for development` | The technical review set it, as [technical review](technical-review.md) sets. No other step sets this state. |

A change to any rule after an approval produces a new version, which starts in `iterating`; the approved version stays frozen as it was.

## Versions

Versions are numbered from 1. A new version starts each time the document changes after the skill reported it, at its step 9, not at each edit before that. Write the version history as a list, one item per version and never a table: the version, its date and the state it reached on the item's first line, then, nested under it, who approved it and the link or path of its frozen copy when it has them, and its list of changes:

- **Added**: the codes of the new rules.
- **Decided**: each rule that left `to decide` or `assumption`, with the answer and who gave it.
- **Changed**: each rule whose wording changed without changing its behavior, with what changed.
- **Cancelled**: each cancelled rule, with the reason and the code of the rule that replaces it.

Never delete a rule's entry. When an answer changes the behavior a rule states, cancel that rule and add the new behavior as a rule with a new code; keep the code only when the wording changes and the behavior does not.
