---
name: flow-brainstorming
description: >
  Iterate a feature or business idea BEFORE any spec exists — the entry point of the
  process chain. Use when exploring whether an idea is worth doing: "¿y si el sistema
  hiciera…?", "el cliente quiere…", "¿vale la pena…?". Takes the idea in any form (a
  sentence, rough notes, a client email) — no document required. Produces a business
  decision (proceed / discard / defer), not a spec or technical design. Entry is
  question-gated: on detected brainstorming intent the session asks ONCE ("¿Iniciamos
  modo brainstorming?"); a No is sticky — manual invocation only for the rest of the
  session.
argument-hint: "<idea or topic>"
disable-model-invocation: true
---

# /flow-brainstorming — iterate the idea before it becomes a spec

This is the front of the chain: the business decision that `/flow-specs` later formalizes.
It deliberately stops short of technical design — the goal is to decide *whether and in
what shape* an idea is worth pursuing, not *how* to build it.

## OPEN

Take `$ARGUMENTS` (or the running conversation) as the idea — a phrase, notes, a client
message. No documentary prerequisite; if the idea is only in the user's head, draw it out
by asking.

Follow the flow contract (`~/.claude/skills/flow-core/SKILL.md`) *lightly*: if a workspace
ledger (`<project>/_support/PROJECT.md`) exists, read it for context so the idea is weighed
against what the project already is. No ledger yet — this is a brand-new project idea —
proceed conversationally; a decision to proceed on a greenfield idea suggests `/flow-start`
to stand the workspace up.

## Iterate

Explore the idea AND what carrying it out would take — conversationally, one thread at a
time, not a form to fill:

- **Business value** — what problem it solves, for whom, and why now. Is the pain real?
- **Minimal scope** — the smallest version that delivers the value (`development-principles.md
  > Solution proportional to the problem`). Name what to cut, not just what to build.
- **Contract / repo impact** — which services, contracts, or repos it touches; whether it
  fits the current shape or forces a new boundary.
- **Rough effort** — coarse sizing by scope and risk, never wall-clock hours.
- **Risks & unknowns** — what could make it not worth it, and what would have to be true.

Challenge the idea rather than sell it (`critical-thinking.md`): if a cheaper path reaches
the same value, say so; if the pain is speculative, name that.

**Contested questions route out.** When a question is genuinely disputed or a decision is
consequential — competing approaches with real trade-offs, a claim that needs verifying —
invoke `/adversarial-research` as a sub-tool and fold its canon back into the iteration.
Don't resolve a load-bearing dispute by assertion.

## Decide & record

Converge on one of: **proceed** / **discard** / **defer** — each with its *why* and the
scope it commits to (or the condition that would revive a deferred idea).

Write ONE light decision note — a short markdown record, not a ceremonial document:
- Workspace with a specs repo → `<project>-specs/decisions/<slug>.md`.
- Workspace without one yet → the session-capture home per `project-structure.md`
  (`<project>/_support/sessions/YYYY-MM-DD-<slug>/`), so it survives to be promoted later.
- No workspace at all → keep it beside the user's notes and name where it should land once
  a workspace exists.

The note carries: the idea, the decision, the why, the minimal scope, and the open
questions a spec will need to close. It is the input `/flow-specs` consumes — never a spec
itself and never technical design (schemas, endpoints, algorithms are out of scope here).

## CLOSE

If a ledger exists, record the decision note in it. On **proceed**, offer `/flow-specs` as
the next step (the exact invocation), pointing at the note. On **discard** or **defer**,
the note is the durable record of why — no next step is forced.
