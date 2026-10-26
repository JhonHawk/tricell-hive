# Questions and decision pauses

## When to read

Read before formulating a decision for the user, with or without a native card, including route, delivery, review, validation and close questions, and before ending a message with an offer or request that waits on the user's answer. This reference does not invoke planning or grant any effects. For any question or accompanying message containing citations, links or IDs defined outside it, also read [reporting and citations](../../flow-close/references/reporting.md#citations-and-links).

## Decide whether a question is needed

**Condition:** a choice genuinely belongs to the user and changes what you do next, such as scope, design, delivery, a route, a validation, or the close.

**Actions:** ask only at that decision point. Look up facts yourself. Make a minor call, one with a conventional default that is neither consequential nor an effect needing authorization, yourself and state it in one line.

**Condition:** about to ask a question.

**Actions:** first check whether an answer in force already settles it: an earlier answer in this session, the effects an authorization already covers, or a route, scope or design decision the user made for this work item and recorded in its plan or in persistent memory from an earlier session. When one does, act on it and say in one line which answer covers it. An authorization of an effect carries over from an earlier session only as [effect authorization](effect-authorization.md#what-an-answer-covers) sets.

**Condition:** a message would end with an offer or request that waits on the user's answer, in any form, such as "should I review it?", "if you want, I'll close it", or "let me know if you want the report".

**Actions:** when the action is already authorized, including the checks the global guidance says not to ask about separately, such as reading or showing a pull request's diff or running the affected tests, do it instead of offering it. When it is a choice that belongs to the user under this section's first condition, or a question another rule requires, such as the route or close question, ask it under this procedure, never as a closing sentence alone. Otherwise, such as further work the answer does not call for, leave the offer out.

**Condition:** formulating any question for the user, whether it stands alone or belongs to a round.

**Actions:** give each question an ID such as `D1`, the context needed to answer, and options with IDs such as `D1-A`, each with its effect, the Recommended option first with its reason. A yes-or-no question, such as an authorization, is formulated the same way: its two options carry IDs and effects, such as `D5-A — run Bugbot on the pull request` and `D5-B — no dedicated review`, never a bare "do you authorize it?". Never leave a choice's options only in a ticket, a document, or a question tool that returns before the answer. State as fact in a question and its options only what you verified in this session; mark the rest as an assumption to confirm. Silence, preselection, and a dismissed or timed-out question grant nothing, except the silence after the route answer that activates [unattended delegation](../../unattended-delegation/SKILL.md) under its own rule.

**Condition:** several open questions do not depend on one another.

**Actions:** ask all of them in one round.

**Condition:** a question's options, context, or relevance depend on the answer to another open question, such as one that changes the course of an investigation, opens a different set of questions, or assumes that an option of the other, typically the Recommended one, is chosen.

**Actions:** ask the other question first and the dependent one in a later turn, after that answer, worded for the answer given. In one round, put only questions whose options hold with any answer to the others; when a question needs only its wording adjusted, write its options to hold with any answer instead, such as "with the review you choose in D2".

**Condition:** an objective the user already decided needs several authorizations in sequence, such as the records of one test setup or the steps of an operation the user approved.

**Actions:** ask them as one question that lists the whole set, even when a later step depends on an earlier one's result; ask again only when a result changes the set, naming what changed. Effects that [effect authorization](effect-authorization.md) asks apart, such as each kind of production effect or a critical effect, and a production deployment the user has not already named keep their own question in that same round.

**Condition:** a consequential choice changes scope, contracts, data, or cost, or is hard to reverse.

**Actions:** also state the main tradeoff and uncertainty.

**Condition:** an option of the question starts work on another ticket, adds a component, a language or a dependency, or is, as far as you can tell without estimating, at least ten times the size of another option of the same question.

**Actions:** give every option of that question that starts work, not one such as stop, its estimated size: the number of tasks and a time range, with the basis for it, such as the ticket's criteria or the files you read. When you did not read enough to estimate it, write "size not estimated: not investigated" for that option. Never give a number without a basis.

## Use a waiting native card

**Condition:** the host offers a native question tool that waits for the answer, and the question is not the round that [unattended delegation](../../unattended-delegation/SKILL.md#pre-authorization-round) asks at activation.

**Actions, in order:**

1. Complete memory, tracker and plan updates that apply before the message.
2. Write the message as visible text in the response, never only in your reasoning: a brief output style shortens it but does not remove it, because it is the result that style asks you to lead with. It may be short: the explanation, in one or two sentences when they suffice, then one line per question with its ID, the question, and the Recommended option with its reason. Gloss that option's ID; omit the other options from this message.
3. End the message with the pause line below. Make no tool call between the message and the question call. A response holding only the call is the failure.
4. Ask through the native tool. Write the card to stand alone, since a host may hide the message: the question opens with the stage and ticket that [reporting](../../flow-close/references/reporting.md#open-with-the-context-line) sets and carries the context to answer, each option its effect in its description, or in its label when the tool has no description field, and the Recommended option is marked.

Split questions or options beyond the tool's limit across consecutive calls, each after a line of text. For a delivery choice with more modes than one card supports, use [delivery decisions](delivery-decisions.md)'s provisional-shortlist procedure: the first subset is not delivery authorization; show every mode and its effects before recording the final choice.

## Use text when no waiting card carries the question

**Condition:** no waiting card carries the question: no such tool exists; the only tool returns before the answer, such as an asynchronous card that closes when the turn ends; or a dismissed or timed-out question is asked once more in text; or the round that [unattended delegation](../../unattended-delegation/SKILL.md#pre-authorization-round) asks at activation, whose pause line adds that the work continues meanwhile without commit or publication.

**Actions:** write each option in the text, one line each with only what sets it apart. State once what several options share and any route project policy rules out. Add no "Other" option or reply example. A tool that returns before the answer may be called in addition, only while independent work continues in the turn, and never replaces the options in the text: the waiting-card procedure above, which omits them from the message, does not apply to it. Take the first answer that arrives.

## Make user-dependent work self-sufficient

**Condition:** work waits on the user for review, validation, approval, manual action, or information only they hold.

**Actions:** state what is needed and what it unblocks; how to do or check it, with concrete targets (clickable URL, command, commit or diff, file path, account without secrets) and what to compare against; what you already verified; and the possible answers with what each triggers, in the card when one carries the question.

When the user must check several things, give each flow its own block, headed by the ticket and a short description, with bullet steps from the start to what they should see. A manual validation request, such as the walkthrough for the user to validate a result, correction rounds, or the interactive mode's stop before push, goes complete in the message and ends with the pause line.

## Pause and handle the answer

**Condition:** a message waits on the user's choice.

**Actions:** end it with one pause line, its last line, after any closing items: a fixed bold label in the session language meaning "waiting for your decision", such as `**Waiting for your decision:**`, then what is asked, by ID with a short gloss when it has one. Use the same label every time. When a native question follows, the pause line ends the message before it. If the turn ends with a question unanswered, never append a second report. A notification turn that [reporting](../../flow-close/references/reporting.md#report-a-running-or-completed-item) answers in one line while the question waits is not a second report and carries no pause line.

**Condition:** an open choice has already been asked.

**Actions:** ask each open choice once; only a dismissed or timed-out question is asked again. When a later result changes what is asked, ask a new question with a new ID that replaces the open one.

**Condition:** an answer arrives, including free text.

**Actions:** say in one line before acting whether it settles the question, opens the next round, or returns the work to an earlier flow, such as from plan to research or from build to plan, and what it invalidates; a retained plan the answer changes follows [Revise a plan in progress](plan-revision.md). Re-ask an answered question only when the change needs effects not yet authorized.

When the answer leaves a selection to your judgment, such as which findings to fix now and which to keep under observation, or which items to include, list the resulting selection before acting on it: what goes in and what stays out, each item with its gloss. Then ask whether the user agrees, recommending to continue as listed and taking corrections in free text. This confirmation is a new question with its own ID, not a re-ask of the answered one; when another question follows anyway, put the list and the confirmation in that round. An answer that picks one of the offered options leaves no selection to confirm. Two exceptions: a selection within the bounds of an iterative route the user chose goes in that round's progress line, and under [unattended delegation](../../unattended-delegation/SKILL.md) the selection is recorded with the decisions taken instead of asked.

## Worked example: waiting card

Assume this session verified that the required reviewer cannot run, on its configured model or on the implementer's. The task remains unverified. The next call is the waiting native question; no other tool intervenes.

```text
**Build · ARK-123 · 2 of 3 tasks verified**
The required reviewer could not access its model, so the task is still
unverified. Retrying preserves the agreed independent check.
D1: How should I handle this check? D1-A (wait and retry) is Recommended
because it preserves the agreed verification.
**Waiting for your decision:** D1 (the unavailable verification).
```

The card carries all options and repeats enough context to stand alone:

```text
Question: Build · ARK-123 — D1 (the unavailable verification): the required reviewer could
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
**Build · ARK-123 · 2 of 3 tasks verified**
The required reviewer could not access its model, so the task is still
unverified. D1: How should I handle this check?
- D1-A — Wait and retry (Recommended): preserve the agreed check.
- D1-B — Use a model you name: run the check with your replacement.
- D1-C — Accept unverified: record acceptance without an independent verdict.
None of these routes publishes the change.
**Waiting for your decision:** D1 (the unavailable verification).
```
