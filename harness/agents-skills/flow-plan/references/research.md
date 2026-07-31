# /flow-plan `research` — technical investigation (read-only, no gate)

Reduces the unknowns so `write` can produce a complete recipe. Scoped to **implementation**, not
product or UX — those are `flow-specs` (epics/ACs) and mock work units' UX material, already upstream.

1. **OPEN** per the flow contract (`~/.claude/skills/flow-core/SKILL.md`): read
   `<project>/_support/PROJECT.md`; missing → suggest `/flow-start` and stop. Consume any
   `## Current handoff`.
2. Read the target epic in `<project>-specs/epics/` (PRODUCT.md, tasks.md, TECH.md if present)
   and the naming table if infra is in scope. Collect the epic's **parked technical questions**
   (Open Questions marked `technical — resolves in TECH.md` by the business gate) — resolving
   them is part of this investigation, and the resolutions land in the epic's TECH.md, not in
   PRODUCT.md. `$ARGUMENTS` after the subcommand overrides scope (epic ID or task IDs).
3. **EXPLORE in subagents, never inline** (context hygiene — discovery noise stays out of the
   orchestrator). Dispatch read-only explorers per the handoff protocol
   (`~/.claude/skills/flow-core/references/handoff-protocol.md`) — independent
   areas in ONE message, in parallel — to establish:
   - **Current state** — how this repo already does the thing; patterns, conventions, and
     existing code to reuse (`development-principles.md > Search/Observe before creating`).
   - **Approaches** — 2-3 technical options with trade-offs and a recommendation.
   - **Gaps and prerequisites** — missing interfaces, uninstalled libs, migrations, env/config,
     credentials (the `gap-resolution.md` Investigate step, materialized as an artifact).
   - **Risks** — failure modes and where the plan will need judgment.
   - **Version-sensitive surfaces** — query context7 for the installed versions so `write` can
     cite exact signatures.
4. **Write `<slug>-findings.md`** to the session/initiative home (per `project-structure.md`):
   current-state map, approaches + recommendation, gaps/prerequisites, risks, version notes.
   Read-only — no source changes, no gate.
5. **Divergence routes back.** If research finds the epic infeasible as specced, or a clearly
   better approach implies changing the spec, surface it and route to `/flow-specs` — do not
   silently deviate (`gap-resolution.md > Divergence Between Sources`).
6. **CLOSE**: update the ledger handoff with the findings path; suggest `/flow-plan write`.
