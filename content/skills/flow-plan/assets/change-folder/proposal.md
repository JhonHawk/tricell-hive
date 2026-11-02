# <TICKET, or the change-id without a ticket> — <one-line title>

<!-- Template from flow-plan. Write the text, headings and field names in the session language; keep code and identifiers as they are. Name the first field `Status`, or `Estado` in Spanish, and start its value with the phase, since the archive helper reads it. Delete these comments. -->

| Field | Value |
| --- | --- |
| Status | <Draft · Ready for flow-build · In build · Completed> |
| Tracker | <ticket link · state · priority · related tickets, or none> |
| Git | <repository, base branch and commit, proposed branch> |

## Decisions

<!-- One line per decision the user took, under its date: the question in plain words, the answer and what it changes, and the decision ID last, in parentheses, such as "- Who links each event to its stage? The server; the agent no longer declares it. (D62 B)". Readers outside the conversation do not know the IDs: no line here or in any other section relies on an ID alone. -->

<date>
- <question?> <answer and what it changes>. (<D<n> letter>)

## Request

<The user's request, quoted. The ticket's acceptance criteria, quoted, with the date they were read. What the ticket leaves out of scope.>

## Objective

<What happens today, and what happens when the change is done.>

## Scope and acceptance

<What is in scope and what is out of it.>

- AC1 — <one observable behavior> · false today because <the observation that shows it>
- AC2 — <…>

## Prerequisites

<!-- What the build needs from outside the repository, usually from a third party or a person: an API key, an account or user in a platform, an access grant, a test number, data. Never write a secret's value: name where it is stored. When there are none, keep only the next line. -->

None: the build needs nothing from outside the repository.

- [ ] <what is needed> · owner: <who provides it> · needed by: <T<n>>

## Verified context

<What was read and observed before planning, with paths, commits and dates.>

## Delivery

<The delivery mode, its validation stop and its endpoint. The code review mechanism. What this choice does not authorize.>

## Review status and progress

<Date — the plan review and what it changed. Then one line per progress event.>
