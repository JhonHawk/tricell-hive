#!/usr/bin/env bash
# corpus-ratchet.sh — PostToolUse (Write|Edit) hook, NON-BLOCKING.
#
# Hive-only: after writing/editing an always-on rule (a global/rules/ file with no
# paths: frontmatter) or global/CLAUDE.md, recompute the always-on corpus and compare
# it against the declared ceiling. Advisory only — always exits 0, never blocks.
#
# The ceiling is READ from .claude/skills/manage-rules/SKILL.md (criterion 10), never
# duplicated here: manage-rules stays the single source of the number, so moving the
# ceiling there moves it for this hook too.
#
# Lives as a script (not an inline command) so shell locals are not mistaken for
# required env vars by the hook runner.

set -uo pipefail

input=$(cat)
file_path=$(printf '%s' "$input" | jq -r '.tool_input.file_path // .tool_response.filePath // empty' 2>/dev/null)

[ -n "$file_path" ] || exit 0

case "$file_path" in
  */global/rules/*.md|*/global/CLAUDE.md) ;;
  *) exit 0 ;;
esac

repo="${file_path%/global/*}"
[ -d "$repo/global/rules" ] || exit 0

skill="$repo/.claude/skills/manage-rules/SKILL.md"
[ -f "$skill" ] || exit 0

# Ceiling is written as **140,770 B (...)** — strip everything but the digits.
ceiling=$(grep -o '\*\*[0-9,]\{1,\} B' "$skill" | head -1 | tr -cd '0-9')
[ -n "$ceiling" ] || exit 0

total=0
while IFS= read -r f; do
  # A rule with paths: is conditional — it is not part of the always-on corpus.
  if grep -q '^paths:' "$f"; then
    continue
  fi
  size=$(wc -c < "$f" | tr -d '[:space:]')
  total=$((total + size))
done < <(find "$repo/global/rules" -name '*.md')

if [ -f "$repo/global/CLAUDE.md" ]; then
  size=$(wc -c < "$repo/global/CLAUDE.md" | tr -d '[:space:]')
  total=$((total + size))
fi

[ "$total" -gt "$ceiling" ] 2>/dev/null || exit 0

over=$((total - ceiling))
printf '%s\n' "{\"systemMessage\":\"Always-on corpus ${total} B — over the ${ceiling} B ceiling by ${over} B.\",\"hookSpecificOutput\":{\"hookEventName\":\"PostToolUse\",\"additionalContext\":\"The always-on rule corpus now exceeds the ceiling declared in .claude/skills/manage-rules/SKILL.md (criterion 10). Pay for the addition with a cut in the same commit and state both sides in the commit message; only then does the ceiling move to the new measurement.\"}}"

exit 0
