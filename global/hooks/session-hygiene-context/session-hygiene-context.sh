#!/usr/bin/env bash
# session-hygiene-context.sh — SessionStart (startup|clear) hook, NON-BLOCKING.
#
# Deterministic backstop for the session-close hygiene ritual
# (rules/workflow/git-workflow.md > Session close). Most session closes are
# silent (terminal closed, /clear, context exhausted), so there is no
# close-time signal the model can react to; this hook runs the ceremony at
# the next deterministic seam instead — session start — by injecting
# pending-hygiene FACTS: local branches fully merged into the integration
# branch, and branches whose upstream is gone. It injects state, never
# routing instructions; the always-on git-workflow rule owns what to do.
#
# Advisory only — NEVER blocks, exits 0 always. Local git queries only (no
# fetch, no network). Memory-side close (Engram session summary, /memory-sync)
# is owned by the Engram plugin and memory-routing.md — out of scope here.

set -uo pipefail

input=$(cat)
source_evt=$(printf '%s' "$input" | jq -r '.source // empty' 2>/dev/null)
cwd=$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null)

# Only fresh contexts: skip resume (context already has it) and compact (mid-task).
case "$source_evt" in startup|clear) ;; *) exit 0 ;; esac
[ -n "$cwd" ] && [ -d "$cwd" ] || exit 0

git -C "$cwd" rev-parse --is-inside-work-tree >/dev/null 2>&1 || exit 0

# Long-lived branches (environment branches + trunks) are never prune candidates.
protected='^(development|qa|production|main|master)$'

# Integration target: development if it exists, else origin/HEAD, else main/master.
target=""
if git -C "$cwd" show-ref --verify --quiet refs/heads/development; then
  target="development"
else
  target=$(git -C "$cwd" symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's|^origin/||')
  if [ -z "$target" ]; then
    for b in main master; do
      if git -C "$cwd" show-ref --verify --quiet "refs/heads/$b"; then target="$b"; break; fi
    done
  fi
fi

merged=""
if [ -n "$target" ]; then
  merged=$(git -C "$cwd" branch --format='%(refname:short)' --merged "$target" 2>/dev/null \
    | grep -Ev "$protected" | grep -Fxv "$target" | head -8)
fi

gone=$(git -C "$cwd" branch --format='%(refname:short) %(upstream:track)' 2>/dev/null \
  | awk '$2 == "[gone]" {print $1}' | grep -Ev "$protected" | head -8)

# Quiet by default: no pending hygiene -> no injection.
[ -n "$merged" ] || [ -n "$gone" ] || exit 0

facts=""
if [ -n "$merged" ]; then
  facts="local branches fully merged into ${target}: $(printf '%s' "$merged" | tr '\n' ' ' | sed 's/ *$//')"
fi
if [ -n "$gone" ]; then
  [ -n "$facts" ] && facts="${facts}; "
  facts="${facts}branches whose upstream is gone: $(printf '%s' "$gone" | tr '\n' ' ' | sed 's/ *$//')"
fi

ctx="Pending git hygiene from a previous session (deterministic session-close backstop): ${facts}. The ritual that owns this is git-workflow.md > Session close (standing-authorized: prune confirmed-merged branches, report divergences; unmerged branches are decisions, not noise)."
jq -n --arg ctx "$ctx" '{
  hookSpecificOutput: {
    hookEventName: "SessionStart",
    additionalContext: $ctx
  }
}'

exit 0
