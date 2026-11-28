# <TICKET, or the work name without a ticket> — <one-line title>

<!-- Template from flow-plan. Write the text and the headings in the session language; keep code, identifiers and the headings `## Effects`, `## Business rules`, `## Checkpoint` and `## Review status and progress` as they are, since the rules and `scripts/tasks.py` look for them. Name the first field `Status` and start its value with the phase; keep the field name and its values `Draft`, `Ready for flow-build`, `In build` and `Completed` in English whatever the session language, since `scripts/tasks.py status-set` writes them and the ticket-status playbook reads `Completed`. This file lives at `<support>/workspace/<YYYY-MM-DD>-<work>/<YYYY-MM-DD>-<work>.plan.md` and is never committed. Delete these comments and any section that does not apply. -->

| Field | Value |
| --- | --- |
| Status | <Draft · Ready for flow-build · In build · Completed> |
| Tracker | <ticket link · state · priority · related tickets, or none> |
| Git | <repository, base branch and commit, proposed branch> |
| Playbook | <the playbook `hive-mode` opened, such as bugfix or feature, or none> |

## Decisions

<!-- One line per decision the user took, under its date: the decision ID and letter in bold, a dash, the question in plain words, and the answer and what it changes, such as "- **D62 B** — Who links each event to its stage? The server; the agent no longer declares it.". No line here or in any other section relies on an ID alone. A decision taken without asking, as «Decide and list» of the `questions` leaf sets, goes here too, marked with the session language's word for "assumption" in place of the ID, with the alternative discarded; a business rule taken that way also gets its `assumption` line under `## Business rules`. -->

<date>
- **<D<n> letter>** — <question?> <answer and what it changes>.

## Business rules

<!-- One line per business rule the work adopts that its ticket and documentation did not state, and per rule a decision cancels, in the format the business-rules leaf sets; never delete a line. The ticket's control lines and the `card stale:` and `frame stale:` lines go here too. When the work adopts none, keep only the `None:` line; otherwise delete it. -->

None: the work adopts no business rule beyond what its ticket and documentation state.

rule: <the rule> · screen <name, or none> · <status> · found by <drawer | reviewer | research | plan | build | verify>

## Request

<The user's request, quoted. The ticket's acceptance criteria, quoted, with the date they were read. What the ticket leaves out of scope.>

## Objective

<What happens today, and what happens when the change is done.>

## Scope and acceptance

<What is in scope and what is out of it.>

<!-- Format reminder, rule in the plan-record leaf, «Revise a plan in progress»: a revision marks the new decision or criterion with `**Replaces:** <old IDs>` on its line, and a dropped criterion keeps its line with `**Retired:** <reason>`. Both markers stay in English, whatever the session language, and so does `Observed in:`. -->

- AC1 — <one observable behavior> · false today because <the observation that shows it> · Observed in: <where its real effect shows>
- AC2 — <…>

## Prerequisites

<!-- What the build needs from outside the repository, such as an API key, an account, an access grant or data. Never write a secret's value: name where it is stored. When there are none, keep only the next line. -->

None: the build needs nothing from outside the repository.

- [ ] <what is needed> · owner: <who provides it> · needed by: <T<n>>

## Verified context

<What was read and observed before planning, with paths, commits and dates.>

<!-- Only when the approach reads or changes an external system: one line per external fact, as the external-facts leaf sets; a fact no read-only read could show is written as not measured, with the task that meets it first. -->

- <external fact> · <read-only command or source> · <result, or not measured> · <date>

<!-- One line per command a task's `Verification:` line names, as flow-plan's Verification lines (`flow-plan/SKILL.md#verification-lines`) sets: the command, its result or not measured, and the date. -->

- <verification command> · <result, or not measured with the task that meets it first> · <date>

## Effects

<!-- One line per class of effect asked in the plan's effects round. Keep this heading in English: flow-deploy looks for it. When the plan foresees no effect on an external system, keep only the `None:` line; otherwise delete it and keep one line per class. -->

None: the tasks take no effect on an external system.

- <class of effect> · target: <account, repository or environment> · ceiling: <amount, or 0> · tasks: <T<n>> · answer: <approved, refused or unanswered, with the decision ID and letter when answered>

## Delivery

<The delivery mode, its validation stop when the work has a surface, and its endpoint. The code review mechanism. What this choice does not authorize. The public record the pull request body carries: objective, criteria with their evidence, decisions and the business-rule record.>

## Design

<!-- Only when there are design decisions to record; delete the section otherwise. -->

**Approaches.** <The two or three ways of building it that were presented, each with its tradeoff, the one chosen and why; or why only one way was sensible.>

**Decisions.** <Each design decision with its options, the one chosen and why. Contracts, data shapes, components and how they connect.>

**UI.** <Only for a change to a screen: the pattern, where each action goes and its weight, the states (loading, empty, error, permissions) and an ASCII sketch of the region, or the record of its approved example.>

**Copy.** <Only when the change adds or changes text people read: each text, in every language the product ships.>

**Impact.** <What uses what changes, found by searching its literal values and callers, and whether it touches production.>

**Recovery.** <How to undo the change, and what cannot be undone.>

## Tasks

<The base commit of each repository when the plan was written, and the branch. Which tasks run in the main thread or in children, and why.>

### T1 — <what it delivers>

- [ ] <the observable result of this task>

**Closes:** <AC<n>, …>

**Depends on:** <T<n>, or none>

**Locations:** <files and packages it changes>

**Execution:** <main thread, or the role that runs it>

**Test approach:** <tdd · characterization · check, as the tests leaf `hive-core/build/tests.md` defines them; omit for a mechanical change>

**Risk:** <its risky-change category from global «Proportionality», or none>

**Changes:** <what changes and what deliberately does not>

**Verification:** <the command or observation, and its expected result>

## Delivery units

<!-- Optional: keep this section only when the plan ships in two or more delivery units (PRs, commits to the base under direct-base, or candidates under hold); delete it otherwise. Boxes here belong to no task, so `scripts/tasks.py done` never ticks them; flow-verify ticks a box by hand when its check passes and records the result under `### Evidence`, not in the box. `n/a: <reason>` replaces the predicate of a box that does not apply. -->

### E1 — <PR, commit or candidate> · depends on: <E<m>, or none> · files: <paths>

- [ ] unit: <command>
- [ ] live: <scenario> · pass: <predicate>
- [ ] perf: <metric> · base: <measured value> · fail: <threshold>

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | <command> | Covered by the implementation |
| UI check | Final candidate, locally | <viewports, themes, reviewer> | Covered by the implementation |
| Code review | Final candidate, before the push or the validation stop | <mechanism> | The review answer |

<The human walk, when the delivery stops for validation: the stack to serve, the account, and the steps.>

## Checkpoint

<Written by `scripts/tasks.py checkpoint`: Updated, Stage, Last commit, Open findings, Next, Worktree when the work has one (`- Worktree: <path> · <branch>`), Unwritten at a pause (`--unwritten`), Rules with the stage packets read (`--rules`), Principles with the code principles that changed a decision (`--principles`), the `- Execution record:` link when an E2E execution exists, then `### Evidence` with the checks recorded so far.>

## Review status and progress

<Date — the plan review and what it changed. Then one line per task done: its commit and anything left unverified, and one line per progress event.>
