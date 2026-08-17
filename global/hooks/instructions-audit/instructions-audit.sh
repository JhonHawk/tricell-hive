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

# Stamp arrival time: the event carries a reason but not a clock, and the
# ordering between session_start and later lazy loads is the whole point.
printf '%s\n' "$input" \
  | jq -c --arg at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" '. + {_at: $at}' >> "$log" 2>/dev/null \
  || printf '{"_at":"%s","_raw_unparseable":true}\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$log" 2>/dev/null

exit 0
