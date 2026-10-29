# Reporting and citations

## When to read

Read the applicable conditions and actions before writing a report of work or findings, a relay of a child's or tool's result, an explanation of how something works or why it failed, or a message that cites links or IDs defined outside it; format rules still apply to an explanation with no citation, and a greeting or a short direct answer needs no read. Reading this reference does not invoke the closing workflow. For a decision or manual validation request, also read [questions](../../hive-core/references/questions.md).

## Lead with the result

Lead with the answer or outcome; when several openings apply, open with work still running, then a completed work item, then the answer. Match detail to the question and the decision it supports, not the effort spent. Explain necessary context without assuming prior knowledge or belaboring the obvious. Write in the session language. Report the evidence that affects the conclusion, not each tool call.

## Open with the context line

**Condition:** a message to the user asks for a decision or a validation, delivers a stage's result or a blocker, or closes work.

**Actions:** make its first line the context line, in bold: `<stage> · <ticket(s)> · <progress>`. The stage, untranslated despite the Language rule, is the outermost running flow's name without `flow-`, such as Build, or another skill's, such as harness-audit; outside one, Research for an investigation and Build for a change; never invent one. The rest is in the session language. Name the ticket by its ID, or the change by its name when there is no ticket. The progress is the stage's own counter when it has one, such as tasks verified of the total or children returned; omit it when the stage has none. The context line is a heading, not the first sentence: the opening that [Lead with the result](#lead-with-the-result) sets, or a plan's pause block, comes after it, then the body, and the pause line stays last. A native question card, which a host may show alone, opens its question with the stage and ticket.

Greetings, a short answer outside that Condition, a message sent mid-work, such as the sentence that [a turn that ends to wait](#report-a-running-or-completed-item) carries, a child's return to its parent, and files take no context line.

## Citations and links

**Condition:** a message includes a link, including tickets, issues and pull requests.

**Actions:** write each link as a Markdown link with a short descriptive label rather than a bare URL, using known context rather than querying solely for a title. Link a ticket or issue only at its first mention in a message, with a URL the tracker returned or a verified link, trimmed to its shortest stable form, such as one without a title slug; never compose its workspace or domain. Without one, cite the ID without a link. Cite later mentions by ID alone, since some hosts print each link's full target.

**Condition:** citing an ID outside the message that defines it.

**Actions:** add a short gloss of what it is, every time, such as "R2 (split the PR)"; give a range one gloss that says what its items do, such as "S1–S6 (merge, CI and cleanup checks)"; "checks S1–S6" names only their kind and is not a gloss. This applies to message text, the close question, and relayed subagent results. A native question counts as its own message, because it shows only the choice: gloss every ID it cites that was defined outside it, including in the text just before it; option IDs it defines, such as `D1-A`, need none.

**Condition:** a message holds more than one numbered list.

**Actions:** give each list its own ID prefix, such as findings H1, H2 and recommendations R1, R2, and refer to items by those IDs, so a reply like "apply R1" or "drop H4" is unambiguous.

## Shape the message by its content

- Make multi-topic findings easy to scan: distinguish the outcome, material findings, pending decisions, and next action with short labeled sections.
- In reports, give each area a short bold label that reads on its own, in the session language, such as delivery, tickets, specs, cleanup, pending items, and recommendations, followed by its detail.
- Link the retained document for technical detail instead of reproducing it; alternatives, recommendations, and any finding the user must weigh stay in the message.
- When reporting a newly opened PR or one awaiting CI or review, include its link, and relevant run links when useful.
- In chat replies, use a table only to compare three or more items across two or more attributes, with at most four short columns; a list for parallel items; prose for causes and reasoning.
- Put each ticket on its own list line, or each set of tickets meant to be worked together, with a description the reader understands without opening the ticket; never put three or more tickets in one paragraph.
- A format that a skill or template prescribes takes precedence. State an item's status in words in the session language, such as done, in progress, blocked, or pending on the user, not emoji. Add no formatting that carries no structure.

**Condition:** explaining how something works, why it failed, or how parts relate, and the shape branches, fans out or in, crosses three or more layers, passes gated stages, or changes between a before and an after.

**Actions:** add a small ASCII diagram in a code block beside the prose, at most 70 characters wide; brevity rules never suppress it, since it replaces prose. A short linear sequence stays a numbered list. Keep Mermaid and SVG out of the chat, since several hosts show their source raw.

**Condition:** the user will share or keep a comparison, explainer, or long report.

**Actions:** offer a `flow-report` page as an option of the question the message ends with, such as a route or close question, or, when the message ends with none, ask it by following [questions](../../hive-core/references/questions.md).

## Report a running or completed item

**Condition:** a child's or a background process's notification starts the turn, not a message from the user.

**Actions, in order:**

1. When the result or the notice says the child stopped while it still waits on work of its own, it has stopped: apply [Check returned work](../../flow-research/references/delegation.md#check-returned-work) before writing anything.
2. When the notice reports the exit of a process you stopped on purpose, such as a server you stopped at cleanup or to restart it, send no message, also while a question waits, and go to step 6. The exit status that your stop caused, such as 143, is not a failure. If the host requires a reply, apply step 5 instead.
3. While a question waits for the user, a result that does not change what it asks takes step 5, even when it decides something else; a result that changes what it asks is a new question with a new ID that replaces the open one, asked by following [questions](../../hive-core/references/questions.md).
4. Otherwise, when the result decides something, brings a new blocker, or needs a new decision from the user, report it as the rest of this reference sets; while work of your own remains, send that report with your next tool call, as step 6 sets, and ask a decision it needs when [questions](../../hive-core/references/questions.md) says to ask it.
5. Otherwise the turn changes nothing the user needs: send no report and do not repeat an open question or its pause line; a turn that then ends to wait carries only the sentence the next condition sets, which is not a second report.
6. Continue with your next tool call when work of your own remains; otherwise end the turn.

**Condition:** a turn ends to wait on work still running, such as subagents, reviews, CI, or deployments, except a notification turn that step 2 above answers without a message.

**Actions:** say in one sentence what is still running and what its result decides; a report the turn carries opens with it.

**Condition:** a message delivers a stage's result: a validation request, a completion report, or a stop at a blocker.

**Actions:** state what changed; each check not met, accepted without verification or skipped, and the verification level when it is below the agreed one; any unresolved failure or uncertainty; each decision taken without asking that the user could reasonably have made differently, such as a behavior or contract chosen, a blocking defect fixed, or the ticket you chose when the user left it to you, with its reason and how to revert it; the first rendered view's artifact path when no validation stop showed it; and the next step when there is one. Never list the checks that passed, except those a mechanical change relies on and the per-environment results a deployment report sets.

**Condition:** a work item completes.

**Actions:** write its completion report as a chat message before calling the question tool. Open it by saying it is complete and what that means: the completion point reached, such as merged into the base branch at a named commit or pull request with the ticket moved to a named state, and what it does not include, such as promotion to later environments. Then give the pull-request URLs and what the previous condition lists. Add, only when there is one: cleanup that failed or left something behind, such as a running process or a kept branch or worktree, with why; and any secret that reached a durable or shared place or other incident the user must know. Name each open topic with its proposed destination, as the open-topics rule in the global guidance sets.

**Condition:** a pending item remains.

**Actions:** do not say the work is complete. Open instead with the point reached and what is still pending.

## Closing categories

Close findings and completion reports with the closing items that actually apply. No category is required: a report may have no pending item, next step, recommendation, or reminder, and none is invented to fill one. A follow-up that ends a piece of work still closes with the items that apply. Put each item in exactly one category, in this order. Translate labels to the session language.

### Pending

**Condition:** an obligation of the current work in this project is unmet, such as a check, publication, or agreed manual action.

**Actions:** list resulting user actions here. An open choice is not a Pending item: ask it by following [questions](../../hive-core/references/questions.md). The work is not complete while one remains. A build assumption awaiting confirmation at the close is neither an open choice nor a Pending item. Another project's items belong in its session. Start each item with its sole owner in the session language, user or you; make the user's request self-sufficient under [questions](../../hive-core/references/questions.md#make-user-dependent-work-self-sufficient), and state what you will do and when for yours. Before giving an item to the user, verify its premise and whether this session's access and authorization already let you do it; if they do, do it instead. An item stays the user's only for their decision or authorization, access you checked and lack, or an action off this machine; say what you checked.

### Next step

**Condition:** there is a single continuation of a route already agreed, such as running `flow-build` on a ready plan.

**Actions:** state who executes it and whether it needs new authorization; do not present it as a choice. While the user still has to choose, there is no next step: the choice belongs in Recommendations.

### Recommendations

**Condition:** several reasonable routes exist and your judgment helps choose, such as the next ticket, a scope change, or a fresh session with a concise handoff when little active context remains useful.

**Actions:** number them, give each its reason, and mark one Recommended when it is clearly better. When the user asked for one or what to take next, ask the routes as a choice, each option naming the flow it starts as [flow-research](../../flow-research/SKILL.md#deliver-the-finding) sets, iterative route included. Start the chosen flow without a new command.

The choice authorizes that flow only, or that iterative route's rounds, plus opening the ticket that a research route option names, as [flow-research](../../flow-research/SKILL.md#deliver-the-finding) sets. The flow asks for commits, pushes, pull requests, deployments, tracker writes or secret reads through its delivery question before its first edit, or reuses an earlier answer. When the option or the user's request names a delivery, such as a promotion to a named environment or a release, the answer covers what [effect authorization](../../hive-core/references/effect-authorization.md#what-an-answer-covers) sets.

### Reminders

**Condition:** a user control does not block the current work, such as delivered work not yet promoted, including to a later completion point as [project settings](../../hive-core/references/project-settings.md#later-completion-point) sets, since promotions batch several tickets and never hold the current work open, or a pull request awaiting the user's merge.

**Actions:** state it in one line without advice; repeat it only when the state changes or related work closes. Make it a recommendation only when waiting has a concrete cost the user may not see, such as a production defect or a security exposure, and name that cost.

Distinguish work already completed, work you will continue under existing authorization, and actions that need new authorization. Ask once for a set of actions when one answer can authorize them all, as [questions](../../hive-core/references/questions.md) sets for the steps of one objective, since each extra question adds a round-trip without a new decision; ask separately only for an action whose risk or target needs its own decision, such as a production deployment.

## Work-item boundary

For readers who reach this reference from any flow: before the first message of authorized work while work of your own remains, before ending a turn while it remains, and before ending a turn that completes a work item, read [work-item boundary](work-item-boundary.md): it sets when the turn may end.

## Worked example: report with pending work

Assume the local checks and the required review passed, publishing remains an agreed obligation the user already authorized, and the push fails because the remote rejects this machine's credentials, a blocker that stops all remaining work, so the turn ends on this report. The link below is a fictional template, not a bundled dependency.

```text
**Build · ARK-123 · 2 of 2 tasks verified**
The local change is verified and reviewed; publication is blocked.

**Delivery:** Updated the retry handling.
The [change record](<change-folder>/proposal.md) holds the technical detail.
**Pending:** You — restore push access to the remote for this machine;
I checked that the stored credential is rejected. Me — I will publish by
the already agreed route once the push succeeds.
```

This is not a completion claim and does not invent a choice or a reminder. If a new user decision were needed instead, ask it by following [questions](../../hive-core/references/questions.md); do not put that open choice under Pending or announce an unchosen next step.
