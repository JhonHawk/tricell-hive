# Retained plan format

Use this reference when creating or reviewing a retained plan. Keep one canonical `<topic>.plan.md` in the established work folder. Adapt the sections to the task; omit irrelevant sections and resolve or remove template placeholders before delivery. Do not create a companion research or report file just because the template mentions one.

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

Keep the task list in this document. Link substantial research rather than duplicating it, and retain the same work identifier and initial date across sessions. Reconcile the record with current sources on resume; preserve valid completed work and explain material changes or reopened tasks.

## Opening control sheet

Start with the five-row control sheet below, translated into the session language. Use short labels or phrases rather than paragraphs. Status combines phase and execution authorization; Git names the mode, agreed destination and next human stop; verification lists applicable gates only. Keep the objective outside the table as a meaningful prose explanation of the problem, intended change and expected result.

Infer the project tracker from project guidance, existing issue links or configured project context; do not ask again when it is clear. Include its name and a list of the issues covered by this plan inside the Tracker cell, one real linked issue per line (use `•` and `<br>` in Markdown tables). Preserve verified existing URLs; never invent an issue or URL. Ask only when tracker identity or issue scope is consequentially ambiguous. If there is no tracker or linked issue, state that briefly instead of creating one. Reading or linking issues does not authorize tracker writes.

Keep branch details, the full agreed delivery sequence, review mechanism, stack, accounts, walkthrough, revision hashes and history in the body. The concise Git row does not replace those decisions. An uncreated branch is proposed, not existing. Record relevant exclusions with their rationale in the scope; do not add a mandatory Boundaries row or infer an environment exclusion merely from a Git mode. Reconcile the sheet on resume so it summarizes the detailed decision record without becoming a competing authority.

Use these phases, translated for the reader: Draft (decisions or review still pending), Ready to implement (plan readiness satisfied, execution may still await consent), In progress (implementation underway), In validation (acceptance or delivery gates underway/pending), Completed (agreed endpoint achieved with required gates satisfied), Blocked (a named missing prerequisite prevents progress), or Paused (explicitly paused by the user). Completed follows the selected delivery boundary: merge for the usual interactive route, verified working-tree work for hold. A phase never grants authorization. Identify the concrete reason and next action for Blocked or Paused.

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

## Context and decisions

<What relevant source inspection established, with links and freshness limits.>
<Selected approach, affected contracts, rationale, assumptions, scope clarifications or proposed expansions, and open decisions.>

| Decision | Rationale / evidence / pending condition |
| --- | --- |
| <consequential choice> | <reason, user answer or source and date; unresolved condition if any> |

<Keep detailed authorization evidence and conditions here when needed; do not repeat the control sheet verbatim.>

## Tasks and verification

- [ ] <Coherent result>. Locations: <source paths>. Depends on: <task, if needed>.
  Check: <actual command or observation>. Expected: <specific outcome>.
  When relevant: <test approach, environment/prerequisites, gate timing, execution mechanism and authorization, and unavailable coverage>.

## Progress and next step

<Completed work with observed evidence; failed or omitted checks; next action.>
```

For larger work, use short task subsections instead of dense checkbox paragraphs. Include only the detail needed to execute and verify safely; do not embed a full implementation by default. If a material decision is still unresolved, identify its impact and the affected tasks instead of labeling the plan ready. The skill's readiness check applies to the contents, not the presence of these headings.

Before closing, read back the canonical saved plan and confirm that the latest user answers, delivery endpoint and conditions are present and consistent with its status and closing recommendation. If native planning restrictions require an internal draft, state which copy is current; do not claim the canonical path was updated until its write and readback succeed within the permitted scope. On resume, use the canonical decision record and current evidence; a native approval message alone does not replace the user's instructions.
