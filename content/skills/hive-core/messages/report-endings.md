# How a message ends

Read before ending a message that reports, hands over or closes, or offering a single action. The action is one that changes no repository file.

Each row below ends a message differently. Every message that closes work ends with the waiting line, whichever row it takes.

| Way of ending | It ends with |
| --- | --- |
| Analysis, explanation, review or diagnosis the user asked for without asking for a change | Its findings, then [the reply block](#the-reply-block); it recommends no flow the user did not ask for, and when the findings confirm a defect in the current project, one reply fixes it. |
| Assessment of a ticket the user picked for research | When the findings confirm work the ticket asks for, [the next-step block](#the-next-step-block) toward `flow-plan`, or `flow-build` for a small understood change ([proportionality](../plan/proportionality.md)), unless the design stage of [hive-mode](../../hive-mode/SKILL.md) sends it to its own block; otherwise as the first row. |
| Hand-over to another flow | [The next-step block](#the-next-step-block). |
| Close with the footer (the first row of [close modes](../close/close-modes.md)) | The report, then [the close footer](#the-close-footer), then the waiting line of [questions](questions.md) «Format». |
| Close without the footer (the other rows of [close modes](../close/close-modes.md)) | The report, as the row of [close report](../close/close-report.md#variants) that fits, with no footer or command. |
| A single action that changes no repository file and is not a tracker write «Scope and authorization» already authorizes, such as creating a ticket, closing one as not planned, deleting a merged branch or saving to memory | Never routed through a flow's command: offer it as one more reply of the block, such as `C - Create the ticket`, when the message ends with one, and otherwise ask it as a decision of the round. The answer authorizes only that action; creating the ticket then follows [create-issue](../../create-issue/SKILL.md), in the same turn. An answer to another decision authorizes a ticket or other single action only when the chosen option's label or description, as the question tool shows it, names that action. |

## The reply block

Every reply the user can give, in one list, the recommended first; no command:

```
**▶ D<n> — <a question, such as How do we continue?>**

A - (★) <reply>: <why, when one is better>

B - <reply>

<waiting line>
```

## The next-step block

Moving from one flow to the next is never asked with the question tool, because a card hides the report it closes. End the report with a separate next-step block:

```
---
**▶ D<n> — Next step: <flow> (recommended)**

<the reason in one line>

`<the flow's command, once, in this host's syntax from the line below, with a real argument: a ticket key, a change folder or a short description of the work>`

A - <another reply, such as keep investigating or what to change>

<waiting line>
```

Command syntax by host, for the next-step block: `/flow-x` in Claude Code, Grok, Cursor and Antigravity; `$flow-x` in Codex; `/skill:flow-x` in Pi; `@flow-x` in OpenCode. Write the commands of several hosts only when you cannot tell which host you run in.

Offer a flow's command only when the next step needs that flow's procedure: an investigation, a design in Figma, a plan, or a change to the repository's files with its delivery. No flow needs a ticket: when the work has none, the command's argument is a short description of the work or its change folder, never a placeholder, and a ticket is proposed only as [scope changes](../plan/scope-changes.md) sets.

Only the user's reply to this block, or a flow command they type, starts the next flow; a reply to a question or the acceptance of a proposal only settles or redirects the scope, except a fix the close footer's **Resolve findings now** reply or a findings round sends, and the [same-turn continuations](#same-turn-continuations). Keep the question tool for decisions inside a flow: scope, design, delivery and review, validation and findings.

When the block leads from an approved plan to its build in the same session, or starts another work item after a close, add compacting first as one of its replies: the host's compaction command followed, where the host accepts instructions, by what to keep (the change folder or plan summary, the approved criteria and delivery answers, or the files and decisions the next work reuses). Offer it for that transition only, never for an estimated share of the context window, which the model cannot see. After it, end again with the same block; it is not one «during work» of the [checkpoint](../plan/checkpoint.md).

## The close footer

`flow-close` step 8 ends the report with this block, not with the question tool:

```
---
**▶ D<n> — <Close | See the findings | Continue> (recommended)**

<the reason in one line>

A - (★) <reply>

B - <reply>

<waiting line>
```

It recommends closing, unless a proposed finding is P1 or higher or a plan is in progress (continue it); with no item pending and neither case, the heading is **▶ D<n> — Closed**, with no reason or `(★)`, or **▶ Closed** alone when no reply applies. Its replies, each only when it applies:

- **Close:** runs the pending items, as [close modes](../close/close-modes.md) sets.
- **Resolve findings now**, naming them, such as "resolve H2 and H3 in this branch": they enter `flow-build` on the work's branch, with the code review of the fix and its checks, or `flow-plan` when the fix is not a small understood change.
- **See the findings**, with their count and highest severity: it shows the table and the round of [findings at the close](../close/close-report.md#findings-at-the-close).
- **Turn a candidate into a mechanism**, naming it, only for a candidate whose record counts two sessions or more; it enters `flow-build` as **Resolve findings now** does.
- **Find the next ticket**, run only when the user picks it, as [Find the next ticket](../../hive-mode/playbooks/ticket-status.md#find-the-next-ticket) sets.
- **Hygiene**, with the count of pending archives, only when [close hygiene](../close/close-hygiene.md#archive-the-change-folder) «When to offer hygiene» offers it.
- **Continue** with what remains, such as the next phase of a plan in progress.
- **Offer a `create-report` page**, as [close report](../close/close-report.md#final-report) sets.

## Same-turn continuations

Five continuations need no reply and no block; with the stages covered by the explicit [E2E grant](../../flow-build/references/e2e-coordination.md), they are the exceptions to «only the user's reply starts the next flow»:

- Research into `flow-build`, when the user's own request asked for a change and named it with its target, not whether or how to make one, and research finds it a small understood change ([proportionality](../plan/proportionality.md)), unless it recommends the create-design stage for a new screen. A reply to a question or a proposal never counts as that request.
- Research into create-design, for a request whose deliverable is the design, as [hive-mode](../../hive-mode/SKILL.md) step 0 sets.
- Research into [create-prd](../../create-prd/SKILL.md), for a request whose deliverable is a requirement document, as [hive-mode](../../hive-mode/SKILL.md) step 0 sets.
- Build into verify, as [flow-build](../../flow-build/SKILL.md) step 5 sets.
- Plan into build, after the plan's summary, when the user's own message asked beforehand to build after the plan without stopping, or [unattended delegation](../../unattended-delegation/SKILL.md) is active and its objective reaches the build.

Typing a flow's command loads that skill whole for that flow, and a flow or skill entered by continuation is read whole before its first step.
