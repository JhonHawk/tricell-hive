# Ticket types

Read before drafting a ticket and before checking one, at [create-issue](../SKILL.md) steps 5 and 6, and when another flow judges whether a ticket is ready, as [flow-research](../../flow-research/SKILL.md) does for a chosen ticket. Reading this reference creates no ticket.

## Sections

Every ticket has these sections, with their names written in the ticket's language:

- **Title**: the outcome in one line, without implementation detail.
- **Outcome**: what changes, for whom and why, in two to four sentences.
- **Done when**: the acceptance criteria, at most 7, each one rule that a person other than the author can observe and check.
- **Out of scope**: at least one thing a reader could expect from the title and will not get.
- **Open questions and assumptions**: each open question with who answers it, and each assumption made while writing.
- **Dependencies**: the tickets it waits on or blocks, each with its link, or none.
- **Hint, not binding**: optional, as «Hint» below sets.

Pick one type and add its sections:

| Type | Use it when | It adds |
| --- | --- | --- |
| Story | The outcome is a capability a user gets | **For whom and why**: who needs it and what it lets them do; leave out the why when it adds nothing to the outcome |
| Bug | Behavior contradicts what was specified or what worked before | **Steps to reproduce**, **Expected and observed**, **Environment and version**, and **Evidence**: the error text, log lines or capture that show the failure, never a secret |
| Task | The outcome is a change no user sees as a capability | **The change**, and **Verify with**: the command or observation and its expected result. Technical debt is a task that also adds **What it costs today** |
| Spike | The outcome is knowledge a decision needs | **Question**, **Timebox**, **Deliverable** (a decision or a write-up) and **Decision it feeds**; its Done when is the deliverable |

What the ticket needs from a document it links, such as a requirement or a design, is summarized inside the ticket in at most three sentences, next to the link: a ticket that only links leaves its reader to guess which part applies.

## Ready list

A ticket is ready when all eight hold; otherwise it is not ready. A ticket that fails one because it lacks something nobody in the session can supply is blocked.

1. The title and one sentence of the outcome name the change without implementation detail.
2. At least one criterion exists, and each one is observable: a person other than the author can check it.
3. Each criterion states one rule.
4. No vague term, such as "fast", "user-friendly" or "as needed", and no "to be defined" marker stands in a criterion.
5. An out-of-scope line exists.
6. Every open business question is listed with who answers it, and none is unresolved.
7. The sections its type adds are present: for a bug, its steps to reproduce, its expected and observed results and its environment.
8. The work is one change that is integrated and accepted on its own, with at most 7 criteria; a spike has its timebox and its deliverable.

A draft that fails check 8 for its size goes to [breakdown](breakdown.md). A ticket that already exists and fails it is reported as too large, with the split proposed and nothing split: [flow-research](../../flow-research/SKILL.md) settles its scope with the user, and [existing ticket](existing-ticket.md) proposes the split to its owner.

## Hint

A ticket carries at most one hint, in its own section titled as not binding, written only with the code at hand: up to five locations, each a file, module, screen or endpoint, and one sentence of direction. In a session without the code, write no hint and say so among the assumptions.

A ticket never carries a list of implementation steps, a choice between approaches, a contract definition such as an API shape, a schema or an event, or a file-by-file change list: those are decisions of the plan. When writing the hint would need one of them, write no hint and state in the ticket, in its language, that it requires a plan.

Whoever works the ticket treats its hint as a claim to verify, never as a decision.
