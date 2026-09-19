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
#   0. Delegation counter        (all tools)  — agent-routing.md > Delegation Gates
#   1. Full-suite run counter    (shell)      — testing.md > Execution Scope
#   4. zsh-signature teacher     (shell)      — CLAUDE.md > Shell
#   5. Git-mode ask advisory     (edit tools) — git-mechanics.md > Commits
#   6. Remote-apply counter      (shell)      — debugging.md > remote-apply breaker
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

# ---------------------------------------------------------------------------
# Shared: subagent identity test. A non-empty result means this payload was
# emitted by a child agent, never the main thread. Same key set
# rule-delivery.py reads — Claude sets `agent_id`/`agentId`/`agent_type`/
# `agentType` only inside a child, Grok `subagentType`, `agent_name`
# generically; all are absent on the main thread.
# ---------------------------------------------------------------------------
subagent_identity() {
  printf '%s' "$input" | jq -r '
    [.agent_id, .agentId, .agent_type, .agentType, .subagentType, .agent_name]
    | map(select(type == "string" and . != "")) | first // empty' 2>/dev/null
}

# ---------------------------------------------------------------------------
# 0. Delegation counter — every non-delegation MAIN-THREAD tool call
# increments; a main-thread Task/Agent/spawn_subagent call resets to 0; a
# reminder fires on each multiple of 20.
#
# Subagent noise: PostToolUse also fires for tool calls made by subagents
# (same session_id). "Main-thread tool calls since the last delegation" is
# the gate's own definition, so a child's calls must not count toward it and
# must never surface the advisory to a child that has no delegation gate of
# its own — a payload carrying `subagent_identity` returns before the marker
# is read, written, or reset, whether or not the call itself is a further
# (nested) delegation.
# ---------------------------------------------------------------------------
delegation_section() {
  local marker count
  [ -z "$(subagent_identity)" ] || return 0
  marker="${TMPDIR:-/tmp}/claude-delegation-reminder-${session_id}"

  # Main-thread delegation observed -> reset the counter and stay silent.
  case "$tool_name" in
    Task|Agent|spawn_subagent|subagent)
      printf 'main|0' > "$marker" 2>/dev/null || true
      return 0 ;;
  esac

  count=0
  if [ -f "$marker" ]; then
    IFS='|' read -r _ count < "$marker" 2>/dev/null || true
    count=${count:-0}
  fi

  case "$count" in *[!0-9]*|'') count=0 ;; esac
  count=$((count + 1))
  printf 'main|%s' "$count" > "$marker" 2>/dev/null || true

  # Remind exactly when crossing each multiple of 20.
  if [ "$count" -ge 20 ] && [ $((count % 20)) -eq 0 ]; then
    # Inject the SIGNAL only; `agent-routing.md > Delegation Gates` owns what to do about it.
    printf 'Delegation gate (non-blocking): ~%s main-thread tool calls since the last delegation — see agent-routing.md > Delegation Gates.' "$count"
    # From the third firing on, the pointer alone has failed twice (often because the rule
    # left context at a compaction): name the act, still non-blocking.
    if [ "$count" -ge 60 ]; then
      printf ' The rule may be out of context: an investigation or research in progress routes to sdd-explore, a state sweep to state-fetcher; delegate the remainder now instead of the next call.'
    fi
  fi
}

# ---------------------------------------------------------------------------
# 1. Full-suite run counter — counts repeated whole-suite verification runs.
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
#
# MAIN THREAD ONLY. A subagent never owns the session's git mode: it cannot ask
# the question, its Write is not the session's first edit-intent, and the
# advisory reaches the user only as relayed noise. Uses the shared
# `subagent_identity` test (section 0's header documents the key set). The
# check runs before the markers so a child's write neither records the ask
# nor consumes the main thread's once-per-session evaluation.
# ---------------------------------------------------------------------------
git_mode_section() {
  [ "$harness" != "pi" ] || return 0
  local askq_marker marker cwd
  [ -z "$(subagent_identity)" ] || return 0
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
