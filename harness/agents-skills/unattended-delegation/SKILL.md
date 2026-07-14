---
name: unattended-delegation
description: >
  Activate ONLY when the user explicitly hands over unattended control — "full control
  tonight", "don't ask until I'm back", "tienes control total esta noche", "me voy a
  dormir, sigue tú", "run unattended". Read BEFORE declaring the mode accepted, then
  declare the scope, the gates that stay closed, and the decision-log location. Covers
  proceed-and-log widened scope, queueing gated decisions instead of timing out, the
  dedicated-branch requirement, run-bound expiry, and the close report. Absolute safety
  gates never relax in any mode. Silence or a long-running task never activates this;
  only an explicit user declaration does. Codex/opencode channel; Claude Code receives
  these rules always-on via its own rules — do not load there.
---

# unattended-delegation — explicitly-delegated unattended runs

The full canonical mode lives in `references/unattended-autonomy.md` (injected at build
time from `global/rules/workflow/` — single source of truth). The mode is a trade:
removed confirmations are compensated by ADDED controls (decision log, reversible
checkpoints, queued escalations, automatic expiry) — without those controls the
delegation is not in effect.

## Rules of use

- Read `references/unattended-autonomy.md` in full BEFORE declaring the mode active.
- On activation, declare visibly: the scope accepted, the absolute gates that stay
  closed, and where the decision log lives. Name the mode in every report produced
  under it.
- The absolute gates (destructive/irreversible ops, production, secrets, history
  rewrites/force-push, shared-ref mutation, CRITICAL/HIGH supply chain, data deletion)
  are also in the always-on Safety floor — they apply whether or not this skill loads.
- A non-delegated unattended turn (cron, background job, workflow stage) is NOT this
  mode: it stays fail-closed — stop and report blocked rather than assume approval.
