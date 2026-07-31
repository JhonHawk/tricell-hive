#!/usr/bin/env bash
# agent-line-count.sh — PostToolUse (Write|Edit) hook, NON-BLOCKING.
#
# Hive-only: after writing/editing an agent under global/agents/, warn if the
# file exceeds the 120-line agent design limit (AGENTS.md / CLAUDE.md).
# Advisory only — always exits 0; never blocks the tool call.
#
# Lives as a script (not an inline command) so shell locals like $lines are
# not mistaken for required env vars by the hook runner.

set -uo pipefail

input=$(cat)
file_path=$(printf '%s' "$input" | jq -r '.tool_input.file_path // .tool_response.filePath // empty' 2>/dev/null)

[ -n "$file_path" ] || exit 0
[ -f "$file_path" ] || exit 0

# `*` crosses `/` in a case pattern, so this one pattern already covers both
# global/agents/<agent>.md and global/agents/<role>/<agent>.md.
case "$file_path" in
  */global/agents/*.md) ;;
  *) exit 0 ;;
esac

line_count=$(wc -l < "$file_path" | tr -d '[:space:]')
[ -n "$line_count" ] || exit 0

if [ "$line_count" -gt 120 ] 2>/dev/null; then
  name=$(basename "$file_path")
  printf '%s\n' "{\"systemMessage\":\"Agent ${name} has ${line_count} lines (max 120). Consider trimming filler.\",\"hookSpecificOutput\":{\"hookEventName\":\"PostToolUse\",\"additionalContext\":\"The agent file exceeds the 120-line limit defined in CLAUDE.md. Review and trim.\"}}"
fi

exit 0
