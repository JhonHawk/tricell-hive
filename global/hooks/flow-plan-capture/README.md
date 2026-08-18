# flow-plan-capture

PostToolUse hook on `ExitPlanMode` (Claude Code only) — the organic bridge of the flow
pack ("Artifact-Attached Flow" design, 2026-07-10). When a native plan-mode plan is
approved inside a flow workspace (ledger `_support/PROJECT.md` at or above cwd), the plan
is captured to the session-capture layer (`sessions/YYYY-MM-DD-<slug>/<slug>-plan.md`,
`Status: planned`, sessions-index row) and the model is told the canonical path is the
working plan from there. `/flow-build` adopts that file (ADOPT step). Codex and opencode
have no equivalent event — they get the same convention as instructions via the workspace
`AGENTS.md` template in `flow-core/templates/workspace-agents.md`.

Design decisions (user-approved 2026-07-10):
- **No size threshold** — entering plan mode is the proportionality filter.
- **Opt-out rides the plan** (`Session: no` line) — the skip is explicit in the approved
  text, never a silent heuristic.
- **Idempotent by slug**: overwrite while `Status: planned`; `-2` suffix once advanced.
- **Breadcrumb log** at `$TMPDIR/claude-flow-plan-capture.log` — every write/skip auditable.
- Non-blocking: exit 0 always; silent outside flow workspaces.

Deployed by `/deploy-global` (script → `~/.claude/hooks/`, block merged into
`settings.json` from `settings-config.json`).
