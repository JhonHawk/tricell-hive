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
# Dual-runtime: Claude snake_case | Grok camelCase (keep aligned with bash-policy).
cwd=$(printf '%s' "$input" | jq -r '.cwd // .workspaceRoot // empty' 2>/dev/null)
[ -n "$cwd" ] || cwd="$PWD"
session_id=$(printf '%s' "$input" | jq -r '.session_id // .sessionId // empty' 2>/dev/null)
mode=$(printf '%s' "$input" | jq -r '.permission_mode // .permissionMode // empty' 2>/dev/null)

# Flow workspace? Walk up for the ledger. Not a flow workspace → nothing to do.
dir="$cwd"; ledger=""
while [ -n "$dir" ] && [ "$dir" != "/" ]; do
  if [ -f "$dir/_support/PROJECT.md" ]; then ledger="$dir/_support/PROJECT.md"; break; fi
  dir=$(dirname "$dir")
done
[ -n "$ledger" ] || exit 0

# ─── Phase section: captured plans that still have work outstanding ──────────
#
# The trigger is the ARTIFACT, never the ledger's declared phase. A plan file
# exists because native plan mode produced one and the plan-capture hook adopted
# it — so the signal is produced upstream of any flow command, instead of by the
# very command it would suggest (the old ledger-phase source could only be fed by
# a run that never happened, so it never fired). Entering plan mode IS the
# proportionality filter: a small change makes no plan and earns no offer.
phase_out=""
root=${ledger%/_support/PROJECT.md}
sessions_home=""
specs_sessions=$(find "$root" -maxdepth 2 -type d -path '*-specs/sessions' 2>/dev/null | head -1)
if [ -n "$specs_sessions" ]; then
  sessions_home="$specs_sessions"
elif [ -d "$root/_support/sessions" ]; then
  sessions_home="$root/_support/sessions"
fi

pending=""
if [ -n "$sessions_home" ]; then
  while IFS= read -r plan; do
    [ -n "$plan" ] || continue
    st=$(grep -m1 -E '^Status:' "$plan" 2>/dev/null | sed -E 's/^Status:[[:space:]]*([a-z]+).*/\1/')
    case "$st" in
      planned|building) pending="${pending}"$'\n'"- ${plan#"$root"/} (Status: $st)" ;;
    esac
  done < <(find "$sessions_home" -maxdepth 2 -type f -name '*-plan.md' 2>/dev/null | sort)
fi

# Re-fire when the pending set or any Status changes; stay quiet otherwise.
phase_fire=1
sig=$(printf '%s' "$pending" | cksum | cut -d' ' -f1)
if [ -n "$session_id" ]; then
  phase_marker="${TMPDIR:-/tmp}/claude-flow-context-phase-${session_id}"
  if [ -f "$phase_marker" ] && [ "$(cat "$phase_marker" 2>/dev/null)" = "$sig" ]; then
    phase_fire=0
  fi
fi
if [ "$phase_fire" = "1" ] && [ -n "$pending" ]; then
  phase_out="Flow workspace — captured plans with work outstanding (paths relative to $root):${pending}"
  phase_out="${phase_out}"$'\n'"- The Status header plus git IS the execution state; a checked box is not. Verify against git before trusting a Status."
  phase_out="${phase_out}"$'\n'"- Offer \`/flow-build\` ONCE this session to execute or resume, naming the direct route as the alternative; a no is sticky. Never invoke it uninvited (global rule: Skill Auto-invocation)."
  [ -n "$session_id" ] && printf '%s' "$sig" > "$phase_marker" 2>/dev/null || true
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
