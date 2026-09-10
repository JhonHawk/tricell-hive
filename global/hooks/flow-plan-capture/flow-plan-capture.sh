#!/usr/bin/env bash
# flow-plan-capture.sh — legacy advisory capture; explicit PI capture fails closed.
#
#   (default)             Claude Code PostToolUse:ExitPlanMode, or Grok
#                         PostToolUse:exit_plan_mode — plan approved. One entry
#                         point, two payload shapes, told apart by the payload
#                         itself: both harnesses run this same command (Grok
#                         merges ~/.claude/settings.json via compat), so a flag
#                         could not distinguish them.
#   --from-transcript     Claude Code recovery, called by post-tool-hub: the
#                         "clear context" approval denies the tool, so no
#                         PostToolUse fires; the plan is read from the fresh
#                         session's first user turn (`planContent`).
#   --from-codex-prompt   Codex UserPromptSubmit — no ExitPlanMode tool exists
#                         there, but Codex fires this event for its own injected
#                         message, so BOTH approval buttons land here.
#
#   --from-pi-command    PI explicit approval; structured result, exact plan bytes,
#                         Git-root fallback and nonzero capture errors.
#
# Entry points share slug, opt-out and session index conventions. PI uses
# immutable snapshots; legacy captures overwrite revisions while still planned.
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
# a `-2`/`-3` suffixed sibling, never clobbering executed state. PI always suffixes
# differing content; sequential identical approvals reuse the exact snapshot.
#
# Legacy entry points exit 0 and stay silent outside flow workspaces. Every write or
# skip appends one breadcrumb line to $TMPDIR/claude-flow-plan-capture.log so
# the mechanism is auditable.

set -uo pipefail

BREADCRUMB="${TMPDIR:-/tmp}/claude-flow-plan-capture.log"
crumb() { printf '%s %s\n' "$(date '+%F %T')" "$1" >> "$BREADCRUMB" 2>/dev/null || true; }

mode="hook"
harness="claude"
case "${1:-}" in
  --from-pi-command) mode="pi"; harness="pi" ;;
  --from-transcript)  mode="recovery" ;;
  --from-codex-prompt) mode="codex" ;;
esac

capture_error() {
  crumb "error($mode): $1"
  if [ "$mode" = "pi" ]; then
    jq -cn --arg reason "$1" '{status:"error",reason:$reason}'
    exit 1
  fi
  exit 0
}

capture_success() {
  local digest
  digest=$(tail -n +5 "$1" | shasum -a 256) || capture_error "cannot-hash-plan"
  digest=${digest%% *}
  jq -cn --arg path "$1" --arg sha256 "$digest" '{status:"captured",path:$path,sha256:$sha256}'
}

input=$(cat)
if [ "$mode" = "pi" ]; then
  printf '%s' "$input" | jq -e '(.cwd | type) == "string" and (.tool_response.plan | type) == "string" and (.tool_response.plan | length) > 0' >/dev/null 2>&1 \
    || capture_error "invalid-payload"
fi
cwd=$(printf '%s' "$input" | jq -r '.cwd // .workspaceRoot // empty' 2>/dev/null)
[ -n "$cwd" ] || cwd="$PWD"
cwd=$(cd "$cwd" 2>/dev/null && pwd -P) || capture_error "invalid-cwd"

if [ "$mode" = "pi" ]; then
  plan_b64=$(printf '%s' "$input" | jq -r '.tool_response.plan | @base64')
  plan=$(printf '%s' "$plan_b64" | base64 -d) || capture_error "invalid-plan"
elif [ "$mode" = "codex" ]; then
  # Codex UserPromptSubmit — see "Codex: one hook, both approval paths" in the
  # README. Codex has no ExitPlanMode tool, but it DOES fire UserPromptSubmit for
  # the message it injects itself (Claude Code does not), so both plan-approval
  # buttons land here. Which one fired decides where the plan text lives.
  prompt=$(printf '%s' "$input" | jq -r '.prompt // empty' 2>/dev/null)
  [ -n "$prompt" ] || exit 0
  case "$prompt" in
    "A previous agent produced the plan below"*)
      # "Yes, clear context and implement": the prompt IS the plan, behind one
      # instruction paragraph. Drop through the first blank line.
      plan=$(printf '%s\n' "$prompt" | awk 'p{print} /^[[:space:]]*$/{p=1}')
      [ -n "$plan" ] || { crumb "skip(codex): prefix present but no plan body"; exit 0; }
      ;;
    "Implement the plan.")
      # "Yes, implement this plan": context is kept, so the prompt carries no
      # plan — the rollout does, as a structured `item.type == "Plan"`. Matched
      # EXACTLY, not as a prefix: Codex injects this literal with nothing
      # appended, so an exact match keeps a user who types the same words plus
      # their own request from tripping it.
      transcript=$(printf '%s' "$input" | jq -r '.transcript_path // empty' 2>/dev/null)
      [ -n "$transcript" ] && [ -f "$transcript" ] || { crumb "skip(codex): no transcript for keep-context approval"; exit 0; }
      plan_b64=$(jq -r 'select(.payload.item.type == "Plan") | (.payload.item.text | @base64)' "$transcript" 2>/dev/null | tail -1)
      [ -n "$plan_b64" ] || { crumb "skip(codex): no Plan item in $transcript"; exit 0; }
      plan=$(printf '%s' "$plan_b64" | base64 -d 2>/dev/null)
      [ -n "$plan" ] || { crumb "skip(codex): Plan item undecodable"; exit 0; }
      ;;
    *) exit 0 ;;   # Any other prompt: this is every ordinary turn. Silent.
  esac
elif [ "$mode" = "recovery" ]; then
  # Recovery path — see "Why a recovery mode exists" in the README. Approving a
  # plan WITH context cleared resolves ExitPlanMode as a permission DENY, so
  # PostToolUse never fires and the hook path above never sees the plan. The
  # fresh session's first user turn carries it verbatim in `planContent`, a
  # field unique to that path, so the marker is exact — no heuristics.
  transcript=$(printf '%s' "$input" | jq -r '.transcript_path // .transcriptPath // empty' 2>/dev/null)
  [ -n "$transcript" ] && [ -f "$transcript" ] || exit 0
  plan_b64=$(head -60 "$transcript" 2>/dev/null \
    | jq -r 'select(.type == "user" and ((.planContent // "") != "")) | (.planContent | @base64)' 2>/dev/null \
    | head -1)
  # No planContent = the ordinary session. Silent: a crumb per session is noise.
  [ -n "$plan_b64" ] || exit 0
  plan=$(printf '%s' "$plan_b64" | base64 -d 2>/dev/null)
  if [ -z "$plan" ]; then
    crumb "skip(recovery): planContent present but undecodable ($transcript)"
    exit 0
  fi
else
  plan=$(printf '%s' "$input" | jq -r '.tool_response.plan // empty' 2>/dev/null)
  plan_file=$(printf '%s' "$input" | jq -r '.tool_response.filePath // empty' 2>/dev/null)

  # Grok's payload is camelCase and its tool output is `toolResult`, a plain
  # STRING (not Claude's structured object): a banner naming the saved plan file,
  # then a `## Plan:` line and the plan text. Prefer the FILE — the string field
  # is subject to Grok's free-text clipping — and keep the inline parse as the
  # fallback for when the banner wording changes.
  if [ -z "$plan" ] && [ -z "$plan_file" ]; then
    grok_result=$(printf '%s' "$input" \
      | jq -r 'if (.toolResult | type) == "string" then .toolResult else empty end' 2>/dev/null)
    if [ -n "$grok_result" ]; then
      harness="grok"
      plan_file=$(printf '%s\n' "$grok_result" | sed -n 's/^Your plan has been saved at: //p' | head -1)
      [ -f "${plan_file:-}" ] || plan=$(printf '%s\n' "$grok_result" | awk 'p{print} /^## Plan:/{p=1}')
    fi
  fi

  # Plan text: inline field first, else the plan file the tool wrote.
  if [ -z "$plan" ] && [ -n "$plan_file" ] && [ -f "$plan_file" ]; then
    plan=$(cat "$plan_file")
  fi
  if [ -z "$plan" ]; then
    crumb "skip: no plan content (cwd=$cwd)"
    exit 0
  fi
fi

# User opt-out: a `Session: no` line anywhere in the plan header block.
if printf '%s\n' "$plan" | head -20 | grep -qiE '^Session:[[:space:]]*no[[:space:]]*$'; then
  crumb "skip: plan declares Session: no (cwd=$cwd)"
  [ "$mode" != "pi" ] || printf '%s\n' '{"status":"skipped","reason":"session-no"}'
  exit 0
fi

# Flow workspace? Walk up for the ledger.
dir="$cwd"; root=""
while [ -n "$dir" ] && [ "$dir" != "/" ]; do
  if [ -f "$dir/_support/PROJECT.md" ]; then root="$dir"; break; fi
  dir=$(dirname "$dir")
done
if [ -z "$root" ] && [ "$mode" = "pi" ]; then
  root=$(git -C "$cwd" rev-parse --show-toplevel 2>/dev/null) || root=""
  if [ -z "$root" ]; then
    printf '%s\n' '{"status":"session_only","reason":"no-repository"}'
    exit 0
  fi
fi
if [ -z "$root" ]; then
  crumb "skip: no ledger above $cwd (plan present)"
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

ensure_index() {
  local entry="$1" index="$sessions_home/README.md"
  if [ ! -f "$index" ]; then
    if ! (set -C; printf '# Sessions index\n\n' > "$index") 2>/dev/null && [ ! -f "$index" ]; then
      [ "$mode" != "pi" ] || capture_error "cannot-create-index"
      return 0
    fi
  fi
  if ! grep -Fq -- "$entry" "$index" 2>/dev/null; then
    # shellcheck disable=SC2016 # Markdown backticks are literal.
    if ! printf -- '- `%s` — in-progress (plan captured from native plan mode)\n' "$entry" >> "$index" 2>/dev/null; then
      [ "$mode" != "pi" ] || capture_error "cannot-update-index"
    fi
  fi
}

# Re-entry guard for both non-ExitPlanMode paths. A `/resume` of a
# context-cleared session replays the same `planContent` under a NEW session id,
# so the once-per-session marker in post-tool-hub cannot stop it; on Codex, a
# re-approval of an unchanged plan reaches the hook again. Already captured,
# whatever the date or Status → stop here; without this, a plan already advanced
# to `building` would land a `-2` sibling reset to `planned`.
if [ "$mode" != "hook" ]; then
  for existing in "$sessions_home"/*/"$slug"-plan.md "$sessions_home"/*/"$slug"-[0-9]*-plan.md; do
    [ -f "$existing" ] || continue
    if { [ "$mode" = "pi" ] && [ "$(tail -n +5 "$existing" | base64 | tr -d '\n')" = "$plan_b64" ]; } ||
       { [ "$mode" != "pi" ] && [ "$(tail -n +5 "$existing" 2>/dev/null)" = "$plan" ]; }; then
      crumb "skip(recovery): already captured at $existing"
      if [ "$mode" = "pi" ]; then
        ensure_index "$(basename "$(dirname "$existing")")"
        capture_success "$existing"
      fi
      exit 0
    fi
  done
fi

# Legacy revisions overwrite while planned; PI reserves immutable snapshots.
if [ "$mode" = "pi" ]; then
  mkdir -p "$sessions_home" 2>/dev/null || capture_error "cannot-create-destination"
fi
candidate="$slug"; n=1
while :; do
  target_dir="$sessions_home/$today-$candidate"
  target="$target_dir/$candidate-plan.md"
  if [ "$mode" = "pi" ]; then
    # Reserve a distinct snapshot atomically across concurrent approvals.
    if mkdir "$target_dir" 2>/dev/null; then break; fi
    [ -e "$target_dir" ] || capture_error "cannot-reserve-destination"
  else
    if [ ! -f "$target" ]; then break; fi
    if head -3 "$target" | grep -qiE '^Status:[[:space:]]*planned[[:space:]]*$'; then break; fi
  fi
  n=$((n + 1)); candidate="$slug-$n"
done

# Recovery can fire long after the approval (a resumed session, a purged TMPDIR
# marker), so its `Status: planned` is an assumption, not an observation — the
# work may already be in git. The artifact says so itself: whoever reads it
# without going through /flow-build's reconciler sees the caveat here.
captured_caveat=""
[ "$mode" = "recovery" ] && captured_caveat="; recovered after the fact — Status is unverified against git"

mkdir -p "$target_dir" 2>/dev/null || capture_error "cannot-create-destination"
write_target="$target"
if [ "$mode" = "pi" ]; then
  write_target=$(mktemp "$target_dir/.capture.XXXXXX") || capture_error "cannot-create-temporary-file"
  trap 'rm -f "$write_target"' EXIT
fi
{
  printf '%s\n' "Status: planned"
  printf '%s\n' "Implements: <fill: epic/task refs or —>"
  printf '%s\n' "Captured: $today (native plan mode, flow-plan-capture)${captured_caveat}"
  if [ "$mode" = "pi" ]; then
    printf '\n'
    printf '%s' "$plan_b64" | base64 -d
  else
    printf '\n%s\n' "$plan"
  fi
} > "$write_target" 2>/dev/null || capture_error "cannot-write-plan"
if [ "$mode" = "pi" ]; then
  mv "$write_target" "$target" || capture_error "cannot-publish-plan"
  trap - EXIT
fi

ensure_index "$today-$candidate"

crumb "write${mode:+($mode)}${harness:+[$harness]}: $target"

if [ "$mode" = "pi" ]; then
  capture_success "$target"
  exit 0
fi

rel_target="${target#"$root"/}"
whence="captured"
[ "$mode" = "recovery" ] && whence="captured (recovered: the approval cleared the context, so the capture ran now instead of at approval time)"
[ "$mode" = "codex" ] && whence="captured (on plan approval)"
context="Approved plan $whence to $rel_target (Status: planned) — that sessions copy is the working plan from here; keep Status current there. At close: fill its Implements: line and update the ledger's '## Current handoff' (plan pointer), then commit session artifacts (standing-authorized, chore(sessions): $candidate). /flow-build adopts this plan for reconciler execution when asked."

# Recovery mode is invoked BY post-tool-hub, which owns the envelope: emit the
# bare text and let the caller fold it into its single additionalContext.
if [ "$mode" = "recovery" ]; then
  printf '%s\n' "$context"
  exit 0
fi
event="PostToolUse"
[ "$mode" = "codex" ] && event="UserPromptSubmit"
jq -cn --arg ctx "$context" --arg ev "$event" \
  '{hookSpecificOutput:{hookEventName:$ev,additionalContext:$ctx}}' 2>/dev/null \
  || printf '%s\n' "$context"
exit 0
