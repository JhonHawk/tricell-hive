#!/usr/bin/env bash
# delegation-reminder.sh — PostToolUse (all tools) hook, NON-BLOCKING.
#
# Deterministic backstop for the ~20-undelegated-tool-calls delegation gate
# (rules/workflow/agent-routing.md > Delegation Gates). Counts main-thread
# tool calls per session; each time the count crosses a multiple of 20 with
# no delegation in between, injects a reminder to delegate the remainder or
# justify staying inline. A Task/Agent tool call resets the counter to 0.
#
# Subagent noise: PostToolUse may also fire for tool calls made by subagents
# (same session_id; agent_id semantics are undocumented as of Jul 2026).
# Mitigations: (1) the counter keys to the FIRST-SEEN agent_id — events
# reporting a different agent_id are ignored; (2) the reset on Task/Agent
# leaves the counter at 0 after every delegation, so any residual noise
# cannot accumulate across delegations.
#
# Advisory only — it NEVER blocks and never fails the tool call (exit 0 always).

set -uo pipefail

input=$(cat)
session_id=$(printf '%s' "$input" | jq -r '.session_id // empty' 2>/dev/null)
tool_name=$(printf '%s' "$input" | jq -r '.tool_name // empty' 2>/dev/null)
agent_id=$(printf '%s' "$input" | jq -r '.agent_id // "main"' 2>/dev/null)

[ -n "$session_id" ] || exit 0
[ -n "$tool_name" ] || exit 0

marker="${TMPDIR:-/tmp}/claude-delegation-reminder-${session_id}"

# Delegation observed -> reset the counter and stay silent.
case "$tool_name" in
  Task|Agent)
    printf '%s|0' "$agent_id" > "$marker" 2>/dev/null || true
    exit 0 ;;
esac

owner="$agent_id"
count=0
if [ -f "$marker" ]; then
  IFS='|' read -r owner count < "$marker" 2>/dev/null || true
  owner=${owner:-$agent_id}
  count=${count:-0}
  # Ignore tool calls attributed to agents other than the first-seen one.
  [ "$agent_id" = "$owner" ] || exit 0
fi

case "$count" in *[!0-9]*|'') count=0 ;; esac
count=$((count + 1))
printf '%s|%s' "$owner" "$count" > "$marker" 2>/dev/null || true

# Remind exactly when crossing each multiple of 20.
if [ "$count" -ge 20 ] && [ $((count % 20)) -eq 0 ]; then
  # Inject the SIGNAL only; `agent-routing.md > Delegation Gates` owns what to do about it.
  reminder="Delegation gate (non-blocking): ~${count} main-thread tool calls since the last delegation — see agent-routing.md > Delegation Gates."
  jq -n --arg ctx "$reminder" '{
    hookSpecificOutput: {
      hookEventName: "PostToolUse",
      additionalContext: $ctx
    }
  }'
fi

exit 0
