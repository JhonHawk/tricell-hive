#!/usr/bin/env bash
# code-search-routing.sh — deterministic code-search tool routing per repo.
#
# Two modes, selected by $1:
#   pre  — PreToolUse (Bash + codegraph/jbcontext MCP tools): DENIES a tool in
#          repos where it is contraindicated (map: code-search-routing.json).
#          Deny = exit 2 (blocks the call; stderr redirects the agent).
#   post — PostToolUse (same matchers): injects the anti-conclusion reminder
#          right after an index tool ran. Advisory, never blocks (exit 0).
#
# Evidence base: benchmarks v3/v4 (tricell-hive/_support/archive/audits/
# 2026-07-21-benchmark-*.html). Policy: global/rules/tools/code-search.md.
# The map is versioned config — extend it with evidence, not intuition.

set -uo pipefail

MODE="${1:-post}"
MAP="$HOME/.claude/hooks/code-search-routing.json"

input=$(cat)
tool_name=$(printf '%s' "$input" | jq -r '.tool_name // empty' 2>/dev/null)
[ -n "$tool_name" ] || exit 0

# Which index tool (if any) does this call involve?
tool=""
haystack=""
case "$tool_name" in
  Bash)
    cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)
    haystack="$cmd"
    if printf '%s' "$cmd" | grep -qE '(^|[^[:alnum:]_-])codegraph([^[:alnum:]_-]|$)'; then tool="codegraph"; fi
    if printf '%s' "$cmd" | grep -qE '(^|[^[:alnum:]_-])jbcontext([^[:alnum:]_-]|$)'; then tool="${tool:+$tool }jbcontext"; fi
    ;;
  mcp__codegraph__*)
    tool="codegraph"
    haystack=$(printf '%s' "$input" | jq -r '.tool_input.projectPath // .cwd // empty' 2>/dev/null)
    ;;
  mcp__jbcontext__*)
    tool="jbcontext"
    haystack=$(printf '%s' "$input" | jq -r '.tool_input.pathFilter // .cwd // empty' 2>/dev/null)
    ;;
  *) exit 0 ;;
esac
[ -n "$tool" ] || exit 0

cwd=$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null)
haystack="$haystack $cwd"

if [ "$MODE" = "pre" ] && [ -f "$MAP" ]; then
  # Map rows: {"repos": ["substr", ...], "deny": ["codegraph"], "reason": "..."}
  n=$(jq 'length' "$MAP" 2>/dev/null || echo 0)
  i=0
  while [ "$i" -lt "$n" ]; do
    row=$(jq -c ".[$i]" "$MAP")
    i=$((i + 1))
    for t in $tool; do
      jq -e --arg t "$t" '.deny // [] | index($t)' >/dev/null 2>&1 <<<"$row" || continue
      while IFS= read -r repo; do
        [ -n "$repo" ] || continue
        if printf '%s' "$haystack" | grep -qF "$repo"; then
          reason=$(jq -r '.reason // "contraindicated here"' <<<"$row")
          echo "code-search-routing: '$t' is DENIED in repo '$repo' — $reason. Use the routed alternative (rules/tools/code-search.md)." >&2
          exit 2
        fi
      done < <(jq -r '.repos[]?' <<<"$row")
    done
  done
  exit 0
fi

if [ "$MODE" = "post" ]; then
  cat <<'EOF'
code-search discipline (rules/tools/code-search.md): an index result is a pointer, never a verdict. Read the cited file before citing it; never conclude absence from an index — "does not exist" requires an exhaustive rg sweep (0 hits, broad vocabulary); check a hit is alive (callers/imports) before building on it; on index-vs-disk conflict, disk wins.
EOF
fi
exit 0
