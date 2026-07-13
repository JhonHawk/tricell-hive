#!/usr/bin/env bash
# flow-phase-context.sh — UserPromptSubmit hook, NON-BLOCKING, flow workspaces only.
#
# The offer layer of the flow pack ("organic entry" design, 2026-07-13): inside a
# flow workspace, injects the ledger's `Current phase` and `Next suggested` lines —
# once per session, re-firing only when the ledger file changes (a phase transition
# mid-session must not leave the injected state stale). The model can then OFFER the
# right /flow-* command in conversation. Payload is STATE, not routing instructions:
# the retired flow-route-reminder proved per-prompt routing payloads fail (ignored
# instructions, false positives, token pollution); flow skills stay user-gated
# (disable-model-invocation) — the offer-don't-invoke behavior itself is a rule in
# global/CLAUDE.md > Skill Auto-invocation, not something re-argued here per prompt.
#
# Never blocks (exit 0 always; stdout becomes model context).

set -uo pipefail

input=$(cat)
cwd=$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null)
[ -n "$cwd" ] || cwd="$PWD"
session_id=$(printf '%s' "$input" | jq -r '.session_id // empty' 2>/dev/null)

# Flow workspace? Walk up for the ledger.
dir="$cwd"; ledger=""
while [ -n "$dir" ] && [ "$dir" != "/" ]; do
  if [ -f "$dir/_support/PROJECT.md" ]; then ledger="$dir/_support/PROJECT.md"; break; fi
  dir=$(dirname "$dir")
done
[ -n "$ledger" ] || exit 0

# Once per session, keyed to the ledger's mtime so a phase transition re-fires.
mtime=$(stat -f %m "$ledger" 2>/dev/null || stat -c %Y "$ledger" 2>/dev/null || echo 0)
if [ -n "$session_id" ]; then
  marker="${TMPDIR:-/tmp}/claude-flow-phase-context-${session_id}"
  if [ -f "$marker" ] && [ "$(cat "$marker" 2>/dev/null)" = "$mtime" ]; then exit 0; fi
  printf '%s' "$mtime" > "$marker" 2>/dev/null || true
fi

phase=$(grep -m1 -E '^\| *Current phase *\|' "$ledger" 2>/dev/null | sed -E 's/^\| *Current phase *\| *(.*[^ ]) *\|.*$/\1/')
next=$(grep -m1 -E '^- *Next suggested:' "$ledger" 2>/dev/null | sed -E 's/^- *Next suggested: *//')

# Ledger present but both fields missing/blank → stay silent rather than inject noise.
[ -n "$phase$next" ] || exit 0

echo "Flow workspace — state from the ledger ($ledger):"
[ -n "$phase" ] && echo "- Current phase: $phase"
[ -n "$next" ] && echo "- Next suggested step: $next"
echo "- Advisory: the ledger can be stale — verify before relying on it. Flow commands are offered to the user, never invoked uninvited (global rule: Skill Auto-invocation)."
exit 0
