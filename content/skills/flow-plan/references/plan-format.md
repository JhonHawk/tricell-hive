# Retained plan format

Use this reference when creating or reviewing a retained plan. A new retained plan is a change folder: its sections are distributed across `proposal.md`, `design.md`, and `tasks.md` as described in [change records](change-records.md). Continuing work keeps an established `<topic>.plan.md` as its single canonical record. Adapt the sections to the task; omit irrelevant sections and resolve or remove template placeholders before delivery. The plan needs no companion research, report, or history file.

Write the actual human-facing plan in the session language, including headings and current status. The English template below is reusable guidance, not a requirement that the delivered document be English. Preserve code, identifiers, and source names.

## What the document must carry

- The outcome, boundaries, constraints, and observable acceptance criteria.
- The inspected context and evidence behind consequential decisions, with relevant versions or dates. Use repository-relative links so another CLI can resolve sources in its own checkout. A path alone does not show what was learned there.
- The selected approach, affected contracts, meaningful assumptions, and unresolved blockers. Include migration or recovery behavior only when the work needs it. Record material scope clarifications and proposed expansions with their disposition. The essential decisions must be understandable without recovering the original conversation.
- For UI work, the relevant view decisions from [UI planning](ui-planning.md): application/area, guide status, shell/pattern and inspected reference, reuse, justified differences, states, and acceptance. Group equivalent views; omit this for work without a UI change.
- Delivery decisions from [the planning decision reference](delivery-decisions.md): Git mode or narrower instruction, repositories/base, authorized effects and stop conditions, implementation-review mechanism and pending consent.
- Local in-vivo applicability and rationale; if deemed unnecessary, the offered walk and user response or pending choice. For UI, include proportional coverage, built local environment, accounts/data prerequisites and the human walkthrough; record remote exceptions explicitly.
- Checkbox tasks grouped by a coherent result, with dependencies, source locations, and an actual verification command or observation and its expected outcome.
- For a saved implementation plan, the reviewed revision, relevant domain coverage, reconciled findings and unresolved review limits from [plan review](plan-review.md).
- Current progress, evidence, and the next actionable step. Distinguish work that is ready, work that is authorized, and work that has been verified; one does not imply the others.

Keep the task list in this document. Link already authorized retained research rather than duplicating it or creating a companion file, and retain the same work identifier and initial date across sessions. Reconcile the record with current sources on resume; preserve valid completed work and explain material changes or reopened tasks.

## Opening control sheet

Start with the five-row control sheet below, translated into the session language. Use short labels or phrases rather than paragraphs. Status combines phase and execution authorization; Git names the mode, agreed destination and next human stop; verification lists applicable gates only. Keep the objective outside the table as a meaningful prose explanation of the problem, intended change and expected result.

Infer the project tracker from project guidance, existing issue links or configured project context; do not ask again when it is clear. Include its name and a list of the issues covered by this plan inside the Tracker cell, one real linked issue per line (use `•` and `<br>` in Markdown tables). Preserve verified existing URLs; never invent an issue or URL. Ask only when tracker identity or issue scope is consequentially ambiguous. If there is no tracker or linked issue, state that briefly instead of creating one. Reading or linking issues does not authorize tracker writes.

Keep branch details, the full agreed delivery sequence, review mechanism, stack, accounts, walkthrough, revision hashes and history in the body. The concise Git row does not replace those decisions. An uncreated branch is proposed, not existing. Record relevant exclusions with their rationale in the scope; do not add a mandatory Boundaries row or infer an environment exclusion merely from a Git mode. Reconcile the sheet on resume so it summarizes the detailed decision record without becoming a competing authority.

Use these phases, translated for the reader: Draft (decisions or review still pending), Ready to implement (plan readiness satisfied, execution may still await consent), In progress (implementation underway), In validation (acceptance or delivery gates underway/pending), Completed (agreed endpoint achieved with required gates satisfied), Blocked (a named missing prerequisite prevents progress), or Paused (explicitly paused by the user). Completed follows the selected delivery endpoint: merge for interactive and automatic (the open PR under a human-review route), push to base for direct-base, verified working tree for hold. A phase never grants authorization. Identify the concrete reason and next action for Blocked or Paused.

## Adaptable template

```markdown
# <Work title>

| Field | Current value |
| --- | --- |
| Status | <phase · implementation authorized/pending; brief blocker if any> |
| Tracker · <detected name> | • [<issue ID — title>](<verified issue URL>)<br>• [<another included issue>](<verified issue URL>) |
| Git | <mode · agreed destination · next human stop, if any> |
| Verification | <applicable gates, e.g. tests · local in-vivo · UI> |
| Next step | <one concrete action; skill when useful> |

## Objective

<Explain the problem, intended change and expected user-visible or operational result in prose, with enough detail to understand why this work matters.>

## Scope and acceptance

<Included work, relevant exclusions with their rationale, constraints and concrete acceptance criteria.>

## Verified context

<Relevant current behavior and inspected sources, with revision/date where useful. Separate observed facts, assumptions and unverified conditions. Include only context that supports the proposed work.>

## Proposed design

<Selected approach and consequential choices, with their rationale. Name unresolved decisions and the tasks they block. Reference decisions here from tasks rather than repeating the explanation.>

<Add only applicable subsections: behavior and concurrency; contracts and version compatibility; data and persistence; UI by application/area; infrastructure; migration and recovery. Preserve important invariants, failure cases and ordering constraints.>

<For UI, identify each affected application/area and route, guide availability, shell/pattern and inspected reference, reuse, intended differences and relevant states. Group equivalent views; do not impose one shell across a monorepo.>

<For migration or recovery, distinguish design constraints needed for implementation from procedures for an environment transition that is not yet authorized.>

## Tasks

### T1 — <Coherent result>

- [ ] <Observable result>.

**Depends on:** <task IDs or none>.

**Locations:** <relevant files and symbols>.

**Execution:** <main thread or delegated to a role chosen by deliverable, with the reason>.

**Test approach:** <tdd, characterization, or check naming the check; omit when the task changes no behavior>.

**Changes:** <necessary steps; refer to design decisions instead of restating them>.

**Verification:** <actual command or observation and expected result; refer to a shared gate below when appropriate>.

## Verification and human review

<Define applicable tests, local in-vivo and UI coverage, prerequisites and expected outcomes. Keep task-specific checks in their tasks; collect shared gates here without duplicating commands.>

| Gate | Timing and environment | Mechanism and expected evidence | Authority / pending condition |
| --- | --- | --- | --- |
| <applicable gate> | <when, candidate and target> | <command or walkthrough and pass criteria> | <covered by implementation, explicit agreement, or pending choice> |

<For a rendered UI effect that the verification reference does not exempt, include the independent `review-ux` and `sdd-verify` gates with their base revision, viewports and themes, data and browser tool; they are part of implementation, not a pending choice.>

<For human UI review, specify the built local stack, account/role and data prerequisites without secrets, relevant viewports/states, and short navigation/actions with expected results. State that the verified stack stays available and selected temporary session captures are linked at handoff. Fill actual clickable URLs when known. Record remote exceptions explicitly.>

## Delivery

<Preserve user answers and their scope: repositories, base/work branches, Git mode, authorized effects, next human stop, code-review mechanism and agreed endpoint. Use a compact table when repositories or conditions differ. Do not infer merge, tracker writes or deployment authority from plan readiness.>

<Describe the sequence before and after the human stop and the applicable CI/review gates, referencing verification above. Include deployment and recovery steps only within the selected scope; distinguish a future offered promotion from an authorized action.>

## Review status and progress

<Current reviewed revision, relevant domains, reconciled material findings and unresolved coverage. Distinguish plan review from executed implementation tests.>

<Current progress and observed evidence, failed or omitted checks, blockers and the next actionable step. Keep the task checkboxes above current. Retain a short history only for material decisions, changed authority or reopened work; do not append every review round or repeat unchanged limits.>
```

Scale this structure to the work. Small tasks may combine sections and use a short checklist; larger tasks benefit from separate design subsections and task blocks. Omit irrelevant sections instead of filling them with boilerplate. No fixed word count or section count establishes completeness. Keep the objective in one or two focused paragraphs when sufficient, and the control sheet as short phrases.

Give each detailed decision, command and procedure one home within the plan. The control sheet summarizes; tasks reference the design and shared gates. Preserve details that prevent a plausible implementation error, such as race behavior or migration ordering; avoid embedding a full implementation by default. Keep current instructions prominent and replace superseded summaries, retaining only material history and review provenance needed to understand the current verdict.

If a material decision remains unresolved, identify its impact and affected tasks rather than labeling the plan ready. The skill's readiness check applies to the contents, not the presence of these headings.

Before closing, read back the canonical saved plan and confirm that the latest user answers, delivery endpoint and conditions are present and consistent with its status and closing recommendation. If native planning restrictions require an internal draft, state which copy is current; do not claim the canonical path was updated until its write and readback succeed within the permitted scope. On resume, use the canonical decision record and current evidence; a native approval message alone does not replace the user's instructions.
