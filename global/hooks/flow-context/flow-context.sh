#!/usr/bin/env bash
# flow-context.sh — UserPromptSubmit hook, NON-BLOCKING, flow workspaces only.
#
# Merges the two former UserPromptSubmit hooks (flow-phase-context +
# flow-plan-injector) into one script with TWO INDEPENDENT emit decisions —
# deliberately NOT collapsed, because their re-fire semantics differ:
#
#   - Phase section: injects the ledger's `Current phase` / `Next suggested`
#     lines so the model can OFFER the right /flow-* stage. Keyed to the
#     ledger's MTIME so a phase transition mid-session RE-FIRES (stale injected
#     state would misroute the offer).
#
#   - Plan section: only in native plan mode, injects the artifact conventions
#     the plan should follow so flow-plan-capture lands well-formed. Fires at
#     most ONCE per session (a plan's conventions don't change mid-plan).
#
# One marker per section (a shared marker would let the once-per-session plan
# gate suppress the mtime-driven phase re-fire, or vice versa). Payload is
# STATE, not routing instructions. Never blocks (exit 0 always; stdout becomes
# model context).

set -uo pipefail

input=$(cat)
cwd=$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null)
[ -n "$cwd" ] || cwd="$PWD"
session_id=$(printf '%s' "$input" | jq -r '.session_id // empty' 2>/dev/null)
mode=$(printf '%s' "$input" | jq -r '.permission_mode // .permissionMode // empty' 2>/dev/null)

# Flow workspace? Walk up for the ledger. Not a flow workspace → nothing to do.
dir="$cwd"; ledger=""
while [ -n "$dir" ] && [ "$dir" != "/" ]; do
  if [ -f "$dir/_support/PROJECT.md" ]; then ledger="$dir/_support/PROJECT.md"; break; fi
  dir=$(dirname "$dir")
done
[ -n "$ledger" ] || exit 0

# ─── Phase section: once per session, re-firing on ledger mtime change ───────
phase_out=""
mtime=$(stat -f %m "$ledger" 2>/dev/null || stat -c %Y "$ledger" 2>/dev/null || echo 0)
phase_fire=1
if [ -n "$session_id" ]; then
  phase_marker="${TMPDIR:-/tmp}/claude-flow-context-phase-${session_id}"
  if [ -f "$phase_marker" ] && [ "$(cat "$phase_marker" 2>/dev/null)" = "$mtime" ]; then
    phase_fire=0
  fi
fi
if [ "$phase_fire" = "1" ]; then
  # Accept both vocabularies: `Current stage` (v2 ledger template) and `Current phase` (pre-v2 ledgers).
  phase=$(grep -m1 -E '^\| *Current (stage|phase) *\|' "$ledger" 2>/dev/null | sed -E 's/^\| *Current (stage|phase) *\| *(.*[^ ]) *\|.*$/\2/')
  next=$(grep -m1 -E '^- *Next suggested:' "$ledger" 2>/dev/null | sed -E 's/^- *Next suggested: *//')
  # Ledger present but both fields missing/blank → skip the phase section.
  if [ -n "$phase$next" ]; then
    phase_out="Flow workspace — state from the ledger ($ledger):"
    [ -n "$phase" ] && phase_out="${phase_out}"$'\n'"- Current stage: $phase"
    [ -n "$next" ] && phase_out="${phase_out}"$'\n'"- Next suggested step: $next"
    phase_out="${phase_out}"$'\n'"- Advisory: the ledger can be stale — verify before relying on it. Flow commands are offered to the user, never invoked uninvited (global rule: Skill Auto-invocation)."
    # Record the mtime only once we actually emit (so a blank ledger re-checks next prompt).
    [ -n "$session_id" ] && printf '%s' "$mtime" > "$phase_marker" 2>/dev/null || true
  fi
fi

# ─── Plan section: plan mode only, once per session ──────────────────────────
plan_out=""
if [ "$mode" = "plan" ]; then
  plan_fire=1
  if [ -n "$session_id" ]; then
    plan_marker="${TMPDIR:-/tmp}/claude-flow-context-plan-${session_id}"
    if [ -f "$plan_marker" ]; then plan_fire=0; fi
  fi
  if [ "$plan_fire" = "1" ]; then
    plan_out=$(cat <<'EOF'
Flow workspace, plan mode. Conventions for the plan you are writing:
- On approval it is captured automatically to sessions/YYYY-MM-DD-<slug>/<slug>-plan.md (flow-plan-capture); that copy becomes the working plan.
- Include a `Session: yes` header line; the user flips it to `Session: no` to decline the session folder — their call, never yours.
- Large scope (multi-session): structure the plan as an initiative — parts with a per-part Status — so /flow-build can adopt and resume part by part.
- Investigation conclusions the user asks to keep go to <slug>-findings.md in the same session folder.
- Do not embed git/integration semantics (branching, PR/merge, CI) the conversation did not settle — /flow-build confirms that delta at adoption.
EOF
)
    [ -n "$session_id" ] && : > "$plan_marker" 2>/dev/null || true
  fi
fi

# ─── Emit whichever sections passed their gates (phase first) ────────────────
[ -n "$phase_out$plan_out" ] || exit 0
if [ -n "$phase_out" ]; then
  printf '%s\n' "$phase_out"
fi
if [ -n "$plan_out" ]; then
  [ -n "$phase_out" ] && printf '\n'
  printf '%s\n' "$plan_out"
fi
exit 0
