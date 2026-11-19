---
name: hive-rule-questions
description: "Hive rule leaf: read before you write a decision, finding, blocker or list for the user, before you ask a round, and when a card returns unanswered. Format, channel per host."
disable-model-invocation: true
---

# Questions to the user

## What to ask

- Ask only a choice that is the user's and changes what happens next: find the facts yourself, state minor calls as assumptions (a minor call leaves scope, contracts, stored data, cost and reversibility unchanged under every option), and reuse decisions already given. Ask in rounds of every decision whose prerequisites are settled; a decision is a prerequisite when one of its answers would change or void another's question or options, such as one about carrying out its recommended option, which waits for a later round unless its options hold under every answer.
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
- Each question stands alone: it opens with the stage, the ticket and the ID or header with the title, and carries the situation; each label is `A - (★) <two to five words>`; each description is the `If you choose it:` text, or the option's effect in one line, the starred one ending with the recommendation's reason and what stays uncertain. Show artifact options in the preview or a code block.
- One tool per pending question: never repeat it through an asynchronous and a waiting tool, nor send another card while one waits. A receipt is not an answer. A plain acceptance, such as "ok", "dale" or "con los recomendados", applies the starred options of the decisions the user was shown; free text still answers, and a decision a reply leaves out, or one not shown, stays open.

## A card returns unanswered

Follow the host's required response first. The answers the card did return stand; ask none of the round's remaining cards first.

- **A card that carried the round alone closed by itself:** an empty answer or a timeout with no rejection, as a Codex card after about 120 seconds. Write in one message every open decision and finding of the round in full, those not yet asked included, ending with a waiting line saying "ok" applies every starred option and a reply such as `D2 B` changes one.
- **A card after a round written in the text** comes back empty: add one line that the card closed and names the decisions still open.
- **The user left it:** a rejection, dismissal or decline, also as an error or after the host's own timeout; the host ended the turn with the card open; a request to talk; answers to only some questions; or free text that adds, removes or replaces a decision or an option. Never write the round out on your own; act on what they said, first fit:
  1. Answers by ID or a plain acceptance: apply them and ask the rest through the card again.
  2. The cancel was a mistake, or they ask for the questions: ask the same round again, same IDs.
  3. The direction or the premise changed, or a decision or an option was added, removed or replaced: rework those decisions and ask them as a new round with new IDs.
  4. Drop the questions: drop them, except one a later step cannot do without, asked again with that reason.
  5. Another subject or question: answer it and end with one line naming the decisions still open and how to get them back; ask them through the card when the user says so or asks to continue work that needs them.

  With nothing said about the decisions yet, reply with that one line alone. A decision asked again keeps its ID, after one line naming which return. Later rounds still go through the card.
