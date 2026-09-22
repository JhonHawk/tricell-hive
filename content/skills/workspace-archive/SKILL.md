---
name: workspace-archive
description: Archive explicitly selected, explicitly closed support sessions while preserving their contents and links. Use for a requested archival operation, not routine cleanup.
---

# Archive selected closed sessions

Archive only sessions the user selected and whose closure is explicit in their retained record. The helper reads only the first status field in each record (the control-sheet row in `flow-plan` plans). Its value must be `closed`, `completed`, or `cancelled` (Spanish: `cerrado`, `completado`, `cancelado`), case-insensitive, either alone or followed by `·`, `.`, `;`, or a table border: `Status: closed`, `**Estado:** cerrado. …`, or `| Status | Completed · … |`. A qualified value such as `completado parcialmente` counts as open. It does not normalize those files. A closed research record may be archived without a plan. If a session has plans, every plan must carry one of those closed states. Do not classify work by age, normalize names or status values, stage files, commit, or push.

Use [the helper](scripts/archive_sessions.py) with a declared support root and selected session names. It reports a deterministic mapping without changes by default. `--apply` is allowed when the user's authorization already covers those selected moves. An ambiguous or open session stays in place and is reported.

Before apply, the helper checks every selected destination for collisions. It moves directories under `sessions/archived/` and rewrites relative targets of inline Markdown links in `*.md` files under the support root: links that point into a moved session, and relative links inside moved records that point elsewhere. Reference definitions, titled or angle-bracket links, and non-Markdown files such as HTML reports are not rewritten. It changes no other content and never moves evidence or scratch with a session. After apply, search the support root and the declared repository or workspace for the old `sessions/<folder>` and `../<folder>/` paths, and fix or report remaining links.

Report selected sessions, closure evidence, mappings, preserved or blocked sessions, and link updates. Broader artifact organization follows [workspace-conventions](../workspace-conventions/SKILL.md).
