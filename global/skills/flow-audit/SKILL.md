---
name: flow-audit
description: >
  Multi-lens preventive audit of a project's runtime repos, delivering a refuted,
  severity-ranked inventory with per-repo health score. Use at epic close, before a major
  architectural change, or when 2+ chained bugs share one origin ("audita el proyecto",
  "línea base antes de producción"). Lens argument scopes the run
  (full | smells | security | db | perf | arch | secrets). Produces
  <project>-specs/audit/reports/YYYY-MM-DD.html plus a history-table row.
argument-hint: "[full | smells | security | db | perf | arch | secrets] [<repos…>]"
disable-model-invocation: true
---

# /flow-audit — multi-lens preventive audit

Executes `~/.claude/skills/flow-core/references/audit-playbook.md` — read it first; it is the
protocol (lenses, severity scale, finding format, dedup + refuter stage, report structure).

Skill-specific mechanics on top of the playbook:

1. **OPEN per the flow contract** (`~/.claude/skills/flow-core/SKILL.md`). Locate the project
   instance `<project>-specs/audit/README.md`; if absent, offer to bootstrap it from the
   playbook's instantiation section (that write is part of this run, versioned in the specs
   repo). No specs repo at all → no carve-out: route the report through the standard
   session-capture form (`<repo>/_support/sessions/YYYY-MM-DD-audit-<lens>/reports/`),
   flagged in the report as pre-specs-repo.
2. **Resolve the lens set** from the argument (default: ask, recommending `full` for a first
   baseline and a scoped lens for follow-ups). Resolve lens→agent against the CURRENT
   `agent-routing.md` roster — never the instance README's remembered names.
3. **Dispatch readers in ONE message** with the playbook's briefing contract (paths, doctrine,
   format, cap, known-debt exclusions). Readers are read-only; they identify, never remediate.
4. **Dedup → refute → consolidate** per the playbook. Render via flow-report, passing the
   explicit versioned destination (`<project>-specs/audit/reports/YYYY-MM-DD.html` — the
   caller-directed record; never let it default to `_support/workspace/`); append the
   history row; propose tracker tickets for P0/P1 batched to the close confirmation.
5. **CLOSE per the contract**: ledger pointer, handoff line, per-criterion state (lenses run /
   skipped, refuter verdicts, report path).

Comparisons across runs are valid only same-lens; the report's baseline section says which.
