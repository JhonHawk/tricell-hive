#!/usr/bin/env bash
# Lightweight dual-runtime checks for bash-policy.sh (not deployed).
# Run from anywhere: bash global/hooks/bash-policy/test-dual-runtime.sh
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$ROOT/bash-policy.sh"
fail=0

run_case() {
  local name="$1" expect_exit="$2" payload="$3"
  local out rc
  out=$(printf '%s' "$payload" | bash "$SCRIPT" 2>&1) || rc=$?
  rc=${rc:-0}
  if [ "$rc" -ne "$expect_exit" ]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$name" "$expect_exit" "$rc" "$out"
    fail=$((fail + 1))
    return
  fi
  if [ "$expect_exit" -eq 2 ]; then
    if ! printf '%s' "$out" | grep -q '"decision": "deny"\|"decision":"deny"'; then
      printf 'FAIL %s: deny JSON missing on stdout\n%s\n' "$name" "$out"
      fail=$((fail + 1))
      return
    fi
  fi
  printf 'ok   %s (exit %s)\n' "$name" "$rc"
}

# Claude-shaped pip ban
run_case "claude-pip" 2 '{
  "tool_name": "Bash",
  "tool_input": {"command": "pip install requests"},
  "cwd": "/tmp"
}'

# Grok-shaped pip ban
run_case "grok-pip" 2 '{
  "toolName": "run_terminal_command",
  "toolInput": {"command": "pip install requests"},
  "cwd": "/tmp"
}'

# Grok-shaped safe command
run_case "grok-pnpm-test" 0 '{
  "toolName": "run_terminal_command",
  "toolInput": {"command": "pnpm test --filter pkg"},
  "cwd": "/tmp"
}'

# Claude-shaped safe command
run_case "claude-ls" 0 '{
  "tool_name": "Bash",
  "tool_input": {"command": "ls -la"},
  "cwd": "/tmp"
}'

# --- (b2) zsh special-name deny (Grok scope) — heredoc bodies are not shell --
run_case "grok-path-assign-plain" 2 '{
  "toolName": "run_terminal_command",
  "toolInput": {"command": "path=/x"},
  "cwd": "/tmp"
}'

run_case "grok-path-in-python-heredoc" 0 '{
  "toolName": "run_terminal_command",
  "toolInput": {"command": "python3 <<'\''PY'\''\npath=\"/x\"\nprint(path)\nPY"},
  "cwd": "/tmp"
}'

run_case "grok-status-after-heredoc-marker" 0 '{
  "toolName": "run_terminal_command",
  "toolInput": {"command": "cat <<EOF\nstatus=1\nEOF"},
  "cwd": "/tmp"
}'

run_case "grok-path-assign-before-heredoc" 2 '{
  "toolName": "run_terminal_command",
  "toolInput": {"command": "path=/x; cat <<EOF\nhello\nEOF"},
  "cwd": "/tmp"
}'

# --- (d) protected environment branches -------------------------------------
# Fixture repos: one per branch state. No commits needed except the detached
# case (symbolic-ref works on an unborn branch).
FIX=$(mktemp -d "${TMPDIR:-/tmp}/bash-policy-test.XXXXXX")
trap 'rm -rf "$FIX"' EXIT
git init -q -b qa "$FIX/qa-repo"
git init -q -b feature-x "$FIX/feature-repo"
git init -q -b master "$FIX/master-repo"
git -C "$FIX/master-repo" -c user.email=t@t -c user.name=t commit -q --allow-empty -m x
git -C "$FIX/master-repo" branch production
git init -q -b production "$FIX/prod-repo"
git init -q -b main "$FIX/detached-repo"
git -C "$FIX/detached-repo" -c user.email=t@t -c user.name=t commit -q --allow-empty -m x
git -C "$FIX/detached-repo" checkout -q --detach

run_case "claude-commit-on-qa" 2 '{
  "tool_name": "Bash",
  "tool_input": {"command": "git commit -m \"fix\""},
  "cwd": "'"$FIX/qa-repo"'"
}'

run_case "claude-commit-on-feature" 0 '{
  "tool_name": "Bash",
  "tool_input": {"command": "git commit -m \"fix\""},
  "cwd": "'"$FIX/feature-repo"'"
}'

run_case "claude-push-refspec-production" 2 '{
  "tool_name": "Bash",
  "tool_input": {"command": "git push origin HEAD:production"},
  "cwd": "'"$FIX/feature-repo"'"
}'

run_case "claude-bare-push-on-master" 0 '{
  "tool_name": "Bash",
  "tool_input": {"command": "git push"},
  "cwd": "'"$FIX/master-repo"'"
}'

run_case "claude-bare-push-on-production" 2 '{
  "tool_name": "Bash",
  "tool_input": {"command": "git push"},
  "cwd": "'"$FIX/prod-repo"'"
}'

run_case "claude-push-feature-refspec-while-on-qa" 0 '{
  "tool_name": "Bash",
  "tool_input": {"command": "git push origin feature-x"},
  "cwd": "'"$FIX/qa-repo"'"
}'

run_case "claude-commit-detached-head" 0 '{
  "tool_name": "Bash",
  "tool_input": {"command": "git commit --allow-empty -m x"},
  "cwd": "'"$FIX/detached-repo"'"
}'

# Compound commands that switch branch before committing
run_case "claude-switch-c-then-commit-from-qa" 0 '{
  "tool_name": "Bash",
  "tool_input": {"command": "git switch -c hotfix-1 && git commit -m fix"},
  "cwd": "'"$FIX/qa-repo"'"
}'

run_case "claude-checkout-b-then-commit-from-qa" 0 '{
  "tool_name": "Bash",
  "tool_input": {"command": "git checkout -b feature-y && git commit -m fix"},
  "cwd": "'"$FIX/qa-repo"'"
}'

run_case "claude-switch-to-production-then-commit-from-master" 2 '{
  "tool_name": "Bash",
  "tool_input": {"command": "git switch production && git commit -m fix"},
  "cwd": "'"$FIX/master-repo"'"
}'

run_case "claude-unparseable-switch-then-commit-from-qa" 2 '{
  "tool_name": "Bash",
  "tool_input": {"command": "git switch - && git commit -m fix"},
  "cwd": "'"$FIX/qa-repo"'"
}'

# Grok-shaped, repo resolved from `git -C` rather than cwd
run_case "grok-commit-via-dash-C-on-production" 2 '{
  "toolName": "run_terminal_command",
  "toolInput": {"command": "git -C '"$FIX/prod-repo"' commit -m x"},
  "cwd": "/tmp"
}'

# --- structural: every deny message names its executable continuation --------
# Each deny must name the exact way out — a backticked command to run, or an
# explicit user action. A future deny shipped without one fails here.
# shellcheck disable=SC2016 # literal backticks are the pattern, not expansion
cont_re='`[^`]+`|have the user|ask the user|user explicitly'
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
done < <(grep -oE 'deny "[^"]+"' "$SCRIPT" | sed -E 's/^deny "//; s/"$//')
if [ "$deny_count" -eq 0 ]; then
  printf 'FAIL deny-continuation: no deny messages extracted from %s\n' "$SCRIPT"
  fail=$((fail + 1))
fi

if [ "$fail" -ne 0 ]; then
  printf '\n%d case(s) failed\n' "$fail"
  exit 1
fi
printf '\nall dual-runtime cases passed\n'
