# Questions and decision pauses

## When to read

Read before formulating a decision for the user, with or without a native card, including route, delivery, review, validation and close questions. This reference does not invoke planning or grant any effects. For any question or accompanying message containing citations, links or IDs defined outside it, also read [reporting and citations](../../flow-close/references/reporting.md#citations-and-links).

## Decide whether a question is needed

**Condition:** a choice genuinely belongs to the user and changes what you do next, such as scope, design, delivery, a route, a validation, or the close.

**Actions:** ask only at that decision point. Look up facts yourself. Make a minor call, one with a conventional default that is neither consequential nor an effect needing authorization, yourself and state it in one line. Never ask only for leave to proceed or to show something you could show.

**Condition:** formulating any question for the user, whether it stands alone or belongs to a round.

**Actions:** give each question an ID such as `D1`, the context needed to answer, and options with IDs such as `D1-A`, each with its effect, the Recommended option first with its reason; a yes-or-no question needs no option IDs. Never leave a choice's options only in a ticket or document. State as fact in a question and its options only what you verified in this session; mark the rest as an assumption to confirm. Reuse explicit decisions already in force; silence, preselection, and a dismissed or timed-out question grant nothing.

**Condition:** several open questions do not depend on one another.

**Actions:** ask all of them in one round.

**Condition:** a consequential choice changes scope, contracts, data, or cost, or is hard to reverse.

**Actions:** also state the main tradeoff and uncertainty.

## Use a waiting native card

**Condition:** the host offers a native question tool that waits for the answer.

**Actions, in order:**

1. Complete memory, tracker and plan updates that apply before the message.
2. Write the explanation, then one line per question with its ID, the question, and the Recommended option with its reason. Gloss that option's ID; omit the other options from this message.
3. End the message with the pause line below. Make no tool call between the message and the question call. A response holding only the call is the failure.
4. Ask through the native tool. Write the card to stand alone, since a host may hide the message: the question carries the context to answer, each option its effect in its description, or in its label when the tool has no description field, and the Recommended option is marked.

Split questions or options beyond the tool's limit across consecutive calls, each after a line of text. For a delivery choice with more modes than one card supports, use [delivery decisions](delivery-decisions.md)'s provisional-shortlist procedure: the first subset is not delivery authorization; show every mode and its effects before recording the final choice.

## Use text when no waiting card carries the question

**Condition:** no waiting card carries the question: no such tool exists; a tool returns before the answer, used only while independent work continues; or a dismissed or timed-out question is asked once more in text.

**Actions:** write each option in the text, one line each with only what sets it apart. State once what several options share and any route project policy rules out. Add no "Other" option or reply example. Take the first answer that arrives.

## Make user-dependent work self-sufficient

**Condition:** work waits on the user for review, validation, approval, manual action, or information only they hold.

**Actions:** state what is needed and what it unblocks; how to do or check it, with concrete targets (clickable URL, command, commit or diff, file path, account without secrets) and what to compare against; what you already verified; and the possible answers with what each triggers, in the card when one carries the question.

When the user must check several things, give each flow its own block, headed by the ticket and a short description, with bullet steps from the start to what they should see. A manual validation request, such as the walkthrough for the user to validate a result, correction rounds, or the interactive mode's stop before push, goes complete in the message and ends with the pause line.

## Pause and handle the answer

**Condition:** a message waits on the user's choice.

**Actions:** end it with one pause line, its last line, after any closing items: a fixed bold label in the session language meaning "waiting for your decision", such as `**Waiting for your decision:**`, then what is asked, by ID with a short gloss when it has one. Use the same label every time. When a native question follows, the pause line ends the message before it. If the turn ends with a question unanswered, never append a second report.

**Condition:** an open choice has already been asked.

**Actions:** ask each open choice once; only a dismissed or timed-out question is asked again.

**Condition:** an answer arrives, including free text.

**Actions:** say in one line before acting whether it settles the question, opens the next round, or returns the work to an earlier flow, such as from plan to research or from build to plan, and what it invalidates; a retained plan the answer changes follows [Revise a plan in progress](../SKILL.md#revise-a-plan-in-progress). Re-ask an answered question only when the change needs effects not yet authorized.

## Worked example: waiting card

Assume this session verified that the required reviewer cannot run, on its configured model or on the implementer's. The task remains unverified. The next call is the waiting native question; no other tool intervenes.

```text
The required reviewer could not access its model, so the task is still
unverified. Retrying preserves the agreed independent check.
D1: How should I handle this check? D1-A (wait and retry) is Recommended
because it preserves the agreed verification.
**Waiting for your decision:** D1 (the unavailable verification).
```

The card carries all options and repeats enough context to stand alone:

```text
Question: D1 (the unavailable verification): the required reviewer could
not access its model and the task is unverified. How should I proceed?
D1-A — Wait and retry (Recommended)
Effect: Preserve the agreed independent check and retry when access returns.
D1-B — Use a model you name
Effect: Run the check with the replacement model you specify.
D1-C — Accept unverified
Effect: Record your acceptance without an independent verdict.
```

After a free-text answer naming a replacement: "Your answer settles D1 (the unavailable verification); I will use the model you named, replacing the previous reviewer-model selection."

## Worked example: no waiting card

Assume the same verified failure, and no waiting card is available. Include all options in the message itself, not a second card:

```text
The required reviewer could not access its model, so the task is still
unverified. D1: How should I handle this check?
- D1-A — Wait and retry (Recommended): preserve the agreed check.
- D1-B — Use a model you name: run the check with your replacement.
- D1-C — Accept unverified: record acceptance without an independent verdict.
None of these routes publishes the change.
**Waiting for your decision:** D1 (the unavailable verification).
```
