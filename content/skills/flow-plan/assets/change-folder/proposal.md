# <TICKET, or the change-id without a ticket> — <one-line title>

<!-- Template from flow-plan. Write the text, headings and field names in the session language; keep code and identifiers as they are. Name the first field `Status`, or `Estado` in Spanish, and start its value with the phase, since the archive helper reads it. Delete these comments. -->

| Field | Value |
| --- | --- |
| Status | <Draft · Ready for flow-build · In build · Completed> |
| Tracker | <ticket link · state · priority · related tickets, or none> |
| Git | <repository, base branch and commit, proposed branch> |

## Decisions

<!-- One line per decision the user took, under its date: the decision ID and letter in bold, a dash, the question in plain words, and the answer and what it changes, such as "- **D62 B** — Who links each event to its stage? The server; the agent no longer declares it.". Readers outside the conversation do not know the IDs: no line here or in any other section relies on an ID alone. -->

<date>
- **<D<n> letter>** — <question?> <answer and what it changes>.

## Request

<The user's request, quoted. The ticket's acceptance criteria, quoted, with the date they were read. What the ticket leaves out of scope.>

## Objective

<What happens today, and what happens when the change is done.>

## Scope and acceptance

<What is in scope and what is out of it.>

<!-- Format reminder, rule in flow-plan «Revise a plan in progress»: a revision marks the new decision or criterion with `**Replaces:** <old IDs>` on its line, and a dropped criterion keeps its line with `**Retired:** <reason>`. Both markers stay in English, whatever the session language, and so does `Observed in:`. Delete this comment. -->

- AC1 — <one observable behavior> · false today because <the observation that shows it> · Observed in: <where its real effect shows>
- AC2 — <…>

## Prerequisites

<!-- What the build needs from outside the repository, usually from a third party or a person: an API key, an account or user in a platform, an access grant, a test number, data. Never write a secret's value: name where it is stored. When there are none, keep only the next line. -->

None: the build needs nothing from outside the repository.

- [ ] <what is needed> · owner: <who provides it> · needed by: <T<n>>

## Verified context

<What was read and observed before planning, with paths, commits and dates.>

<!-- Only when the approach reads or changes an external system: one line per external fact, as flow-plan's reference «Outside the repository» sets. A fact no read-only read could show is written as not measured, with the task that meets it first. Delete this comment. -->

- <external fact> · <read-only command or source> · <result, or not measured> · <date>

## Effects

<!-- One line per class of effect asked in the plan's effects round. Keep this heading in English, whatever the session language, like `Observed in:`: flow-deploy looks for it. When the plan foresees no effect on an external system, keep only the `None:` line; otherwise delete the `None:` line and keep one line per class. Delete this comment. -->

None: the tasks take no effect on an external system.

- <class of effect> · target: <account, repository or environment> · ceiling: <amount, or 0> · tasks: <T<n>> · answer: <approved, refused or unanswered, with the decision ID and letter when answered>

## Delivery

<The delivery mode, its validation stop and its endpoint. The code review mechanism. What this choice does not authorize.>

## Review status and progress

<Date — the plan review and what it changed. Then one line per progress event.>
