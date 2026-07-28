---
name: unattended-delegation
description: >
  Load on an explicit user handover of unattended control ("full control tonight",
  "tienes control total", "run unattended", "ve cerrando los tickets") BEFORE
  declaring the mode accepted. Never activates from silence or a long task.
  Declare scope/gates/decision-log first. Codex/opencode channel; Claude Code
  always-on rules cover this.
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
- A tracker-scoped handover ("work the sprint board while I'm away") is the same mode
  with the project's DECLARED tracker bounding the scope: one reversible change-group
  and one decision-log entry per ticket; tracker writes batch to the close report.
- A non-delegated unattended turn (cron, background job, workflow stage) is NOT this
  mode: it stays fail-closed — stop and report blocked rather than assume approval.
