#!/usr/bin/env bash
# verification-loop-reminder.sh — PostToolUse (Bash matcher) hook, NON-BLOCKING.
#
# Deterministic backstop against verification-loop runaway: re-running the FULL
# test suite to accumulate confidence instead of re-running the affected subset
# plus the failed check. Counts full-suite invocations per session and, from the
# second one onward, injects a reminder naming the run number.
#
# Detection is deliberately conservative — a command carrying `--filter` is the
# sanctioned affected-subset run and never counts, and anything followed by a
# non-flag argument (a spec path, a `test:unit` script) is treated as scoped.
# False negatives are acceptable here; false positives are not.
#
# rules/quality/testing.md > Execution Scope owns the policy; this only signals.
# Advisory only — it NEVER blocks and never fails the tool call (exit 0 always).

set -uo pipefail

# Full-suite classification. Matches `turbo run test` and `pnpm [-r] [run] test`
# when followed only by flags — never when `--filter` scopes the run.
is_full_suite() {
  local cmd="$1"

  [[ "$cmd" == *"--filter"* ]] && return 1

  local turbo_re='(^|[[:space:]]|[;&|])turbo[[:space:]]+run[[:space:]]+test([[:space:]]+-[^[:space:]]+)*[[:space:]]*($|[;&|])'
  local pnpm_re='(^|[[:space:]]|[;&|])pnpm[[:space:]]+(-r[[:space:]]+)?(run[[:space:]]+)?test([[:space:]]+-[^[:space:]]+)*[[:space:]]*($|[;&|])'

  [[ "$cmd" =~ $turbo_re ]] && return 0
  [[ "$cmd" =~ $pnpm_re ]] && return 0
  return 1
}

input=$(cat)
session_id=$(printf '%s' "$input" | jq -r '.session_id // empty' 2>/dev/null)
tool_name=$(printf '%s' "$input" | jq -r '.tool_name // empty' 2>/dev/null)
command=$(printf '%s' "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)

[ -n "$session_id" ] || exit 0
[ "$tool_name" = "Bash" ] || exit 0
[ -n "$command" ] || exit 0

is_full_suite "$command" || exit 0

marker="${TMPDIR:-/tmp}/claude-verification-loop-${session_id}"

count=0
[ -f "$marker" ] && read -r count < "$marker" 2>/dev/null
case "$count" in *[!0-9]*|'') count=0 ;; esac
count=$((count + 1))
printf '%s' "$count" > "$marker" 2>/dev/null || true

# The first full-suite run is the legitimate merge-boundary gate; only repeats signal.
if [ "$count" -ge 2 ]; then
  # Inject the SIGNAL only; `testing.md > Execution Scope` owns what to do about it.
  reminder="Full-suite run #${count} this session (non-blocking): a green full gate is not re-earned — re-run the affected subset plus the failed check; repeating a green check needs a named nondeterminism trigger, scoped to the flaky unit, never the full gate. See testing.md > Execution Scope."
  jq -n --arg ctx "$reminder" '{
    hookSpecificOutput: {
      hookEventName: "PostToolUse",
      additionalContext: $ctx
    }
  }'
fi

exit 0
