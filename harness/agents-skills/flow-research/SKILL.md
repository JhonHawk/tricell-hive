---
name: flow-research
description: >
  Investigate a question using source code, documents, official documentation or user
  reports; compare evidence and return a supported answer. Use for research and substantive
  read-only investigation, including questions about how something works or whether a claim
  is supported. Not for a simple factual lookup, live project status, code review, or execution.
---

# Research an answer

Resolve the requested question with inspected evidence and stop at the answer. Infer the
scope and expected deliverable from the conversation; ask only when a missing decision
changes the investigation. Selecting this skill does not authorize implementation,
publication, deployment or a formal development plan.

For research-only work this procedure replaces the common intake in `task-routing` and
the discovery entry through `language-rules`; read the applicable references below
directly. Git operations and work beyond research still use their owning routes.

## Investigate

1. Identify the claim to settle, the sources that can settle it and what would contradict
   it. Check available local evidence before expanding the search.
2. Read the references matching the actual investigation before using their procedure.
   Do not load unrelated rows or reread a reference already in context.
3. Verify that each cited source supports the claim being made. Keep inspected facts,
   source-reported claims, inference and recommendations distinct; name contradictions
   and unresolved questions rather than filling gaps.
4. Return the answer with source locations and relevant versions or dates. Stop when the
   requested question is answered or the remaining evidence is unavailable; report the
   specific gap and its effect on the conclusion. A failed lookup is not evidence of absence.

## Read when needed

| Situation | Read before proceeding |
|---|---|
| External documentation, standards, ecosystem claims or user experiences | [Evidence research](references/evidence.md) |
| Code discovery, usage searches or an absence claim | [Code search](references/code-search.md) |
| A library, framework or CLI claim depends on its version | [Version-anchored documentation](references/context7.md) |
| Delegating, crossing domains or choosing independent verification | [Agent routing](references/agent-routing.md) |

Use the available research tools; report unavailable capabilities without inventing
tool names or results. When delegation is available, apply its routing reference; without
it, perform the bounded investigation directly and disclose any unmet verification criterion.

## Preserve only what was requested

The default deliverable is the conversational answer. Research alone creates no report,
session folder or direct-route task record; this takes precedence over the task-record
retention rule for reusable research conclusions.

When the user asks to keep the result, or before proposing its file location, load
`workspace-conventions` and its matching references. It remains the owner of artifact
placement inside and outside research; do not substitute a path template from this skill.
Keep memory operations under `memory-policy`, independently of artifact creation.
