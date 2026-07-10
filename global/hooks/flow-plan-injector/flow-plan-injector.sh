#!/usr/bin/env bash
# flow-plan-injector.sh — UserPromptSubmit hook, NON-BLOCKING, plan-mode only.
#
# Companion of flow-plan-capture: while the user is IN native plan mode inside a
# flow workspace, injects — once per session — the artifact conventions the plan
# should follow, so the capture lands well-formed. It never fires outside plan
# mode, outside flow workspaces, or twice in a session; small ad-hoc sessions
# never see it. Replaces the retired flow-route-reminder (its per-prompt routing
# payload had documented failure modes; this is transition-scoped and factual).
#
# Never blocks (exit 0 always; stdout becomes model context).

set -uo pipefail

input=$(cat)
cwd=$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null)
[ -n "$cwd" ] || cwd="$PWD"
session_id=$(printf '%s' "$input" | jq -r '.session_id // empty' 2>/dev/null)
mode=$(printf '%s' "$input" | jq -r '.permission_mode // .permissionMode // empty' 2>/dev/null)

# Fire only in plan mode; absent field → stay silent (capture still works alone).
[ "$mode" = "plan" ] || exit 0

# Flow workspace? Walk up for the ledger.
dir="$cwd"; ledger=""
while [ -n "$dir" ] && [ "$dir" != "/" ]; do
  if [ -f "$dir/_support/PROJECT.md" ]; then ledger="$dir/_support/PROJECT.md"; break; fi
  dir=$(dirname "$dir")
done
[ -n "$ledger" ] || exit 0

# Once per session.
if [ -n "$session_id" ]; then
  marker="${TMPDIR:-/tmp}/claude-flow-plan-injector-${session_id}"
  [ -f "$marker" ] && exit 0
  : > "$marker" 2>/dev/null || true
fi

cat <<'EOF'
Flow workspace, plan mode. Conventions for the plan you are writing:
- On approval it is captured automatically to sessions/YYYY-MM-DD-<slug>/<slug>-plan.md (flow-plan-capture); that copy becomes the working plan.
- Include a `Session: yes` header line; the user flips it to `Session: no` to decline the session folder — their call, never yours.
- Large scope (multi-session): structure the plan as an initiative — parts with a per-part Status — so /flow-build can adopt and resume part by part.
- Investigation conclusions the user asks to keep go to <slug>-findings.md in the same session folder.
- Do not embed git/integration semantics (branching, PR/merge, CI) the conversation did not settle — /flow-build confirms that delta at adoption.
EOF
exit 0
