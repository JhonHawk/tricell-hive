---
alwaysApply: true
---

## Unattended Autonomy Mode

> Always-on stub: the activation guard and the gate pointers. The mode's full mechanics live in `rules-situational/unattended-autonomy-mode.md`, loaded through the `unattended-delegation` skill on every harness.

- **Activation is explicit-only.** Only an explicit user handover ("tienes control total esta noche", "no preguntes hasta que vuelva") or a **standing job** declared in the project ledger (`## Standing jobs`: trigger, scope, expiry; max action propose-never-land) activates the mode; silence, absence, or a long-running task never do. A *non-delegated* unattended turn (cron, workflow stage, background job — including a scheduled turn with no matching declared job) stays fail-closed per `quality/reporting-integrity.md > Fix at the Root`.
- **The absolute gates never relax under the mode** — each is owned by an always-on rule: destructive/irreversible ops, production, data deletion (`CLAUDE.md > Destructive Operations`); history rewrites, protected refs, promotion into the production-deploying branch (`workflow/git-workflow.md > Safety gates`); secrets and CRITICAL/HIGH supply chain (`quality/security.md`).
- **Full mode mechanics** (widened proceed-and-log scope, decision log, tracker-scoped runs, standing jobs, queued escalations, reversibility, expiry, close report): the `unattended-delegation` skill — load it BEFORE declaring the mode accepted. Declaring the mode without its controls is not the mode.
