#!/usr/bin/env bash
# flow-plan-capture.sh — PostToolUse (ExitPlanMode) hook, NON-BLOCKING.
#
# The organic bridge of the flow pack: when a native plan-mode plan is APPROVED
# inside a flow workspace (a `_support/PROJECT.md` ledger at or above cwd), the
# plan is captured to the session-capture layer —
#   <sessions-home>/YYYY-MM-DD-<slug>/<slug>-plan.md   (Status: planned)
# plus a row in the sessions index — and the model receives the canonical path
# as additionalContext ("the sessions copy is the working plan from here").
# `/flow-build` adopts that file; day-2 continuity is a repo file, not harness
# state.
#
# Opt-out (user decision C2, 2026-07-10): a plan carrying a `Session: no` line
# is NOT captured — the skip is explicit in the approved plan text, never
# silent. No size threshold by design: entering plan mode IS the
# proportionality filter.
#
# Idempotency: same slug with `Status: planned` → overwrite in place (a
# re-approved revision). Status already advanced (building/built/verified) →
# a `-2`/`-3` suffixed sibling, never clobbering executed state.
#
# Never blocks (exit 0 always); silent outside flow workspaces. Every write or
# skip appends one breadcrumb line to $TMPDIR/claude-flow-plan-capture.log so
# the mechanism is auditable.

set -uo pipefail

BREADCRUMB="${TMPDIR:-/tmp}/claude-flow-plan-capture.log"
crumb() { printf '%s %s\n' "$(date '+%F %T')" "$1" >> "$BREADCRUMB" 2>/dev/null || true; }

input=$(cat)
cwd=$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null)
[ -n "$cwd" ] || cwd="$PWD"
plan=$(printf '%s' "$input" | jq -r '.tool_response.plan // empty' 2>/dev/null)
plan_file=$(printf '%s' "$input" | jq -r '.tool_response.filePath // empty' 2>/dev/null)

# Plan text: inline field first, else the plan file ExitPlanMode wrote.
if [ -z "$plan" ] && [ -n "$plan_file" ] && [ -f "$plan_file" ]; then
  plan=$(cat "$plan_file")
fi
if [ -z "$plan" ]; then
  crumb "skip: no plan content (cwd=$cwd)"
  exit 0
fi

# Flow workspace? Walk up for the ledger.
dir="$cwd"; root=""
while [ -n "$dir" ] && [ "$dir" != "/" ]; do
  if [ -f "$dir/_support/PROJECT.md" ]; then root="$dir"; break; fi
  dir=$(dirname "$dir")
done
if [ -z "$root" ]; then
  crumb "skip: no ledger above $cwd (plan present)"
  exit 0
fi

# User opt-out: a `Session: no` line anywhere in the plan header block.
if printf '%s\n' "$plan" | head -20 | grep -qiE '^Session:[[:space:]]*no[[:space:]]*$'; then
  crumb "skip: plan declares Session: no (root=$root)"
  exit 0
fi

# Sessions home per the session-capture detection rule: sibling specs repo if
# present, else the workspace/repo _support staging.
specs_dir=$(find "$root" -maxdepth 1 -type d -name '*-specs' 2>/dev/null | head -1)
if [ -n "$specs_dir" ]; then
  sessions_home="$specs_dir/sessions"
else
  sessions_home="$root/_support/sessions"
fi

# Slug from the plan's first markdown heading; fallback to the plan filename.
title=$(printf '%s\n' "$plan" | grep -m1 -E '^#{1,3} ' | sed -E 's/^#{1,3} //')
[ -n "$title" ] || title=$(basename "${plan_file:-plan}" .md)
slug=$(printf '%s' "$title" | tr '[:upper:]' '[:lower:]' \
  | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//' | cut -c1-40 | sed -E 's/-+$//')
[ -n "$slug" ] || slug="plan"
today=$(date +%F)

# Idempotency: overwrite while still `planned`; suffix once Status advanced.
candidate="$slug"; n=1
while :; do
  target_dir="$sessions_home/$today-$candidate"
  target="$target_dir/$candidate-plan.md"
  if [ ! -f "$target" ]; then break; fi
  if head -3 "$target" | grep -qiE '^Status:[[:space:]]*planned[[:space:]]*$'; then break; fi
  n=$((n + 1)); candidate="$slug-$n"
done

mkdir -p "$target_dir" 2>/dev/null || { crumb "skip: cannot create $target_dir"; exit 0; }
{
  printf '%s\n' "Status: planned"
  printf '%s\n' "Implements: <fill: epic/task refs or —>"
  printf '%s\n' "Captured: $today (native plan mode, flow-plan-capture)"
  printf '\n%s\n' "$plan"
} > "$target" 2>/dev/null || { crumb "skip: cannot write $target"; exit 0; }

# Sessions index row (idempotent: skip if the slug is already listed).
index="$sessions_home/README.md"
if [ ! -f "$index" ]; then
  printf '# Sessions index\n\n' > "$index" 2>/dev/null || true
fi
if [ -f "$index" ] && ! grep -q "$today-$candidate" "$index" 2>/dev/null; then
  # shellcheck disable=SC2016  # backticks are markdown and %s is the printf format; nothing here is meant to expand
  printf -- '- `%s` — in-progress (plan captured from native plan mode)\n' \
    "$today-$candidate" >> "$index" 2>/dev/null || true
fi

crumb "write: $target"

rel_target="${target#"$root"/}"
context="Approved plan captured to $rel_target (Status: planned) — that sessions copy is the working plan from here; keep Status current there. At close: fill its Implements: line and update the ledger's '## Current handoff' (plan pointer), then commit session artifacts (standing-authorized, chore(sessions): $candidate). /flow-build adopts this plan for reconciler execution when asked."
jq -cn --arg ctx "$context" \
  '{hookSpecificOutput:{hookEventName:"PostToolUse",additionalContext:$ctx}}' 2>/dev/null \
  || printf '%s\n' "$context"
exit 0
