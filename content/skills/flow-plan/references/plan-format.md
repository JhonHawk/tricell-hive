# Retained plan format

Use this reference when creating or reviewing a retained plan. Keep one canonical `<topic>.plan.md` in the established work folder. Adapt the sections to the task; omit irrelevant sections and resolve or remove template placeholders before delivery. Do not create a companion research or report file just because the template mentions one.

Write the actual human-facing plan in the session language, including headings and current status. The English template below is reusable guidance, not a requirement that the delivered document be English. Preserve code, identifiers, and source names.

## What the document must carry

- The outcome, boundaries, constraints, and observable acceptance criteria.
- The inspected context and evidence behind consequential decisions, with relevant versions or dates. Use repository-relative links so another CLI can resolve sources in its own checkout. A path alone does not show what was learned there.
- The selected approach, affected contracts, meaningful assumptions, and unresolved blockers. Include migration or recovery behavior only when the work needs it. Record material scope clarifications and proposed expansions with their disposition. The essential decisions must be understandable without recovering the original conversation.
- For UI work, the relevant view decisions from [UI planning](ui-planning.md): application/area, guide status, shell/pattern and inspected reference, reuse, justified differences, states, and acceptance. Group equivalent views; omit this for work without a UI change.
- Checkbox tasks grouped by a coherent result, with dependencies, source locations, and an actual verification command or observation and its expected outcome.
- Current progress, evidence, and the next actionable step. Distinguish work that is ready, work that is authorized, and work that has been verified; one does not imply the others.

Keep the task list in this document. Link substantial research rather than duplicating it, and retain the same work identifier and initial date across sessions. Reconcile the record with current sources on resume; preserve valid completed work and explain material changes or reopened tasks.

## Adaptable template

```markdown
# <Work title>

Status: <current state and blocking condition, if any>
Authorized scope: <effects and targets already authorized; pending actions if relevant>

## Outcome and acceptance

<Expected user-visible or operational result, scope, constraints, and concrete criteria.>

## Context and decisions

<What relevant source inspection established, with links and freshness limits.>
<Selected approach, affected contracts, rationale, assumptions, scope clarifications or proposed expansions, and open decisions.>

## Tasks and verification

- [ ] <Coherent result>. Locations: <source paths>. Depends on: <task, if needed>.
  Check: <actual command or observation>. Expected: <specific outcome>.
  When relevant: <test approach, environment/prerequisites, gate timing, and unavailable coverage>.

## Progress and next step

<Completed work with observed evidence; failed or omitted checks; next action.>
```

For larger work, use short task subsections instead of dense checkbox paragraphs. Include only the detail needed to execute and verify safely; do not embed a full implementation by default. If a material decision is still unresolved, identify its impact and the affected tasks instead of labeling the plan ready. The skill's readiness check applies to the contents, not the presence of these headings.
