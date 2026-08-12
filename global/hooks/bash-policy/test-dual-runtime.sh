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

if [ "$fail" -ne 0 ]; then
  printf '\n%d case(s) failed\n' "$fail"
  exit 1
fi
printf '\nall dual-runtime cases passed\n'
