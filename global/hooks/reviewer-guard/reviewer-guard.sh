#!/usr/bin/env bash
# reviewer-guard.sh — PreToolUse (Bash), DENY-FIRST. Agent-scoped.
#
# Deterministic backstop for the read-only doctrine of review/audit agents that
# carry Bash for investigation only (code-reviewer, security-reviewer,
# product-critic, spec-quality-reviewer, code-scout, workspace-custodian).
# Wired through each agent's frontmatter `hooks:` block — NOT through
# settings.json — so it fires only inside those agents. Roster and the
# execute-to-observe exemptions: _support/docs/enforcement-layers.md.
#
# Denies: git mutations (commit/push/merge/rebase/reset/checkout/switch/
# restore/clean/stash/cherry-pick/revert/rm/mv/am/apply/pull/worktree),
# rm/rmdir/unlink/shred/truncate, in-place edits (sed -i, perl -i), package
# installs, and writes landing outside temp space (redirections, tee, cp, mv,
# touch, mkdir, ln, chmod, chown). Temp space (TMPDIR, /tmp, /var/folders,
# /dev/null, scratchpad) stays writable — investigation output needs a home.
#
# Guardrail, not a sandbox: an interpreter one-liner (python -c "open(...)")
# can still write. The residual gap is prompt-convention, documented in
# enforcement-layers.md. Payload parsing mirrors bash-policy.sh (dual-runtime
# shapes), though agent-frontmatter hooks run on Claude Code only.

set -uo pipefail

deny() {
  local reason="$1"
  printf '%s\n' "$reason" >&2
  jq -n --arg r "$reason" '{decision: "deny", reason: $r}' 2>/dev/null || true
  exit 2
}

refuse() {
  deny "reviewer-guard: this agent is read-only — $1 denied. Bash here is for investigation (\`rg\`/\`grep\`/\`find\`, \`git log/diff/show/status/fetch\`, \`gh ... view\`). Report the finding instead; the fix belongs to the dispatching thread."
}

input=$(cat)
tool_name=$(printf '%s' "$input" | jq -r '.tool_name // .toolName // empty' 2>/dev/null)
[ -n "$tool_name" ] || exit 0

command=""
case "$tool_name" in
  Bash|run_terminal_command|Shell)
    command=$(printf '%s' "$input" | jq -r '.tool_input.command // .toolInput.command // empty' 2>/dev/null)
    ;;
esac
[ -n "$command" ] || exit 0

# A path is temp-ok when writing there cannot touch the repo.
temp_ok() {
  case "$1" in
    /tmp/*|/private/tmp/*|/var/folders/*|/private/var/folders/*) return 0 ;;
    /dev/null|/dev/stdout|/dev/stderr|"&"*) return 0 ;;
    "\$TMPDIR"*|"\${TMPDIR"*) return 0 ;;
    *scratchpad*) return 0 ;;
  esac
  if [ -n "${TMPDIR:-}" ]; then
    case "$1" in "${TMPDIR%/}"/*) return 0 ;; esac
  fi
  return 1
}

# --- git mutations (verb parsed after `git`, skipping -C/-c and flags) -------
git_mutating_verb() {
  case "$1" in
    commit|push|merge|rebase|reset|checkout|switch|restore|clean|stash|\
    cherry-pick|revert|rm|mv|am|apply|pull|worktree) return 0 ;;
  esac
  return 1
}

while IFS= read -r seg; do
  [ -n "${seg// /}" ] || continue
  read -ra toks <<< "$seg"
  n=${#toks[@]}

  # git <verb>
  i=0
  while [ "$i" -lt "$n" ] && [ "${toks[i]}" != "git" ]; do i=$((i + 1)); done
  if [ "$i" -lt "$n" ]; then
    j=$((i + 1))
    while [ "$j" -lt "$n" ]; do
      case "${toks[j]}" in
        -C|-c) j=$((j + 2)) ;;
        -*)    j=$((j + 1)) ;;
        *)     git_mutating_verb "${toks[j]}" && refuse "'git ${toks[j]}'"
               break ;;
      esac
    done
  fi

  # deleters — never legitimate for a reviewer, temp included
  del_re='(^|[[:space:]])(sudo[[:space:]]+)?(rm|rmdir|unlink|shred|truncate)([[:space:]]|$)'
  [[ "$seg" =~ $del_re ]] && refuse "'${BASH_REMATCH[3]}'"

  # in-place editors
  sedi_re='(^|[[:space:]])sed[[:space:]]+([^;|&]*[[:space:]])?-[a-zA-Z]*i'
  perli_re='(^|[[:space:]])perl[[:space:]]+([^;|&]*[[:space:]])?-[a-zA-Z]*i'
  [[ "$seg" =~ $sedi_re ]] && refuse "'sed -i'"
  [[ "$seg" =~ $perli_re ]] && refuse "'perl -i'"

  # package installs
  pm_re='(^|[[:space:]])(npm|pnpm|yarn|bun)[[:space:]]+(install|i|add|update|up|upgrade|remove)([[:space:]]|$)'
  pm2_re='(^|[[:space:]])(uv|poetry|pip3?|pipx|cargo|gem|brew|apt|apt-get)[[:space:]]+(pip[[:space:]]+)?(install|add|sync|update|remove)([[:space:]]|$)'
  [[ "$seg" =~ $pm_re ]] && refuse "package install ('${BASH_REMATCH[2]} ${BASH_REMATCH[3]}')"
  [[ "$seg" =~ $pm2_re ]] && refuse "package install ('${BASH_REMATCH[2]} ${BASH_REMATCH[4]}')"

  # find with a mutating action
  findmut_re='(^|[[:space:]])find[[:space:]][^;|&]*(-delete|-exec[[:space:]]+(rm|mv|sed|chmod|chown))'
  [[ "$seg" =~ $findmut_re ]] && refuse "'find' with a mutating action"

  # writers: every path argument must be temp-ok — except cp, where only the
  # DESTINATION (last non-flag argument) writes; its sources are reads.
  k=0
  while [ "$k" -lt "$n" ]; do
    case "${toks[k]}" in
      tee|mv|touch|mkdir|ln|chmod|chown)
        writer="${toks[k]}"
        m=$((k + 1))
        while [ "$m" -lt "$n" ]; do
          arg="${toks[m]}"
          case "$arg" in
            -*) m=$((m + 1)); continue ;;
          esac
          temp_ok "$arg" || refuse "'$writer' writing to '$arg'"
          m=$((m + 1))
        done
        ;;
      cp)
        cp_dst=""
        m=$((k + 1))
        while [ "$m" -lt "$n" ]; do
          case "${toks[m]}" in
            -*) ;;
            *) cp_dst="${toks[m]}" ;;
          esac
          m=$((m + 1))
        done
        if [ -n "$cp_dst" ] && ! temp_ok "$cp_dst"; then
          refuse "'cp' writing to '$cp_dst'"
        fi
        ;;
    esac
    k=$((k + 1))
  done
done < <(printf '%s\n' "$command" | tr ';&|' '\n')

# --- output redirections: every > / >> target must be temp-ok ----------------
while IFS= read -r rtarget; do
  [ -n "$rtarget" ] || continue
  temp_ok "$rtarget" || refuse "output redirection to '$rtarget'"
done < <(printf '%s' "$command" \
  | grep -oE '(^|[^<>])>{1,2}[[:space:]]*[^[:space:];&|)]+' 2>/dev/null \
  | sed -E 's/^.*>{1,2}[[:space:]]*//')

exit 0
