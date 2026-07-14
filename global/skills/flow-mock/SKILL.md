---
name: flow-mock
description: >
  Build or review the navigable prototype that replaces Figma in the project flow (F4 of
  the flow pack): build constructs the mock epic-by-epic against the specs using the
  project's design stack (HeroUI or shadcn); review navigates it live and reports UX
  friction with evidence. Use after specs are reviewed and before foundation.
argument-hint: "[build | review [flow-or-epic]]"
disable-model-invocation: true
---

# /flow-mock — navigable prototype

Follow the flow contract (`~/.claude/skills/flow-core/SKILL.md`). The mock is a
**prototype, not a deliverable**: its job is to expose flow/UX problems while changing
them is nearly free, before the backend hardens decisions. Findings that change behavior
flow BACK to specs — the mock never silently becomes the spec.

## `build`

1. OPEN per the contract. Read the product map and vistas in `<project>-specs/product/`
   (the business rules in force, screen by screen) plus the reviewed epics in
   `<project>-specs/epics/` (the deltas being mocked) — the mock implements what the
   vistas say, with realistic fake data (volumes per the spec: include the 0-item and
   the 10,000-item case).
2. Detect or decide the design stack — never as a separate question. Existing mocks repo
   → mirror its stack. New → pick from workspace signals (the client's existing frontends,
   declared conventions) and carry the pick into the plan gate (step 4) as a named
   decision: HeroUI (load the `heroui-react-pro` + `heroui-pro-design-taste` skills) or
   shadcn (load the `shadcn` skill + `design-taste-frontend`). The loaded skill governs
   component usage; don't hand-roll components it provides.
3. The mocks repo is a real repo (`<project>-frontend-mocks` by convention — confirm the
   name against the naming table if it exists): create per `/flow-foundation` conventions
   if missing, session branches per git-workflow.
4. **Light plan gate** (gate ∝ reversibility, per the flow contract): building a whole repo
   is cheap to rebuild but costly to redo from a wrong frame, so present a compact,
   approvable plan BEFORE building — the screens per epic with their spec-defined states
   (empty/loading/error), the stack pick from step 2 (the ONE consolidated interaction —
   no separate stack question), and the build order. Not a full flow-plan executable plan;
   ~5 lines. Gate by harness mode:
   - **Native plan mode active**: the plan document is the presentation — finalize and exit
     via the mode's approval mechanic; do not duplicate it as a summary message.
   - **Any other mode**: present the plan as a normal message, then gate with the decision
     mechanic (`AskUserQuestion` in Claude Code; `references/harness-mechanics.md` for others).
5. Build epic by epic. Implementation can run in the main thread for velocity (prototype
   carve-out: throwaway-quality code is acceptable here) — but UI states the spec defines
   (empty/loading/error) are NOT optional polish; they're what `review` evaluates.
6. CLOSE per the contract: ledger row per epic mocked, routes documented in the mocks
   repo README, and the `## Current handoff` names the mocks repo as input for `/flow-plan`.

## `review [flow-or-epic]`

1. Ensure the mock is running — start the dev server if needed (declare it, no
   authorization needed); stop it at CLOSE unless the user is actively validating
   (global `Execution` rule — never orphan, never ask by default).
2. Dispatch **ux-flow-reviewer** per the handoff protocol
   (`~/.claude/skills/flow-core/references/handoff-protocol.md`) with: the running URL,
   the flows to walk (from `$2`, or every flow of the epics marked mocked in the ledger),
   the persona/role per flow (from the epic's PRODUCT.md), the rubric path
   (`${CLAUDE_SKILL_DIR}/references/ux-rubric.md`), and the evidence path
   (`<project>/_support/evidence/mock-review-<date>/`).
   Do not review from the main thread — the reviewer's fresh eyes on the running UI are
   the point of the gate.
   **The Flow axis is mandatory and blocking: the mock does not close until every flow
   passes it.** A `blocker` finding is **not a pass** — fix it (`mock-fix`) or route it
   (`spec-change`) and re-walk the affected flow before closing. **Visual-craft fails
   block by signal, not uniformly**: the mock is the visual reference the build will copy,
   so fails that would propagate into it (WCAG-AA contrast, action hierarchy, type scale)
   are fixed before close; the remaining craft findings (spacing, elevation, borders,
   component simplicity) — like `friction`/`polish` — may close with the user's explicit
   acknowledgement, recorded in the report.
3. Triage the findings with the user:
   - `mock-fix` → apply directly in the mocks repo (presentation only)
   - `spec-change` → **do not fix in the mock first**; record it and trigger
     `/flow-specs epic <name>` so the source of truth moves, then the mock follows
4. Render the report via `flow-report` as `mock-review-<slug>.html`, routed per
   flow-core's **Session reports** rule (raw screenshots to `_support/evidence/<slug>/`,
   referenced by path). CLOSE per the contract, including the list of spec-changes
   triggered (they are decisions — promotion targets).
