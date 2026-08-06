#!/usr/bin/env bash
# post-tool-hub.sh — PostToolUse (all tools) hub, NON-BLOCKING.
#
# One process, one stdin read, three independent sections. Each section
# self-gates on tool_name and returns its reminder text; whatever fires is
# newline-joined into a SINGLE additionalContext emission.
#
#   1. Delegation counter        (all tools)  — agent-routing.md > Delegation Gates
#   2. Full-suite run counter    (Bash)       — testing.md > Execution Scope
#   3. Index anti-conclusion     (codegraph/jbcontext) — code-search.md
#
# The counters keep SEPARATE state files and are never coupled: one section
# firing must not reset or advance another's count.
#
# Advisory only — it NEVER blocks and never fails the tool call (exit 0 always).

set -uo pipefail

input=$(cat)
session_id=$(printf '%s' "$input" | jq -r '.session_id // empty' 2>/dev/null)
tool_name=$(printf '%s' "$input" | jq -r '.tool_name // empty' 2>/dev/null)

[ -n "$session_id" ] || exit 0
[ -n "$tool_name" ] || exit 0

# ---------------------------------------------------------------------------
# 1. Delegation counter — every non-delegation tool call increments; a
# Task/Agent call resets to 0; a reminder fires on each multiple of 20.
#
# Subagent noise: PostToolUse may also fire for tool calls made by subagents
# (same session_id; agent_id semantics are undocumented as of Jul 2026).
# Mitigations: (1) the counter keys to the FIRST-SEEN agent_id — events
# reporting a different agent_id are ignored; (2) the reset on Task/Agent
# leaves the counter at 0 after every delegation, so any residual noise
# cannot accumulate across delegations.
# ---------------------------------------------------------------------------
delegation_section() {
  local agent_id marker owner count
  agent_id=$(printf '%s' "$input" | jq -r '.agent_id // "main"' 2>/dev/null)
  marker="${TMPDIR:-/tmp}/claude-delegation-reminder-${session_id}"

  # Delegation observed -> reset the counter and stay silent.
  case "$tool_name" in
    Task|Agent)
      printf '%s|0' "$agent_id" > "$marker" 2>/dev/null || true
      return 0 ;;
  esac

  owner="$agent_id"
  count=0
  if [ -f "$marker" ]; then
    IFS='|' read -r owner count < "$marker" 2>/dev/null || true
    owner=${owner:-$agent_id}
    count=${count:-0}
    # Ignore tool calls attributed to agents other than the first-seen one.
    [ "$agent_id" = "$owner" ] || return 0
  fi

  case "$count" in *[!0-9]*|'') count=0 ;; esac
  count=$((count + 1))
  printf '%s|%s' "$owner" "$count" > "$marker" 2>/dev/null || true

  # Remind exactly when crossing each multiple of 20.
  if [ "$count" -ge 20 ] && [ $((count % 20)) -eq 0 ]; then
    # Inject the SIGNAL only; `agent-routing.md > Delegation Gates` owns what to do about it.
    printf 'Delegation gate (non-blocking): ~%s main-thread tool calls since the last delegation — see agent-routing.md > Delegation Gates.' "$count"
  fi
}

# ---------------------------------------------------------------------------
# 2. Full-suite run counter — counts repeated whole-suite verification runs.
#
# Detection is deliberately conservative: a command carrying `--filter` is the
# sanctioned affected-subset run and never counts, and anything followed by a
# non-flag argument (a spec path, a `test:unit` script) is treated as scoped.
# False negatives are acceptable here; false positives are not.
# ---------------------------------------------------------------------------
is_full_suite() {
  local cmd="$1"

  [[ "$cmd" == *"--filter"* ]] && return 1

  local turbo_re='(^|[[:space:]]|[;&|])turbo[[:space:]]+run[[:space:]]+test([[:space:]]+-[^[:space:]]+)*[[:space:]]*($|[;&|])'
  local pnpm_re='(^|[[:space:]]|[;&|])pnpm[[:space:]]+(-r[[:space:]]+)?(run[[:space:]]+)?test([[:space:]]+-[^[:space:]]+)*[[:space:]]*($|[;&|])'

  [[ "$cmd" =~ $turbo_re ]] && return 0
  [[ "$cmd" =~ $pnpm_re ]] && return 0
  return 1
}

verification_loop_section() {
  local command marker count
  [ "$tool_name" = "Bash" ] || return 0
  command=$(printf '%s' "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)
  [ -n "$command" ] || return 0
  is_full_suite "$command" || return 0

  marker="${TMPDIR:-/tmp}/claude-verification-loop-${session_id}"
  count=0
  [ -f "$marker" ] && read -r count < "$marker" 2>/dev/null
  case "$count" in *[!0-9]*|'') count=0 ;; esac
  count=$((count + 1))
  printf '%s' "$count" > "$marker" 2>/dev/null || true

  # The first full-suite run is the legitimate merge-boundary gate; only repeats signal.
  if [ "$count" -ge 2 ]; then
    # Inject the SIGNAL only; `testing.md > Execution Scope` owns what to do about it.
    printf 'Full-suite run #%s this session (non-blocking): a green full gate is not re-earned — re-run the affected subset plus the failed check; repeating a green check needs a named nondeterminism trigger, scoped to the flaky unit, never the full gate. See testing.md > Execution Scope.' "$count"
  fi
}

# ---------------------------------------------------------------------------
# 3. Index anti-conclusion discipline — injected right after an index tool ran.
# Once per session: the discipline is a stance, not a per-call correction, and
# repeating it every codegraph/jbcontext call is pure context tax.
# ---------------------------------------------------------------------------
index_discipline_section() {
  local tool="" marker
  case "$tool_name" in
    Bash)
      local cmd
      cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)
      if printf '%s' "$cmd" | grep -qE '(^|[^[:alnum:]_-])codegraph([^[:alnum:]_-]|$)'; then tool="codegraph"; fi
      if printf '%s' "$cmd" | grep -qE '(^|[^[:alnum:]_-])jbcontext([^[:alnum:]_-]|$)'; then tool="${tool:+$tool }jbcontext"; fi
      ;;
    mcp__codegraph__*) tool="codegraph" ;;
    mcp__jbcontext__*) tool="jbcontext" ;;
  esac
  [ -n "$tool" ] || return 0

  marker="${TMPDIR:-/tmp}/claude-index-discipline-${session_id}"
  [ -f "$marker" ] && return 0
  : > "$marker" 2>/dev/null || true

  printf '%s' 'code-search discipline (rules/tools/code-search.md): an index result is a pointer, never a verdict. Read the cited file before citing it; never conclude absence from an index — "does not exist" requires an exhaustive rg sweep (0 hits, broad vocabulary); check a hit is alive (callers/imports) before building on it; on index-vs-disk conflict, disk wins.'
}

# ---------------------------------------------------------------------------
# Collect and emit once.
# ---------------------------------------------------------------------------
context=""
newline=$'\n'
append() {
  [ -n "$1" ] || return 0
  context="${context:+$context$newline}$1"
}

append "$(delegation_section)"
append "$(verification_loop_section)"
append "$(index_discipline_section)"

if [ -n "$context" ]; then
  jq -n --arg ctx "$context" '{
    hookSpecificOutput: {
      hookEventName: "PostToolUse",
      additionalContext: $ctx
    }
  }'
fi

exit 0
