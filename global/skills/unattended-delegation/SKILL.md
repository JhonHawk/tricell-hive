---
name: unattended-delegation
description: >
  All harnesses: load on an explicit user handover of unattended control ("full control
  tonight", "tienes control total", "run unattended", "ve cerrando los tickets") BEFORE
  declaring the mode accepted. Never activates from silence or a long task. Declare
  scope/gates/decision-log first.
---

# unattended-delegation — explicitly-delegated unattended runs

The mode is a trade: removed confirmations are compensated by ADDED controls (decision
log, reversible checkpoints, queued escalations, automatic expiry) — without those
controls the delegation is not in effect. Loading this skill IS how those controls enter
the run, on every harness.

**Where the reference lives** — same canonical file, two paths:

| Harness | Read from |
|---|---|
| Claude Code | `~/.claude/rules/workflow/unattended-autonomy.md` |
| Codex / opencode | `references/unattended-autonomy.md` (injected at build time) |

## Rules of use

- Read it in full BEFORE declaring the mode active. Declaring the mode without reading it
  is declaring a mode with no decision log, no dedicated branch, and no expiry — that is
  not this mode.
- On activation, declare visibly: the scope accepted, the absolute gates that stay
  closed, and where the decision log lives. Name the mode in every report produced
  under it.
- The absolute gates apply whether or not this skill loads — each is owned by an always-on
  rule, which is what makes the mode safe to keep behind a trigger:
  destructive/irreversible ops · production · data deletion → `CLAUDE.md > Destructive
  Operations`; history rewrites/force-push · push or merge to shared/protected refs →
  `workflow/git-workflow.md > Safety gates`; secrets · CRITICAL/HIGH supply chain with no
  safe path → `quality/security.md`.
- A tracker-scoped handover ("work the sprint board while I'm away") is the same mode
  with the project's DECLARED tracker bounding the scope: one reversible change-group
  and one decision-log entry per ticket; tracker writes batch to the close report.
- A non-delegated unattended turn (cron, background job, workflow stage) is NOT this
  mode: it stays fail-closed — stop and report blocked rather than assume approval.
