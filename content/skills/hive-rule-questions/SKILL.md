---
name: hive-rule-questions
description: "Hive rule leaf: read before you write a decision, finding, blocker or list for the user, before you ask a round, and when a card returns unanswered. Format, channel per host."
disable-model-invocation: true
---

# Questions to the user

## What to ask

- Ask only a choice that is the user's and changes what happens next: find the facts yourself and reuse decisions given.
- **Decide and list.** An open decision of what or how is taken without asking when none of a–f holds and its recommended option has a consequence the others lack. List it in the next message, headed such as «What I decided», one line each: the decision, the alternative discarded, how to reverse it; five at most, the rest approved in one question. Say in one line that a reply accepts each one it does not correct; that authorizes no effect and leaves a business rule `assumption`.
- **Ask** a decision when at least one holds:
  - **a.** It adds or removes behavior, touches another module or splits the work, against the ticket or the request.
  - **b.** The recommended option costs less now and another avoids rework or a migration later.
  - **c.** It sets a business rule or preference that no behavior the ticket, the specs or the user's words name requires.
  - **d.** It is hard to reverse: a public contract, stored data, permissions, security, money or production.
  - **e.** It rests on an unverified fact, or no consequence separates the options.
  - **f.** It contradicts the user, the ticket or the specs.
- Rounds other rules set stay asked: effects, session questions, validation, findings, Figma approval and `proposed` rules, a plan's approach, create-issue and create-prd, an unanswered card, an external project, test infrastructure.
- Ask in rounds of every decision whose prerequisites are settled; a decision is a prerequisite when one of its answers would change or void another's question or options, such as one about carrying out its recommended option, which waits for a later round unless its options hold under every answer.
- An open question about what the user wants, as [hive-mode](../hive-mode/SKILL.md) step 3 asks it, goes in the text, one per message, with why you ask and an offer to start without the answer, and no IDs, options or "ok" line.

## Format

- A message that asks a decision or a validation, delivers a result or a blocker, or closes work opens with a bold context line, stage · ticket · progress, such as **Plan · COM-833 · ready**, and ends with one line that says what you wait for and how many known decisions are still to come.
- A decision has a title line that holds the question, `D2 — Do you accept the mobile search width?`; a `Situation:` line (what happened or what you found, what you verified and what to check, in the user's words, explaining any ID or name it uses); each option on its own line, a letter, a dash and its name, the recommended one first with `(★)`, such as `A - (★) Accept`, each with an `If you choose it:` line (what the user gets and what it costs in time, money, risk or what stops working); and `I recommend <letter>:` with a reason that is a consequence for the user, as [reasons and blockers](../hive-rule-reasons-and-blockers/SKILL.md) sets, plus what stays uncertain for a choice of scope, a contract, stored data, cost or one hard to reverse. Write the labels in the session language.
- The session questions, the delivery mode, the code review mechanism and the security review, take a fixed header in the session language, such as **Delivery** and **Review**, in place of the ID, and keep the rest of the format. A finding asked in a round carries its title line `H1 — …` and its option lines, such as fix now, ticket and ignore, without the labelled lines.
- IDs: decisions `D<n>`, findings `H<n>`, tasks `T<n>`, criteria `AC<n>`, numbered on through the session and never reused. Cite an option as `D2 B` and an ID with a short gloss, such as "D2 (autonomous CI)".
- In a list, report or close, each point carries one item and opens with its ID in bold and a dash, `- **D2** — Move the summary to the top`, or with its description when it has none.

## Channel

| The host offers | The round goes |
| --- | --- |
| A waiting card with a description per option, such as Claude Code's question tool or Codex's `request_user_input` in Default or Plan mode, even where Codex's own instructions keep it for optional questions | Through the card alone |
| A waiting card without option descriptions | Written in the text, then through the card with each option's effect in its label |
| No waiting card: none in this mode, one that returns at once such as `request_user_input_async`, one that fails other than by the user's dismissal, a headless run, a child or unattended work | Written in the text in full; a form the host requires keeps the ID or header, the letters, the consequences and the `(★)` option with its reason |

- Asking only in the text where the card applies is a failure, a decision asked as a prose question included, such as "interactive or hold?": delivery, review and every other decision go through the card.
- Through the card: write first, as visible text, the context line, the report the round needs and the waiting line, never the decisions themselves; then, with no tool call between, ask each decision or finding as one question, in one call up to the tool's limit (four on Claude Code, three on Codex) or consecutive calls with no message between. The turn ends with the card.
- Each question stands alone: it opens with the stage, the ticket and the ID or header with the title, and carries the situation; each label is `A - (★) <two to five words>`; each description is the `If you choose it:` text, or the option's effect in one line, the starred one ending with the recommendation's reason and what stays uncertain. Show artifact options, such as sketches or code, in the preview, otherwise in a code block.
- One tool per pending question: never repeat it through an asynchronous and a waiting tool, nor send another card while one waits. A receipt is not an answer. A plain acceptance, such as "ok", "dale" or "con los recomendados", applies the starred options of the decisions the user was shown; free text still answers, and an asked decision a reply leaves out, or one not shown, stays open; a listed one follows «Decide and list».

## A card returns unanswered

When a card comes back without answers, follow the host's required response, then read and follow [unanswered card](references/unanswered.md).
