# Reporting and citations

## When to read

Read the applicable conditions and actions before composing a chat reply: findings, progress or completion reports, ordinary answers and explanations, questions, native cards and relayed child/tool notifications. This includes messages with citations, links or IDs defined outside them and replies without them; format rules still apply to an explanation with no citation. Reading this reference does not invoke the closing workflow. For a decision or manual validation request, also read [questions](../../flow-plan/references/questions.md).

## Lead with the result

Lead with the answer or outcome; when several openings apply, open with work still running, then a completed work item, then the answer. Match detail to the question and the decision it supports, not the effort spent. Explain necessary context without assuming prior knowledge or belaboring the obvious. Write plain, natural prose in the session language for a reader who did not see the work: complete sentences, everyday words, and concrete names instead of shorthand, metaphors, or internal labels; explain a technical term the first time the reader may not know it. The English terms these instructions use, such as gate, walk, in-vivo, fixture, or cutover, are not the user's vocabulary: say what they mean in everyday words of the session language. Report the evidence that affects the conclusion, not each tool call.

## Citations and links

**Condition:** a message includes a link, including tickets, issues and pull requests.

**Actions:** write each link as a Markdown link with a short descriptive label rather than a bare URL, using known context rather than querying solely for a title. Link a ticket or issue only at its first mention in a message, with a URL the tracker returned or a verified link, trimmed to its shortest stable form, such as one without a title slug; never compose its workspace or domain. Without one, cite the ID without a link. Cite later mentions by ID alone, since some hosts print each link's full target.

**Condition:** citing an ID outside the message that defines it.

**Actions:** add a short gloss of what it is, every time, such as "R2 (split the PR)"; give a range one gloss that says what its items do, such as "S1–S6 (merge, CI and cleanup checks)"; "checks S1–S6" names only their kind and is not a gloss. This applies to message text, the close question, and relayed subagent results. A native question counts as its own message, because it shows only the choice: gloss every ID it cites that was defined outside it, including in the text just before it; option IDs it defines, such as `D1-A`, need none.

**Condition:** a message holds more than one numbered list.

**Actions:** give each list its own ID prefix, such as findings H1, H2 and recommendations R1, R2, and refer to items by those IDs, so a reply like "apply R1" or "drop H4" is unambiguous.

## Shape the message by its content

- Make multi-topic findings easy to scan: distinguish the outcome, material findings, pending decisions, and next action with short labeled sections.
- In progress and completion reports, give each area a short bold label that reads on its own, in the session language, such as delivery, tickets, specs, cleanup, pending items, and recommendations, followed by its detail.
- Link the retained document for technical detail instead of reproducing it; alternatives, recommendations, and any finding the user must weigh stay in the message.
- When reporting a newly opened PR or one awaiting CI or review, include its link, and relevant run links when useful.
- In chat replies, use a table only to compare three or more items across two or more attributes, with at most four short columns; a list for parallel items; prose for causes and reasoning.
- Put each ticket on its own list line, or each set of tickets meant to be worked together, with a description the reader understands without opening the ticket; never put three or more tickets in one paragraph.
- A format that a skill or template prescribes takes precedence. State an item's status in words in the session language, such as done, in progress, blocked, or pending on the user, not emoji. Add no formatting that carries no structure.

**Condition:** explaining how something works, why it failed, or how parts relate, and the shape branches, fans out or in, crosses three or more layers, passes gated stages, or changes between a before and an after.

**Actions:** add a small ASCII diagram in a code block beside the prose, at most 70 characters wide; brevity rules never suppress it, since it replaces prose. A short linear sequence stays a numbered list. Keep Mermaid and SVG out of the chat, since several hosts show their source raw.

**Condition:** the user will share or keep a comparison, explainer, or long report.

**Actions:** offer a `flow-report` page in one line.

## Report a running or completed item

**Condition:** a turn ends while work is still running, such as subagents, reviews, CI, or deployments.

**Actions:** open with a labeled in-progress item stating what is running, what its result decides, and what happens next, before any findings.

**Condition:** a work item completes.

**Actions:** write its completion report as a chat message before calling the question tool. Open it by saying it is complete and what that means: the completion point reached, such as merged into the base branch at a named commit or pull request with the ticket moved to a named state, and what it does not include, such as promotion to later environments. Then give what was delivered with its pull-request URLs, the cleanup, and any secret that reached a durable or shared place or other incident the user must know.

**Condition:** a pending item remains.

**Actions:** do not say the work is complete. Open instead with the point reached and what is still pending.

## Closing categories

Close findings, progress, and completion reports with the closing items that actually apply. No category is required: a report may have no pending item, next step, recommendation, or reminder, and none is invented to fill one. A follow-up that ends a piece of work still closes with the items that apply. Put each item in exactly one category, in this order. Translate labels to the session language.

### Pending

**Condition:** an obligation of the current work in this project is unmet, such as a check, publication, or agreed manual action.

**Actions:** list resulting user actions here. An open choice is not a Pending item: it goes in the pause line. The work is not complete while one remains; another project's items belong in its session. Start each item with its sole owner in the session language, user or you; make the user's request self-sufficient under [questions](../../flow-plan/references/questions.md#make-user-dependent-work-self-sufficient), and state what you will do and when for yours. Before giving an item to the user, verify its premise and whether this session's access and authorization already let you do it; if they do, do it instead. An item stays the user's only for their decision or authorization, access you checked and lack, or an action off this machine; say what you checked.

### Next step

**Condition:** there is a single continuation of a route already agreed, such as running `flow-build` on a ready plan.

**Actions:** state who executes it and whether it needs new authorization; do not present it as a choice. While the user still has to choose, there is no next step: the choice belongs in Recommendations.

### Recommendations

**Condition:** several reasonable routes exist and your judgment helps choose, such as the next ticket, a scope change, or a fresh session with a concise handoff when little active context remains useful.

**Actions:** number them, give each its reason, and mark one Recommended when it is clearly better. When the user asked for one or what to take next, ask the routes as a choice, each option naming the flow it starts: `flow-plan` when the route spans several tickets, repositories, or steps, `flow-build` only for a small understood change, and both when either fits, except the iterative route that [flow-research](../../flow-research/SKILL.md) defines. Start the chosen flow without a new command.

The choice authorizes that flow only, or that iterative route's rounds: never bundle commits, pushes, pull requests, deployments, tracker writes, or secret reads into a route option. The flow asks for them through its delivery question before its first edit, unless the user already requested them, such as a promotion to a named environment. Investigation findings use `flow-research`'s route question.

### Reminders

**Condition:** a user control does not block the current work, such as delivered work not yet promoted, even to a completion point the project declares after integration, since promotions batch several tickets and never hold the current work open, or a pull request awaiting the user's merge.

**Actions:** state it in one line without advice; repeat it only when the state changes or related work closes. Make it a recommendation only when waiting has a concrete cost the user may not see, such as a production defect or a security exposure, and name that cost.

Distinguish work already completed, work you will continue under existing authorization, and actions that need new authorization. Ask once for a set of actions when one answer can authorize them all, since each extra question adds a round-trip without a new decision; ask separately only for an action whose risk or target needs its own decision, such as a production deployment.

## Work-item boundary

A work item is a ticket or the scope the user authorized, such as a plan, one of its phases, or an iterative route; a task inside it is not one. Between tasks, report progress in a line and continue. While authorized work remains, send each progress line with your next tool call, since a response without one can end the turn; answer a question or notification that arrives meanwhile in one line and continue.

End a turn that completes a work item with a close question, after the report and its closing items; a turn that ends on the report is incomplete. Read [the close-question procedure](../SKILL.md#close-question) for its options, ticket selection and exceptions. Investigation-only findings and follow-ups use [flow-research](../../flow-research/SKILL.md#deliver-the-finding)'s route question instead.

## Worked example: report with pending work

Assume the local checks passed, publishing remains an agreed obligation, and the user already authorized the route. The link below is a fictional template, not a bundled dependency.

```text
The local change is verified; publication is still pending.

**Delivery:** Updated the retry handling; the targeted tests passed.
The [change record](<change-folder>/proposal.md) holds the technical detail.
**Cleanup:** Stopped the test process; kept the uncommitted source changes.
**Pending:** Me — I will publish by the already agreed route after the
required reviewer returns a passing verdict.
**Next step:** I will run the already authorized review on this candidate.
```

This is not a completion claim and does not invent a choice or a reminder. If a new user decision were needed instead, name it through the question procedure and end with its pause line; do not put that open choice under Pending or announce an unchosen next step.
