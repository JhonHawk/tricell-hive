#!/usr/bin/env bash
# instructions-audit.sh — InstructionsLoaded hook. OBSERVATION ONLY.
#
# Measuring instrument for the always-on reduction: records which instruction
# files actually load, when, and under which reason. Claude Code's hook
# reference names InstructionsLoaded for exactly this ("debugging path-specific
# rules or lazy-loaded files"); the event cannot block and its exit code is
# ignored, so this can never affect a session.
#
# WHY IT EXISTS
# The 28,768-token always-on figure comes from `grok inspect`, which reports
# what Grok loads — not what a given Claude Code session loads at a given
# moment. Two documented behaviours make the difference matter:
#   - path-scoped rules load on a file READ, not on every tool use, so a
#     session that writes a new file never loads the rule for it;
#   - they are not re-injected after compaction, so a long session silently
#     drops them.
# Both are invisible without a log. This turns the reduction from an estimate
# into a per-session measurement, before and after any rule moves.
#
# The payload is written RAW, one JSON object per line. Claude Code's docs list
# the common fields but do not fully specify the event-specific ones, so this
# deliberately assumes no schema — the first runs are what reveal it. Parse the
# log afterwards, not here.
#
# Log: $TMPDIR/claude-instructions-audit-<session>.jsonl (ephemeral by design —
# this is an instrument, not a record; nothing here belongs in the repo).
#
# Remove the hook once the always-on reduction is verified: it spawns one
# process per instruction load, which is cheap but not free.

set -uo pipefail

input=$(cat)

session_id=$(printf '%s' "$input" | jq -r '.session_id // .sessionId // "unknown"' 2>/dev/null)
log="${TMPDIR:-/tmp}/claude-instructions-audit-${session_id}.jsonl"

# Stamp arrival time AND the file's size at load time.
#
# The size has to be captured here, not computed later from the path: a rule that
# is moved or deleted between the run and the analysis reads as 0 bytes, which
# silently shrinks the BASELINE a comparison is measured against — the before/after
# delta then comes out short, in the flattering direction. Hit exactly that while
# measuring the always-on reduction.
file_path=$(printf '%s' "$input" | jq -r '.file_path // empty' 2>/dev/null)
size=0
[ -n "$file_path" ] && [ -f "$file_path" ] && size=$(wc -c < "$file_path" 2>/dev/null | tr -d ' ')

printf '%s\n' "$input" \
  | jq -c --arg at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" --argjson sz "${size:-0}" \
      '. + {_at: $at, _bytes: $sz}' >> "$log" 2>/dev/null \
  || printf '{"_at":"%s","_raw_unparseable":true}\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$log" 2>/dev/null

exit 0
