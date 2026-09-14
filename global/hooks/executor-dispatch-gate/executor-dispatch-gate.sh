#!/usr/bin/env bash
# executor-dispatch-gate.sh — PreToolUse on `Agent|Task`, flow workspaces only.
#
# Deterministic backstop for a rule that is otherwise prompt-convention: in a
# flow workspace (a `_support/PROJECT.md` ledger above cwd), dispatching an
# EXECUTOR subagent — one whose `tools:` allowlist carries Write or Edit — means
# implementing, and implementing needs a plan with implementation authority or
# the user's explicit verb in this conversation. The main thread has read a
# git-mode answer or a design decision as if it were that approval and
# dispatched anyway; this hook fires at the moment of the act.
#
#   No ledger above cwd            → exit 0, silent (direct route, untouched).
#   Non-executor subagent          → exit 0, silent.
#   A plan with can_implement=true → exit 0, silent (authority exists).
#   Otherwise, ledger row
#     `| Executor dispatch | plan-required |` → DENY (exit 2 + decision JSON,
#                                              the bash-policy shape).
#     anything else                          → ADVISORY via additionalContext
#                                              (the rule-context shape), once per
#                                              session and per plan-state
#                                              signature (marker file).
#
# Never fails loud: any internal error exits 0 with no output — a backstop
# must not break a dispatch through a bug of its own. Timeout is short.
#
# Dual-runtime: Claude Code (tool_input.subagent_type, snake_case fields) and
# Grok (toolInput.subagentType, camelCase). Codex/opencode/PI do not read
# settings.json — out of reach by construction (README).

set -uo pipefail

# Executor roster — agents under global/agents/**/*.md whose frontmatter
# `tools:` line lists Write or Edit, MINUS the exclusions below. Derived from
# disk on 2026-09-14; the test (test-executor-dispatch-gate.sh, case h)
# re-derives "Write/Edit roster minus EXCLUDED" and fails on drift, so a roster
# change surfaces in test, not in production. Keep both lists sorted.
#
# EXCLUDED: agents that write the artifacts the plan gate CONSUMES (specs,
# ADRs, contracts, infra design). By design they are dispatched BEFORE a plan
# exists (spec-writing playbook → business gate → /flow-plan), so gating them
# on a plan would block the path that produces the plan.
readonly EXCLUDED="cloud-architect sdd-design sdd-spec-writer solution-architect"
readonly EXECUTORS="angular-developer backend-developer database-specialist devops-engineer kotlin-multiplatform-developer performance-engineer prompt-engineer react-developer secrets-auditor test-engineer visual-designer"

# `HIVE_PLAN_PY` overrides the validator path (tests point it at the repo copy).
readonly PLAN_PY="${HIVE_PLAN_PY:-$HOME/.claude/skills/flow-core/scripts/plan.py}"

main() {
  local input subagent cwd session_id dir ledger root sessions_home
  local plans n authority plan st insp ci sig marker rel dispatch reason how

  input=$(cat)

  subagent=$(printf '%s' "$input" | jq -r '
    .tool_input.subagent_type // .toolInput.subagentType //
    .toolInput.subagent_type // empty' 2>/dev/null)
  [ -n "$subagent" ] || return 0

  case " ${EXCLUDED} " in
    *" ${subagent} "*) return 0 ;;   # pre-plan artifact writer: never gated
  esac
  case " ${EXECUTORS} " in
    *" ${subagent} "*) ;;
    *) return 0 ;;
  esac

  cwd=$(printf '%s' "$input" | jq -r '.cwd // .workspaceRoot // empty' 2>/dev/null)
  [ -n "$cwd" ] || cwd="$PWD"
  session_id=$(printf '%s' "$input" | jq -r '.session_id // .sessionId // empty' 2>/dev/null)

  # Flow workspace? Walk up for the ledger (same walk as flow-context.sh).
  dir="$cwd"; ledger=""
  while [ -n "$dir" ] && [ "$dir" != "/" ]; do
    if [ -f "$dir/_support/PROJECT.md" ]; then ledger="$dir/_support/PROJECT.md"; break; fi
    dir=$(dirname "$dir")
  done
  [ -n "$ledger" ] || return 0

  root=${ledger%/_support/PROJECT.md}
  sessions_home=""
  sessions_home=$(find "$root" -maxdepth 2 -type d -path '*-specs/sessions' 2>/dev/null | head -1)
  if [ -z "$sessions_home" ] && [ -d "$root/_support/sessions" ]; then
    sessions_home="$root/_support/sessions"
  fi

  # Any plan with implementation authority → silent. plan.py decides; when it
  # is not installed, the Status header is the approximation (and the message
  # says so). A malformed plan authorizes nothing.
  plans=""; n=0; authority=0; how="plan.py"
  [ -f "$PLAN_PY" ] || how="status-grep"
  if [ -n "$sessions_home" ]; then
    while IFS= read -r plan; do
      [ -n "$plan" ] || continue
      n=$((n + 1))
      st=$(grep -m1 -E '^Status:' "$plan" 2>/dev/null | sed -E 's/^Status:[[:space:]]*([a-z]+).*/\1/')
      ci="false"
      if [ "$how" = "plan.py" ]; then
        insp=$(python3 "$PLAN_PY" inspect "$plan" 2>/dev/null) || insp=""
        ci=$(printf '%s' "$insp" | jq -r '.can_implement // false' 2>/dev/null) || ci="false"
      else
        case "$st" in planned|building) ci="true" ;; esac
      fi
      [ "$ci" = "true" ] && authority=1
      plans="${plans}"$'\n'"${plan#"$root"/}|${st}|${ci}"
    done < <(find "$sessions_home" -maxdepth 2 -type f -name '*-plan.md' 2>/dev/null | sort)
  fi
  [ "$authority" -eq 0 ] || return 0

  # Ledger declaration: `| Executor dispatch | <value> |` in the header table.
  # The value cell may carry comments; only its leading token decides.
  dispatch=$(grep -m1 -iE '^\|[[:space:]]*Executor dispatch[[:space:]]*\|' "$ledger" 2>/dev/null \
    | awk -F'|' '{print $3}' | sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//')

  if [ -n "$sessions_home" ]; then
    rel="${sessions_home#"$root"/}"
    reason="executor-dispatch-gate: no plan with implementation authority in this flow workspace (${n} plan(s) checked under ${rel})."
  else
    reason="executor-dispatch-gate: no plan with implementation authority in this flow workspace (no sessions folder found under ${root}; 0 plans checked)."
  fi
  [ "$how" = "status-grep" ] && reason="${reason} plan.py was not found at ${PLAN_PY}, so a Status header of planned/building stood in for its can_implement check."
  reason="${reason} Dispatching ${subagent} means implementing. That needs either an approved /flow-plan (Status: planned with an implement grant) or the user's explicit implementation verb in THIS conversation (\"hazlo\", \"implementa\", \"aplica\"). A git-mode answer, a design-decision answer, or \"los atacaremos\" is not that verb."

  case "$dispatch" in
    plan-required*)
      printf 'DENY\n%s\n' "${reason} The ledger declares Executor dispatch: plan-required — run /flow-plan and get it approved, or the user edits that declaration."
      return 0
      ;;
  esac

  # Advisory: once per session and per plan-state signature.
  sig=$(printf '%s' "$plans" | cksum | cut -d' ' -f1)
  if [ -n "$session_id" ]; then
    marker="${TMPDIR:-/tmp}/claude-executor-dispatch-gate-${session_id}"
    if [ -f "$marker" ] && [ "$(cat "$marker" 2>/dev/null)" = "$sig" ]; then
      return 0
    fi
    printf '%s' "$sig" > "$marker" 2>/dev/null || true
  fi
  printf 'ADVISE\n%s\n' "${reason} If the verb was given, proceed and quote it in the dispatch prompt; otherwise offer /flow-plan or ask \"¿lo aplico?\" first."
  return 0
}

# Fail-open wrapper: the decision travels as `VERB\nmessage` on stdout; any
# error inside main (stderr swallowed, non-zero rc) becomes a silent exit 0.
out=$(main 2>/dev/null) || exit 0
[ -n "$out" ] || exit 0
verb=${out%%$'\n'*}
msg=${out#*$'\n'}

case "$verb" in
  DENY)
    # Shape copied from bash-policy.sh `deny()`: Claude honors exit 2 + stderr;
    # Grok honors {"decision":"deny"} and/or exit 2. Always print both.
    printf '%s\n' "$msg" >&2
    jq -n --arg r "$msg" '{decision: "deny", reason: $r}' 2>/dev/null || true
    exit 2
    ;;
  ADVISE)
    # Shape copied from rule-context.sh: PreToolUse advisory via
    # hookSpecificOutput.additionalContext; never blocks.
    jq -n --arg ctx "$msg" '{
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        additionalContext: $ctx
      }
    }' 2>/dev/null || true
    exit 0
    ;;
esac
exit 0
