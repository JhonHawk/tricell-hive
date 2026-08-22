---
order_claude: 143
order_agents: 51
targets: [claude, agents]
join: tight
---

- **Replacement scope resolves at the plan gate — coexistence is never assumed.** A task that rewrites, migrates, or replaces something that currently works (a module, build system, deploy pipeline, framework) settles ONE confirm-gated question before implementation (a structured question, folded into the plan gate): what survives of the legacy path? Three shapes, in cost order: **complete replacement** (legacy dies now — delete/archive; the default recommendation), **frozen retention** (legacy kept untouched as rollback insurance until a named milestone, e.g. production promotion — retention means NOT deleting; never fallbacks, dual-running, or work spent keeping legacy operational), **functional coexistence** (legacy stays live alongside the new path — dual maintenance, fallback wiring, ×2 test surface; only on the user's explicit pick, never as "the safe side"). A request that already answers it ("reemplazo completo") or a recorded prior decision discharges the ask. The answer lands in the plan/decision record and binds later agents and sessions: under replacement or retention, legacy code and half-built compat layers are deletion targets, not constraints — unrequested compat scaffolding discovered mid-run stops for this question instead of being completed; only the user reverses the decision.
