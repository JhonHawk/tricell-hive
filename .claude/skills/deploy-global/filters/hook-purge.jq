# deploy-global step 5 — purge only the exact Hive command for a
# manifest-confirmed orphan. Invoke with --arg bn "$orphan_basename" and
# --arg kind claude|codex. Filter inner hook entries before pruning empty outer
# entries so a user hook co-located in the same event block survives.
def hive_command:
  if $kind == "claude" then
    . == ("$HOME/.claude/hooks/" + $bn)
  elif $kind == "codex" then
    . == ("\"$HOME/.codex/hooks/" + $bn + "\"")
    or . == ("\"$HOME/.codex/hooks/" + $bn + "\" --from-codex-prompt")
  else false end;

if .hooks then
  .hooks |= with_entries(
    .value |= map(
      .hooks |= map(select(((.command // "") | hive_command) | not))
      | select((.hooks // []) | length > 0)
    )
    | select(.value | length > 0)
  )
else . end
