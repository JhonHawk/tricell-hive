# Questions to the user

Read before you write a decision, finding, blocker or list for the user, before you ask a round, and when a card returns unanswered.

## What to ask

- Ask only a choice that is the user's and changes what happens next: find the facts yourself and reuse decisions given.
- **Decide and list.** An open decision of what or how is taken without asking when none of a–f holds and its recommended option has a consequence the others lack. List it in the next message, headed such as «What I decided», one line each: the decision, the alternative discarded, how to reverse it; five at most, the rest approved in one question. A reply accepts each one it does not correct, which authorizes no effect and leaves a business rule `assumption`; say so in one line.
- **Ask** a decision when at least one holds:
  - **a.** It adds or removes behavior, touches another module or splits the work, against the ticket or the request.
  - **b.** The recommended option costs less now and another avoids rework or a migration later.
  - **c.** It sets a business rule or preference that no behavior the ticket, the specs or the user's words name requires.
  - **d.** It is hard to reverse: a public contract, stored data, permissions, security, money or production.
  - **e.** It rests on an unverified fact, or no consequence separates the options.
  - **f.** It contradicts the user, the ticket or the specs.
- Rounds other rules set stay asked: effects, the [gate questions](questions/gate.md) (a review mechanism the project names is stated instead, as [delivery questions](../git/delivery-ask.md) sets), validation, findings, Figma approval and `proposed` rules, a plan's approach, create-issue and create-prd, an unanswered card, an external project, test infrastructure.
- Ask in one round every decision whose prerequisites are settled. A decision is a prerequisite when one of its answers would change or void another's question or options; the dependent one waits for a later round unless its options hold under every answer.
- An open question about what the user wants, as [hive-mode](../../hive-mode/SKILL.md) step 3 asks it, goes in the text, one per message, with why you ask and an offer to start without the answer, and no IDs, options or "ok" line.

## Format

A round in the text, a blank line between its lines, labels in the session language:

```
**<Stage> · <ticket> · <progress>**

<the report the round needs>

❓ **D<n>** - **<the question?>**

Situation: <what happened or you found, what you verified, what to check>

A - (★) **<two to five words>**: <what the user gets, and its cost in time, money, risk or what stops working>

B - **<two to five words>**: <…>

➡️ A: <a consequence for the user>

---

❓ **D<n+1>** - …

<what you wait for · how many known decisions are still to come>
```

- The context line opens every message that asks a decision or a validation, delivers a result or a blocker, or closes work; the waiting line ends it.
- `Situation:` uses the user's words and explains each ID or name it uses. The `➡️` line gives its reason as [reasons and blockers](reasons-and-blockers.md) sets, plus what stays uncertain for a choice of scope, a contract, stored data, cost or one hard to reverse.
- A finding asked in a round carries its title line `❓ **H1** - **…**` and option lines, such as fix now, ticket and ignore, without `Situation:` or `➡️`.
- IDs: decisions and blocks of lettered replies `D<n>` ([report endings](report-endings.md)), findings `H<n>`, tasks `T<n>`, criteria `AC<n>`, numbered on through the session and never reused. Cite an option as `D2 B` and an ID with a short gloss, such as "D2 (autonomous CI)".
- In a list, report or close, each point carries one item and opens with its ID in bold and a dash, `- **D2** — Move the summary to the top`.

## Channel

Ask a round through the host's question tool when it waits for the answer, such as Claude Code's question tool or Codex's `request_user_input`; otherwise write it in the text in full, as «Format» sets. The round appears once per turn, in the card or in the text, since a repeated round makes the user read and answer it twice; the one exception is a card that closed by itself, as [unanswered card](questions/unanswered.md) sets.

- A choice asked as a prose question or reply offer, such as "interactive or hold?" or "reply «automatic» and I start", is a failure on every host and in every mode: a decision is asked in the round's format, through the card or in the text. A review mechanism stated as [delivery questions](../git/delivery-ask.md) sets is not asked.
- A waiting card without option descriptions carries each option's effect in its label. With no waiting card, the round goes in the text in full; a form the host requires keeps the ID or header, the letters, the effects and the `(★)` option with its reason. There is no waiting card when:
  - none exists in this mode;
  - one returns at once, such as `request_user_input_async` in a mode with no `request_user_input`;
  - one fails other than by the user's dismissal;
  - the run is headless, a child or unattended work.
- Through the card: write first, as visible text, the context line, the report the round needs and the waiting line, never the decisions themselves; then, with no tool call between, ask each decision or finding as one question, in one call up to the tool's limit (four on Claude Code, three on Codex) or consecutive calls with no message between. The turn ends with the card. Each question stands alone:

```
question:    <Stage> · <ticket> · ❓ D<n> - <the question?> Situation: <…>
label:       A - (★) <two to five words>[: <its effect>, on a card without descriptions]
description: > <what the user gets, and its cost>; the starred one ends with ➡️ <the reason> and what stays uncertain
label:       B - <two to five words>
description: > <…>
```

- One tool per pending question: never repeat it through an asynchronous and a waiting tool, nor send another card while one waits. A receipt is not an answer. A plain acceptance, such as "ok" or "dale", applies the starred options of the decisions the user was shown; free text still answers, a decision a reply leaves out or one not shown stays open, and a listed one follows «Decide and list».

## A card returns unanswered

When a card comes back without answers, follow the host's required response, then read and follow [unanswered card](questions/unanswered.md).
