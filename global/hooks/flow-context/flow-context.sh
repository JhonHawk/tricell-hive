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
#   - Planning section: injects the portable Hive planning command and artifact
#     conventions once per session. It does not inspect or depend on a native
#     harness plan mode; native planning remains optional user tooling.
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
# exists because Hive's portable planning flow produced one, so the signal is
# produced upstream of any execution command instead of by the command it would
# suggest. A small change can still take the direct route and earns no offer.
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

# ─── Planning section: portable Hive flow, once per session ──────────────────
planning_out=""
planning_fire=1
if [ -n "$session_id" ]; then
  planning_marker="${TMPDIR:-/tmp}/claude-flow-context-planning-${session_id}"
  if [ -f "$planning_marker" ]; then planning_fire=0; fi
fi
if [ "$planning_fire" = "1" ]; then
  planning_out=$(cat <<'EOF'
Flow workspace, portable planning conventions:
- Planning intent uses `/flow-plan`; it creates or updates Hive's durable plan artifact and does not depend on native harness Plan Mode.
- Explicit approval authorizes the recorded plan revision; native harness planning remains optional and never grants Hive authorization by itself.
- Include a `Session: yes` header line; the user flips it to `Session: no` to decline the session folder — their call, never yours.
- Large scope (multi-session): structure the plan as an initiative — parts with a per-part Status — so `/flow-build` can adopt and resume part by part.
- Investigation conclusions the user asks to keep go to `<slug>-findings.md` in the same session folder.
- Do not embed git/integration semantics (branching, PR/merge, CI) the conversation did not settle — `/flow-build` confirms that delta at adoption.
EOF
)
  [ -n "$session_id" ] && : > "$planning_marker" 2>/dev/null || true
fi

# ─── Emit whichever sections passed their gates (phase first) ────────────────
[ -n "$phase_out$planning_out" ] || exit 0
if [ -n "$phase_out" ]; then
  printf '%s\n' "$phase_out"
fi
if [ -n "$planning_out" ]; then
  [ -n "$phase_out" ] && printf '\n'
  printf '%s\n' "$planning_out"
fi
exit 0
