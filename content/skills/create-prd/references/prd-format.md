# PRD format

Read before drafting a requirement document and before checking one, at [create-prd](../SKILL.md) steps 4, 5 and 7, and when the `hive-review-prd` role reviews one. Reading this reference writes no document.

## Sections

The document opens with its title, naming the module or feature, its version, its state, its date, and its owner: the person who decides its rules. Then it has these sections, in this order, with their names written in the document's language:

1. **Problem**: what happens today, to whom, and what it costs them, in two to five sentences. It names no solution.
2. **Goals**: the outcomes the change must produce, each one observable after release.
3. **Non-goals**: at least one thing a reader could expect from the title and will not get.
4. **User stories**: who does what to get what. Each story names the codes of the rules it needs.
5. **Requirements**: the rules, one per line, as «Rule line» sets.
6. **Acceptance criteria**: what a person other than the author can observe to check a rule or a story. Each criterion names the rule code it checks.
7. **Success metrics**: how anyone will know after release that each goal was met: the measure, its value today when known, and its target. An unknown target is an open question, never an invented number.
8. **Open questions**: each question with who answers it and the codes of the rules it holds back.
9. **Version history**: one row per version, as «Versions» sets.

After a technical review, the document also carries, at its end, the section that [technical review](technical-review.md) sets.

The document never holds a screen layout or a component choice, which belong to the design; a technical how, such as an architecture, a schema, an endpoint or a library, which belongs to the plan; or a date, an estimate or an assignment, which belong to the schedule.

## Rule line

Each requirement is one rule on one line with six parts, written as a line or as a row of a table with these columns:

`R<n> · P0 | P1 | P2 · <the rule> · <state> · source: <where it comes from> · code: <verdict>`

- **Code**: `R1`, `R2` and so on, in the order the rules were created. A code is stable across versions and is never reused or renumbered, since tickets, designs and plans cite a rule by its code.
- **Priority**: `P0` when the feature does not ship without the rule; `P1` when the feature works without it and this delivery still expects it; `P2` when it can wait for a later delivery.
- **The rule**: one behavior of the product in a situation: what an action requires, who may take it, which state follows it, or what it blocks or keeps. A sentence that states two behaviors is two rules.
- **State**: one of the four below.
- **Source**: who stated the rule and when, or the document and section that states it; for an `assumption`, the writer.
- **Verdict**: one of the five below; `not checked` until someone compares the rule with the code.

Write the state and the verdict in the document's language, with the meaning this reference gives them.

| State | Meaning |
| --- | --- |
| `decided` | Its owner decided it; the source says who and when. |
| `assumption` | The writer completed it; it holds unless the owner says otherwise. |
| `to decide` | Nobody decided it yet; the line says who answers it. |
| `cancelled` | A later version replaced it; the line says which version cancelled it and the code of the rule that replaces it. |

| Verdict | Meaning | Effect |
| --- | --- | --- |
| `exists` | The code already does it. | None. |
| `new` | The code does not do it and nothing in the code prevents it. | None. |
| `contradicts` | The code does something else today. | A question for the owner. |
| `blocked` | The rule needs data, a relation or a permission that does not exist and that no rule of the document provides. | A question for the owner. |
| `not checked` | Nobody compared it with the code yet. | Holds back `ready for development`. |

Write a verdict in words its author can read, such as "today a supervisor is not asked", and keep code locations for the technical review section.

## Ready list

A version passes when all ten hold; each check that fails is reported with what is missing.

1. The nine sections are present, in order, and the non-goals name at least one thing.
2. Every rule line has its six parts.
3. Each rule states one behavior.
4. No vague term, such as "fast", "user-friendly", "as needed" or "etc.", and no "to be defined" marker stands in a rule or a criterion.
5. Every `P0` and `P1` rule has at least one acceptance criterion, and a person other than the author can observe each criterion.
6. No two rules contradict each other.
7. For each action a user story names, the rules say who may take it and what happens when it is refused or fails.
8. Every open question says who answers it.
9. No rule states a screen layout, a technical how or a date.
10. Every goal has a success metric, or an open question about it.

## States

The state of the document is set by what the checks support. A label in the source, such as "ready for development", is a claim to check and never sets the state.

| State | It is reached when |
| --- | --- |
| `draft` | The document is being written and its owner has not received it. |
| `iterating` | Its owner received it, and a check of the ready list fails, a rule is `to decide`, an open question is unanswered, or the review did not run. |
| `approved by business` | All four hold for this version: the ready list passes, a `hive-review-prd` child reviewed it, no rule is `to decide`, and the user approved it explicitly, as [create-prd](../SKILL.md) step 7 asks. |
| `technical review` | A session with the repository took a version that is `approved by business` and is comparing it with the code. |
| `ready for development` | The technical review set it, as [technical review](technical-review.md) sets. No other step sets this state. |

A change to any rule after an approval produces a new version, which starts in `iterating`; the approved version stays frozen as it was.

## Versions

Versions are numbered from 1. A new version starts each time the document changes after its owner received it, not at each edit while drafting. Each row of the version history holds the version, its date, the state it reached, who approved it, the link or path of its frozen copy when it has one, and its list of changes:

- **Added**: the codes of the new rules.
- **Decided**: each rule that left `to decide` or `assumption`, with the answer and who gave it.
- **Changed**: each rule whose wording changed without changing its behavior, with what changed.
- **Cancelled**: each cancelled rule, with the reason and the code of the rule that replaces it.

Never delete a rule's line. When an answer changes the behavior a rule states, cancel that rule and add the new behavior as a rule with a new code; keep the code only when the wording changes and the behavior does not.
