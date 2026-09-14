#!/usr/bin/env bash
# Checks for executor-dispatch-gate.sh. Fixtures live in a tmpdir: a ledger,
# plans rendered with the validator's own fixture (test_plan.render_plan), and
# TMPDIR redirected so session markers stay isolated per run.
# Run from anywhere: bash global/hooks/executor-dispatch-gate/test-executor-dispatch-gate.sh
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$ROOT/executor-dispatch-gate.sh"
REPO="$(cd "$ROOT/../../.." && pwd)"
AGENTS_DIR="$REPO/global/agents"
PLAN_SCRIPTS="$REPO/global/skills/flow-core/scripts"
fail=0

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
export TMPDIR="$WORK/tmp"
mkdir -p "$TMPDIR"
export HIVE_PLAN_PY="$PLAN_SCRIPTS/plan.py"

# ─── fixtures ────────────────────────────────────────────────────────────────
make_workspace() {
  # make_workspace <name> <dispatch-value|-> → prints the workspace root
  local name="$1" dispatch="$2" ws
  ws="$WORK/$name"
  mkdir -p "$ws/_support/sessions" "$ws/src"
  {
    printf '# %s — Project Ledger\n\n| | |\n|---|---|\n' "$name"
    printf '| Group / client | fixture |\n| Tracker | none |\n'
    [ "$dispatch" != "-" ] && printf '| Executor dispatch | %s |\n' "$dispatch"
    printf '| Last updated | 2026-09-14 |\n'
  } > "$ws/_support/PROJECT.md"
  printf '%s' "$ws"
}

render_plan() {
  # render_plan <sessions-dir> <slug> <status> <grants: implement|none>
  local sessions="$1" slug="$2" status="$3" grants="$4" folder
  folder="$sessions/$slug"
  mkdir -p "$folder"
  python3 - "$PLAN_SCRIPTS" "$folder" "$status" "$grants" <<'PY'
import importlib.util, sys
from pathlib import Path
scripts, folder, status, grants = sys.argv[1:5]
spec = importlib.util.spec_from_file_location("test_plan", Path(scripts) / "test_plan.py")
mod = importlib.util.module_from_spec(spec)
sys.modules[spec.name] = mod
spec.loader.exec_module(mod)
kwargs = {"status": status}
if grants == "none":
    kwargs["grants"] = []
path = mod.render_plan(Path(folder), **kwargs)
path.rename(Path(folder) / f"{Path(folder).name}-plan.md")
PY
}

# ─── runner ──────────────────────────────────────────────────────────────────
run_case() {
  # run_case <name> <expect: silent|advisory|deny> <payload-json>
  local name="$1" expect="$2" payload="$3"
  local out err rc="" got
  err=$(mktemp)
  out=$(printf '%s' "$payload" | bash "$SCRIPT" 2>"$err") || rc=$?
  rc=${rc:-0}
  if [ "$rc" -eq 2 ] && printf '%s' "$out" | jq -e '.decision == "deny" and (.reason | length > 0)' >/dev/null 2>&1 \
     && [ -s "$err" ]; then
    got="deny"
  elif [ "$rc" -eq 0 ] && printf '%s' "$out" | jq -e '.hookSpecificOutput.hookEventName == "PreToolUse" and (.hookSpecificOutput.additionalContext | length > 0)' >/dev/null 2>&1; then
    got="advisory"
  elif [ "$rc" -eq 0 ] && [ -z "$out" ] && [ ! -s "$err" ]; then
    got="silent"
  else
    got="unexpected(rc=$rc)"
  fi
  rm -f "$err"
  LAST_OUT="$out"
  if [ "$got" != "$expect" ]; then
    printf 'FAIL %s: expected %s, got %s\n%s\n' "$name" "$expect" "$got" "$out"
    fail=$((fail + 1))
    return
  fi
  printf 'ok   %s (%s)\n' "$name" "$got"
}

claude_payload() {
  # claude_payload <subagent> <cwd> <session>
  jq -n --arg a "$1" --arg c "$2" --arg s "$3" \
    '{tool_name: "Agent", tool_input: {subagent_type: $a, prompt: "do it"}, cwd: $c, session_id: $s}'
}
grok_payload() {
  jq -n --arg a "$1" --arg c "$2" --arg s "$3" \
    '{toolName: "Agent", toolInput: {subagentType: $a, prompt: "do it"}, workspaceRoot: $c, sessionId: $s}'
}

# ─── (a) no ledger → silent ──────────────────────────────────────────────────
mkdir -p "$WORK/plain/src"
run_case "a-no-ledger" silent "$(claude_payload backend-developer "$WORK/plain/src" s-a)"

# ─── (b) ledger, no plans → advisory ─────────────────────────────────────────
WS_B=$(make_workspace ws-b -)
run_case "b-ledger-no-plans" advisory "$(claude_payload backend-developer "$WS_B/src" s-b)"
case "$LAST_OUT" in
  *"0 plan(s) checked under _support/sessions"*) printf 'ok   b-message-names-count\n' ;;
  *) printf 'FAIL b-message-names-count: %s\n' "$LAST_OUT"; fail=$((fail + 1)) ;;
esac

# ─── (c) ledger + plan with can_implement → silent ───────────────────────────
WS_C=$(make_workspace ws-c -)
render_plan "$WS_C/_support/sessions" "2026-09-14-feature" planned implement
ci=$(python3 "$HIVE_PLAN_PY" inspect "$WS_C/_support/sessions/2026-09-14-feature/2026-09-14-feature-plan.md" | jq -r .can_implement)
[ "$ci" = "true" ] || { printf 'FAIL c-fixture: can_implement=%s\n' "$ci"; fail=$((fail + 1)); }
run_case "c-plan-with-authority" silent "$(claude_payload backend-developer "$WS_C/src" s-c)"

# ─── (c2) draft plan / planned without grant → advisory ──────────────────────
WS_C2=$(make_workspace ws-c2 -)
render_plan "$WS_C2/_support/sessions" "2026-09-14-draft" draft implement
render_plan "$WS_C2/_support/sessions" "2026-09-14-nogrant" planned none
run_case "c2-plans-without-authority" advisory "$(claude_payload test-engineer "$WS_C2/src" s-c2)"
case "$LAST_OUT" in
  *"2 plan(s) checked"*) printf 'ok   c2-message-counts-both\n' ;;
  *) printf 'FAIL c2-message-counts-both: %s\n' "$LAST_OUT"; fail=$((fail + 1)) ;;
esac

# ─── (d) ledger plan-required + no plan → deny ───────────────────────────────
WS_D=$(make_workspace ws-d "plan-required — set 2026-09-14")
run_case "d-plan-required-deny" deny "$(claude_payload backend-developer "$WS_D/src" s-d)"
case "$LAST_OUT" in
  *"Executor dispatch: plan-required"*) printf 'ok   d-deny-names-declaration\n' ;;
  *) printf 'FAIL d-deny-names-declaration: %s\n' "$LAST_OUT"; fail=$((fail + 1)) ;;
esac
# deny is not marker-gated: a second dispatch denies again
run_case "d2-plan-required-deny-again" deny "$(claude_payload backend-developer "$WS_D/src" s-d)"
# plan-required + plan with authority → silent (authority wins over the declaration)
WS_D3=$(make_workspace ws-d3 "plan-required")
render_plan "$WS_D3/_support/sessions" "2026-09-14-ok" building implement
run_case "d3-plan-required-with-authority" silent "$(claude_payload backend-developer "$WS_D3/src" s-d3)"
# the template's default value starts with `free` → advisory, not deny
WS_D4=$(make_workspace ws-d4 "free (default) / plan-required — comment")
run_case "d4-free-default-advisory" advisory "$(claude_payload backend-developer "$WS_D4/src" s-d4)"

# ─── (e) non-executor agent → silent ─────────────────────────────────────────
for ro in sdd-explore review-code review-refuter state-fetcher workspace-custodian sdd-verify review-ux Explore sdd-spec-writer sdd-design cloud-architect solution-architect; do
  run_case "e-non-executor-$ro" silent "$(claude_payload "$ro" "$WS_B/src" "s-e-$ro")"
done
run_case "e-empty-subagent" silent "$(jq -n --arg c "$WS_B/src" '{tool_name: "Agent", tool_input: {prompt: "x"}, cwd: $c, session_id: "s-e0"}')"

# ─── (f) Grok camelCase payload → same as (b) ────────────────────────────────
run_case "f-grok-camelcase" advisory "$(grok_payload backend-developer "$WS_B/src" s-f)"

# ─── (g) same session, same state → silent; state change → re-fires ─────────
run_case "g-first-dispatch" advisory "$(claude_payload backend-developer "$WS_B/src" s-g)"
run_case "g-second-dispatch-same-state" silent "$(claude_payload react-developer "$WS_B/src" s-g)"
render_plan "$WS_B/_support/sessions" "2026-09-14-late" draft implement
run_case "g2-state-changed-refires" advisory "$(claude_payload react-developer "$WS_B/src" s-g)"
run_case "g3-no-session-id-always-fires" advisory "$(jq -n --arg c "$WS_B/src" '{tool_name: "Agent", tool_input: {subagent_type: "backend-developer"}, cwd: $c}')"

# ─── (i) plan.py missing → Status-grep fallback, message says so ─────────────
WS_I=$(make_workspace ws-i -)
render_plan "$WS_I/_support/sessions" "2026-09-14-nogrant" planned none
HIVE_PLAN_PY=/nonexistent/plan.py run_case "i-fallback-authority-by-status" silent "$(claude_payload backend-developer "$WS_I/src" s-i)"
WS_I2=$(make_workspace ws-i2 -)
render_plan "$WS_I2/_support/sessions" "2026-09-14-draft" draft implement
HIVE_PLAN_PY=/nonexistent/plan.py run_case "i2-fallback-advisory-names-itself" advisory "$(claude_payload backend-developer "$WS_I2/src" s-i2)"
case "$LAST_OUT" in
  *"plan.py was not found"*) printf 'ok   i2-fallback-message\n' ;;
  *) printf 'FAIL i2-fallback-message: %s\n' "$LAST_OUT"; fail=$((fail + 1)) ;;
esac

# ─── (j) malformed plan authorizes nothing, and never breaks the hook ────────
WS_J=$(make_workspace ws-j -)
mkdir -p "$WS_J/_support/sessions/broken"
printf 'Status: planned\n\nno markers at all\n' > "$WS_J/_support/sessions/broken/broken-plan.md"
run_case "j-malformed-plan-advisory" advisory "$(claude_payload backend-developer "$WS_J/src" s-j)"

# ─── (k) specs-repo sessions home is found before _support/sessions ──────────
WS_K=$(make_workspace ws-k -)
mkdir -p "$WS_K/fixture-specs/sessions"
render_plan "$WS_K/fixture-specs/sessions" "2026-09-14-ok" planned implement
run_case "k-specs-sessions-home" silent "$(claude_payload backend-developer "$WS_K/src" s-k)"

# ─── (h) executor list = Write/Edit roster on disk minus the declared exclusions
declared=$(grep -m1 -E '^readonly EXECUTORS=' "$SCRIPT" | sed -E 's/^readonly EXECUTORS="(.*)"$/\1/' | tr ' ' '\n' | sort)
excluded=$(grep -m1 -E '^readonly EXCLUDED=' "$SCRIPT" | sed -E 's/^readonly EXCLUDED="(.*)"$/\1/' | tr ' ' '\n' | sort)
writers=$(for f in "$AGENTS_DIR"/*/*.md; do
  tools=$(grep -m1 -E '^tools:' "$f" | sed -E 's/^tools:[[:space:]]*//')
  case ",${tools// /}," in
    *,Write,*|*,Edit,*) grep -m1 -E '^name:' "$f" | sed -E 's/^name:[[:space:]]*//' ;;
  esac
done | sort)
actual=$(comm -23 <(printf '%s\n' "$writers") <(printf '%s\n' "$excluded"))
if [ -z "$writers" ]; then
  printf 'FAIL h-roster: no Write/Edit agents found under %s\n' "$AGENTS_DIR"
  fail=$((fail + 1))
elif [ "$declared" = "$actual" ]; then
  printf 'ok   h-executor-list-matches-roster-minus-exclusions (%s of %s Write/Edit agents)\n' \
    "$(printf '%s\n' "$actual" | grep -c .)" "$(printf '%s\n' "$writers" | grep -c .)"
else
  printf 'FAIL h-executor-list-matches-roster-minus-exclusions\n--- declared in script\n%s\n--- derived from %s minus EXCLUDED\n%s\n' "$declared" "$AGENTS_DIR" "$actual"
  fail=$((fail + 1))
fi
# every exclusion must still name a Write/Edit agent on disk — a stale entry is drift too
stale=$(comm -13 <(printf '%s\n' "$writers") <(printf '%s\n' "$excluded"))
if [ -n "$stale" ]; then
  printf 'FAIL h-exclusions-stale: not Write/Edit agents on disk: %s\n' "$(printf '%s' "$stale" | tr '\n' ' ')"
  fail=$((fail + 1))
else
  printf 'ok   h-exclusions-all-on-disk (%s)\n' "$(printf '%s' "$excluded" | tr '\n' ' ')"
fi

# ─── structural: every deny names its executable continuation ────────────────
# shellcheck disable=SC2016 # literal backticks/phrases are the pattern, not expansion
cont_re='run /flow-plan|have the user|ask the user|the user edits'
deny_count=0
while IFS= read -r msg; do
  [ -n "$msg" ] || continue
  deny_count=$((deny_count + 1))
  if printf '%s' "$msg" | grep -qE "$cont_re"; then
    printf 'ok   deny-continuation %d\n' "$deny_count"
  else
    printf 'FAIL deny-continuation missing: %s\n' "$msg"
    fail=$((fail + 1))
  fi
done < <(grep -E "printf 'DENY" "$SCRIPT")
if [ "$deny_count" -eq 0 ]; then
  printf 'FAIL deny-continuation: no deny messages extracted from %s\n' "$SCRIPT"
  fail=$((fail + 1))
fi

if [ "$fail" -ne 0 ]; then
  printf '\n%d case(s) failed\n' "$fail"
  exit 1
fi
printf '\nall executor-dispatch-gate cases passed\n'
