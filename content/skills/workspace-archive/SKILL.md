---
name: workspace-archive
description: Archive explicitly selected, explicitly closed support sessions while preserving their contents and links. Use for a requested archival operation, not routine cleanup.
---

# Archive selected closed sessions

Archive only sessions the user selected and whose closure is explicit in their retained record. The helper recognizes only `Status: closed|completed|cancelled` and `Estado: cerrado|completado|cancelado` (case-insensitive); it does not normalize those files. A closed research record may be archived without a plan. If a session has plans, every plan must carry one of those closed states. Do not classify work by age, normalize names or status values, stage files, commit, or push.

Use [the helper](scripts/archive_sessions.py) with a declared support root and selected session names. It reports a deterministic mapping without changes by default. `--apply` is allowed when the user's authorization already covers those selected moves. An ambiguous or open session stays in place and is reported.

Before apply, the helper checks every selected destination for collisions. It moves directories under `sessions/archived/`, preserves file bytes, and updates Markdown links that use the old support-relative session path. It never moves evidence or scratch with a session.

Report selected sessions, closure evidence, mappings, preserved or blocked sessions, and link updates. Broader artifact organization follows [workspace-conventions](../workspace-conventions/SKILL.md).
