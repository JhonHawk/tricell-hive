#!/usr/bin/env bash
# post-tool-hub.sh — PostToolUse (all tools) hub, NON-BLOCKING.
#
# Claude Code, Grok Build, Cursor, and normalized PI payloads. Field fallbacks must stay
# in sync with bash-policy.sh.
#
# One process, one stdin read, five independent sections. Each section
# self-gates on tool name and returns its reminder text; whatever fires is
# newline-joined into a SINGLE additionalContext emission.
#
#   0. Plan-capture recovery     (first call) — delegates to flow-plan-capture.sh
#   1. Delegation counter        (all tools)  — agent-routing.md > Delegation Gates
#   2. Full-suite run counter    (shell)      — testing.md > Execution Scope
#   4. zsh-signature teacher     (shell)      — CLAUDE.md > Shell
#   5. Git-mode ask advisory     (edit tools) — git-mechanics.md > Commits
#   6. Remote-apply counter      (shell)      — debugging.md > remote-apply breaker
#
# Section 0 is the one that WRITES (via the capture hook it calls) rather than
# only advising. It lives here because it needs the earliest event that can see
# the fresh session's first user turn, and this hub already spawns on every tool
# call — a dedicated registration would double the per-call process cost to
# catch a once-per-session condition.
#
# The counters keep SEPARATE state files and are never coupled: one section
# firing must not reset or advance another's count.
#
# Advisory only — it NEVER blocks and never fails the tool call (exit 0 always).

set -uo pipefail

input=$(cat)
harness=$(printf '%s' "$input" | jq -r '.harness // empty' 2>/dev/null)
# Tri-runtime field normalization (Claude snake_case | Grok camelCase | Cursor).
session_id=$(printf '%s' "$input" | jq -r '.session_id // .sessionId // empty' 2>/dev/null)
tool_name=$(printf '%s' "$input" | jq -r '.tool_name // .toolName // empty' 2>/dev/null)
command=$(printf '%s' "$input" | jq -r '.tool_input.command // .toolInput.command // empty' 2>/dev/null)

[ -n "$session_id" ] || exit 0
[ -n "$tool_name" ] || exit 0

is_shell=0
case "$tool_name" in
  Bash|run_terminal_command|Shell) is_shell=1 ;;
esac

HOOK_DIR=$(cd "$(dirname "$0")" 2>/dev/null && pwd) || HOOK_DIR=""

# ---------------------------------------------------------------------------
# 0. Plan-capture recovery — approving a plan WITH context cleared resolves
# ExitPlanMode as a permission deny, so PostToolUse:ExitPlanMode never fires and
# flow-plan-capture never sees the plan. The fresh session's first user turn
# carries it in `planContent`; this fires the capture hook in recovery mode on
# the FIRST tool call of any session and stays quiet everywhere else.
#
# The marker is written BEFORE the attempt: a failure must not retry every call.
# ---------------------------------------------------------------------------
plan_recovery_section() {
  [ "$harness" != "pi" ] || return 0
  local marker out
  marker="${TMPDIR:-/tmp}/claude-flow-plan-recovery-${session_id}"
  [ -f "$marker" ] && return 0
  : > "$marker" 2>/dev/null || true
  [ -n "$HOOK_DIR" ] && [ -x "$HOOK_DIR/flow-plan-capture.sh" ] || return 0
  out=$(printf '%s' "$input" | "$HOOK_DIR/flow-plan-capture.sh" --from-transcript 2>/dev/null)
  [ -n "$out" ] && printf '%s' "$out"
}

# ---------------------------------------------------------------------------
# 1. Delegation counter — every non-delegation tool call increments; a
# Task/Agent/spawn_subagent call resets to 0; a reminder fires on each
# multiple of 20.
#
# Subagent noise: PostToolUse may also fire for tool calls made by subagents
# (same session_id; agent_id semantics are undocumented as of Jul 2026).
# Mitigations: (1) the counter keys to the FIRST-SEEN agent_id — events
# reporting a different agent_id are ignored; (2) the reset on delegation
# leaves the counter at 0 after every delegation, so any residual noise
# cannot accumulate across delegations.
# ---------------------------------------------------------------------------
delegation_section() {
  local agent_id marker owner count
  agent_id=$(printf '%s' "$input" | jq -r '.agent_id // .agentId // "main"' 2>/dev/null)
  marker="${TMPDIR:-/tmp}/claude-delegation-reminder-${session_id}"

  # Delegation observed -> reset the counter and stay silent.
  case "$tool_name" in
    Task|Agent|spawn_subagent|subagent)
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
  local marker count
  [ "$is_shell" -eq 1 ] || return 0
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
# 4. zsh-signature teacher — fires when a shell tool result carries a zsh
# dialect error. In Grok (zsh runtime) it names the exact fix so the retry is
# informed; in Claude (bash 5 expected via CLAUDE_CODE_SHELL) the same
# signature means the override drifted — that IS the alarm. Fires per failure:
# these signatures only appear when a command actually died on them.
# ---------------------------------------------------------------------------
zsh_signature_section() {
  local out sig="" re
  [ "$is_shell" -eq 1 ] || return 0
  # Collect every string in the response, whatever its shape (Claude's
  # {stdout,stderr,…}, Grok's own), joined by REAL newlines. `tostring` on the
  # object would re-escape them as the two characters `\` `n`, which defeats the
  # line anchor below: the character before `zsh:` would read as alphanumeric.
  out=$(printf '%s' "$input" | jq -r '
    (.tool_response // .toolResult // empty) as $r
    | if ($r | type) == "object" then [$r | .. | strings] | join("\n")
      else ($r | tostring) end' 2>/dev/null)
  [ -n "$out" ] || return 0

  # The phrase alone is NOT the signature — zsh's own error prefix is. Matching
  # the bare text fired on any tool output that merely CONTAINED it: a `cat` or
  # `git diff` of the rules that document these signatures was enough. Narrowing
  # the scanned field is not the fix — Claude's Bash tool merges the command's
  # stderr into `stdout`, so a real failure has no separate channel to read.
  # Real forms carry a line number: `zsh:<N>: read-only variable: status`,
  # `(eval):<N>: bad substitution`. Written with `<N>` on purpose — spelling the
  # digit here would make this very comment match, so any `cat` of this file
  # would fire the alarm. Same reason the README uses `zsh:N:`.
  re='(^|[^[:alnum:]_])(zsh|\(eval\)):[0-9]*:?[[:space:]]*(read-only variable|no matches found|bad substitution)'
  if [[ "$out" =~ $re ]]; then
    case "${BASH_REMATCH[3]}" in
      "no matches found") sig="no matches found (nomatch)" ;;
      *)                  sig="${BASH_REMATCH[3]}" ;;
    esac
  elif [[ "$out" =~ \(eval\):[0-9]+: ]]; then
    sig="(eval):N: error"
  else return 0; fi

  # shellcheck disable=SC2016  # single quotes intentional: ${!var}/${(P)var}/$BASH_VERSION are literal text for the model, never expanded here
  if [ "$harness" = "pi" ]; then
    printf 'zsh failure signature detected (%s). Inspect the configured shell and apply the shell-standards reference before retrying.' "$sig"
  elif [ "$tool_name" = "run_terminal_command" ]; then
    printf 'zsh failure signature detected (%s). This tool shell is zsh: `status`/`path` are special — rename such variables (st, repo_path); `${!var}` is `${(P)var}`; no `declare -A`/`mapfile`/`read -p`; unmatched globs abort, even as flag values. Fix per CLAUDE.md > Shell and retry.' "$sig"
  else
    printf 'zsh failure signature under the Bash tool (%s), which should be running bash 5 via CLAUDE_CODE_SHELL — the override may have drifted or a restart is pending. Verify with `echo $BASH_VERSION`; until it says 5.x, apply the zsh rules in CLAUDE.md > Shell.' "$sig"
  fi
}

# ---------------------------------------------------------------------------
# 5. Git-mode ask advisory — the session-mode question (git-mechanics.md >
# Commits) is mandatory at first edit-intent, and two real sessions skipped or
# shrank it. On the session's FIRST Write/Edit/MultiEdit inside a git repo, if
# no AskUserQuestion call was observed earlier, remind once.
#
# Claude-shaped payloads only (`tool_name`, snake_case): AskUserQuestion is
# Claude's ask tool; Grok's equivalent is not observably named here, so firing
# on its payloads would false-positive after a legitimate ask. Heuristic by
# design — a question asked where PostToolUse cannot see it is not counted.
# ---------------------------------------------------------------------------
git_mode_section() {
  [ "$harness" != "pi" ] || return 0
  local askq_marker marker cwd
  askq_marker="${TMPDIR:-/tmp}/claude-askq-seen-${session_id}"

  # Ask observed -> record it and stay silent.
  if [ "$tool_name" = "AskUserQuestion" ]; then
    : > "$askq_marker" 2>/dev/null || true
    return 0
  fi

  case "$tool_name" in
    Write|Edit|MultiEdit) ;;
    *) return 0 ;;
  esac
  # Claude payload shape only (see header note).
  printf '%s' "$input" | jq -e 'has("tool_name")' >/dev/null 2>&1 || return 0

  # Only an edit inside a git repo consumes the once-per-session evaluation.
  cwd=$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null)
  [ -n "$cwd" ] || return 0
  git -C "$cwd" rev-parse --is-inside-work-tree >/dev/null 2>&1 || return 0

  marker="${TMPDIR:-/tmp}/claude-git-mode-ask-${session_id}"
  [ -f "$marker" ] && return 0
  : > "$marker" 2>/dev/null || true
  [ -f "$askq_marker" ] && return 0

  printf 'Session git mode (non-blocking): the mandatory mode question (git-mechanics.md > Commits) does not appear to have been asked this session — ask it before the first commit if this session will touch git. Advisory only; heuristic — it only counts AskUserQuestion calls this hook observed.'
}

# ---------------------------------------------------------------------------
# 6. Remote-apply counter — counts remote apply/deploy/migration ATTEMPTS,
# never failures. Counting failures would mean parsing output, and one failed
# run gets polled repeatedly while it is investigated: the count would inflate
# on a single failure, which is the false positive this hub refuses. An attempt
# is unambiguous from the command alone, and the breaker caps attempts anyway.
#
# Fires from the THIRD, not the second: `debugging.md` ends remote execution at
# the second FAILED attempt, and this hook cannot tell success from failure.
# Same shape as section 1 — inject the SIGNAL, let the rule decide.
# ---------------------------------------------------------------------------
is_remote_apply() {
  local cmd="$1" sep='(^|[[:space:]]|[;&|])'

  [[ "$cmd" =~ ${sep}terraform([[:space:]]+-chdir=[^[:space:]]+)*[[:space:]]+apply ]] && return 0
  [[ "$cmd" =~ ${sep}gh[[:space:]]+workflow[[:space:]]+run ]] && return 0
  [[ "$cmd" =~ ${sep}gh[[:space:]]+run[[:space:]]+rerun ]] && return 0
  [[ "$cmd" =~ ${sep}aws[[:space:]]+ecs[[:space:]]+update-service ]] && return 0
  [[ "$cmd" =~ ${sep}(vercel|flyctl|fly|dokploy)[[:space:]]+deploy ]] && return 0
  # A push whose refspec NAMES an environment branch. A bare `git push` from an
  # already-checked-out env branch is invisible here and stays a false negative,
  # which this hub accepts; a false positive it does not.
  [[ "$cmd" =~ ${sep}git[[:space:]]+push([[:space:]]+-[^[:space:]]+)*[[:space:]]+[^[:space:]]+[[:space:]]+(qa|staging|stage|prod|production)([[:space:]]|$) ]] && return 0
  return 1
}

remote_apply_section() {
  local marker count
  [ "$is_shell" -eq 1 ] || return 0
  [ -n "$command" ] || return 0
  is_remote_apply "$command" || return 0

  marker="${TMPDIR:-/tmp}/claude-remote-apply-${session_id}"
  count=0
  [ -f "$marker" ] && read -r count < "$marker" 2>/dev/null
  case "$count" in *[!0-9]*|"") count=0 ;; esac
  count=$((count + 1))
  printf '%s' "$count" > "$marker" 2>/dev/null || true

  if [ "$count" -ge 3 ]; then
    # Inject the SIGNAL only; `debugging.md` owns what to do about it.
    printf 'Remote apply/deploy/migration attempt #%s this session (non-blocking): the SECOND FAILED such attempt within a milestone ends remote execution for the session. If two of these failed, stop — checkpoint and re-plan from offline evidence (enumerate action×resource×context, versions, the exact plan) before attempting again; a new failure class does not reset the count. The preventive half is the observation path in devops-principles.md. See debugging.md.' "$count"
  fi
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

append "$(plan_recovery_section)"
append "$(delegation_section)"
append "$(verification_loop_section)"
append "$(zsh_signature_section)"
append "$(git_mode_section)"
append "$(remote_apply_section)"

if [ -n "$context" ]; then
  jq -n --arg ctx "$context" '{
    hookSpecificOutput: {
      hookEventName: "PostToolUse",
      additionalContext: $ctx
    }
  }'
fi

exit 0
