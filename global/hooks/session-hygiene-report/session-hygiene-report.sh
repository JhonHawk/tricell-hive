#!/usr/bin/env bash
# session-hygiene-report.sh — SessionStart hook, REPORT-ONLY.
#
# Surfaces processes likely leaked by previous agent sessions so the new
# session can offer cleanup on its first reply:
#   1. agent-browser / Chrome for Testing processes older than AGE_MIN.
#   2. Local dev servers (node/bun/deno) listening on typical dev ports,
#      older than AGE_MIN.
#
# It NEVER kills anything — a listed process may belong to a concurrently
# active session, so cleanup is always proposed to the user, not executed.
# Silent when nothing qualifies, and on resume/compact (fresh sessions only).
#
# Non-goals: Docker containers (docker ps can hang when the daemon is down;
# project DBs are long-lived by design) and non-dev listeners (system
# services, MCP daemons, editors).

set -uo pipefail

AGE_MIN=120   # minutes; younger processes are assumed to be in active use

# Cooldown: report at most once per TTL machine-wide (state file shared by the
# Claude Code and Codex deployments), UNLESS a finding not present in the
# last-reported set appears — new information always breaks the silence.
TTL_MIN=360   # 6h
STATE_FILE="${SESSION_HYGIENE_STATE:-${XDG_CACHE_HOME:-$HOME/.cache}/session-hygiene-report.state}"

# Fresh sessions only (startup/clear). Missing/unparseable input (e.g. the
# Codex hook runner) defaults to reporting rather than silence.
input=$(cat 2>/dev/null || true)
source_evt=$(printf '%s' "$input" | jq -r '.source // "startup"' 2>/dev/null) || source_evt=startup
case "$source_evt" in
  startup|clear|"") ;;
  *) exit 0 ;;
esac

# Session cwd, when the runner provides it — used to separate this workspace's
# leftovers from other projects' processes (which may be in active use there).
session_cwd=$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null) || session_cwd=""

# etime ([[dd-]hh:]mm:ss) -> whole minutes
etime_minutes() {
  awk -v e="$1" 'BEGIN {
    d = 0; rest = e
    if (index(e, "-") > 0) { split(e, a, "-"); d = a[1]; rest = a[2] }
    n = split(rest, t, ":")
    if (n == 2)      m = t[1]
    else if (n == 3) m = t[1] * 60 + t[2]
    else             m = 0
    print d * 1440 + m
  }'
}

fmt_age() {  # minutes -> "3h" / "27h" / "2d"
  local m=$1
  if [ "$m" -ge 2880 ]; then echo "$((m / 1440))d"; else echo "$((m / 60))h"; fi
}

findings=""
finding_keys=""   # stable identities (pids/ports, not ages) for the cooldown fingerprint

# --- 1. Leaked browser-automation processes ---------------------------------
browser_count=0
browser_oldest=0
while IFS= read -r line; do
  [ -n "$line" ] || continue
  pid=$(printf '%s' "$line" | awk '{print $1}')
  etime=$(printf '%s' "$line" | awk '{print $2}')
  mins=$(etime_minutes "$etime")
  [ "$mins" -ge "$AGE_MIN" ] || continue
  browser_count=$((browser_count + 1))
  [ "$mins" -gt "$browser_oldest" ] && browser_oldest=$mins
  finding_keys+="browser:${pid}"$'\n'
done < <(ps -axo pid=,etime=,command= | grep -E 'agent-browser|Chrome for Testing' | grep -v grep)

if [ "$browser_count" -gt 0 ]; then
  findings+="- ${browser_count} agent-browser/Chrome-for-Testing process(es) older than $((AGE_MIN / 60))h (oldest: $(fmt_age "$browser_oldest")) — likely leaked by a previous agent session. Cleanup: review \`agent-browser session list\` with the user and close only sessions confirmed stale. NEVER \`close --all\` as a reflex — it also kills sessions belonging to other projects' concurrently active agents."$'\n'
fi

# --- 2. Long-lived dev-server listeners --------------------------------------
# Typical dev ports only: 3000-3999 (Next/Express), 4200-4299 (Angular),
# 4321 (Astro), 5173-5179 (Vite), 8080-8089. Runtime commands only.
seen_pids=" "
while IFS= read -r line; do
  [ -n "$line" ] || continue
  cmd=$(printf '%s' "$line" | awk '{print $1}')
  pid=$(printf '%s' "$line" | awk '{print $2}')
  port=$(printf '%s' "$line" | awk '{print $NF}' | sed 's/.*://; s/[^0-9].*//')
  case "$cmd" in node*|bun*|deno*|next*|pnpm*|npm*|yarn*|vite*) ;; *) continue ;; esac
  [ -n "$port" ] || continue
  if ! { { [ "$port" -ge 3000 ] && [ "$port" -le 3999 ]; } \
      || { [ "$port" -ge 4200 ] && [ "$port" -le 4299 ]; } \
      || [ "$port" -eq 4321 ] \
      || { [ "$port" -ge 5173 ] && [ "$port" -le 5179 ]; } \
      || { [ "$port" -ge 8080 ] && [ "$port" -le 8089 ]; }; }; then
    continue
  fi
  case "$seen_pids" in *" $pid "*) continue ;; esac   # IPv4+IPv6 dedup
  seen_pids+="$pid "
  etime=$(ps -o etime= -p "$pid" 2>/dev/null | tr -d ' ')
  [ -n "$etime" ] || continue
  mins=$(etime_minutes "$etime")
  [ "$mins" -ge "$AGE_MIN" ] || continue
  finding_keys+="dev:${pid}:${port}"$'\n'
  cwd=$(lsof -a -p "$pid" -d cwd -Fn 2>/dev/null | sed -n 's/^n//p' | head -1)
  scope="unclassified"
  if [ -n "$session_cwd" ] && [ -n "$cwd" ]; then
    case "$cwd" in
      "$session_cwd"|"$session_cwd"/*) scope="this-workspace" ;;
      *) scope="other-workspace" ;;
    esac
  fi
  case "$scope" in
    this-workspace)
      findings+="- Dev server on port ${port} (PID ${pid}, up $(fmt_age "$mins"), cwd: ${cwd}) — THIS workspace's leftover; verify it is still wanted and offer to stop it."$'\n' ;;
    other-workspace)
      findings+="- Dev server on port ${port} (PID ${pid}, up $(fmt_age "$mins"), cwd: ${cwd}) — from ANOTHER workspace: could be a leak or a concurrent session's active server; offer cleanup, but ask the user to confirm that project is idle before stopping it."$'\n' ;;
    *)
      findings+="- Dev server on port ${port} (PID ${pid}, up $(fmt_age "$mins"), cwd: ${cwd:-unknown}) — workspace unknown; verify with the user before any cleanup."$'\n' ;;
  esac
done < <(lsof -nP -iTCP -sTCP:LISTEN 2>/dev/null | awk 'NR > 1 {print $1, $2, $9}')

# --- Emit ---------------------------------------------------------------------
[ -n "$findings" ] || exit 0

# Cooldown gate: within TTL, stay silent unless a current key was not in the
# last-reported set (a shrinking set never breaks the silence — nothing new).
if [ -s "$STATE_FILE" ]; then
  last_ts=$(head -1 "$STATE_FILE" 2>/dev/null)
  case "$last_ts" in ''|*[!0-9]*) last_ts=0 ;; esac
  now=$(date +%s)
  if [ $((now - last_ts)) -lt $((TTL_MIN * 60)) ]; then
    new_key=0
    while IFS= read -r key; do
      [ -n "$key" ] || continue
      grep -Fxq "$key" "$STATE_FILE" || { new_key=1; break; }
    done <<< "$finding_keys"
    [ "$new_key" -eq 1 ] || exit 0
  fi
fi

mkdir -p "$(dirname "$STATE_FILE")" 2>/dev/null
{ date +%s; printf '%s' "$finding_keys"; } > "${STATE_FILE}.tmp" 2>/dev/null \
  && mv "${STATE_FILE}.tmp" "$STATE_FILE" 2>/dev/null

report="Session-hygiene report (report-only): processes likely leaked by previous agent sessions were detected.
${findings}Mention this to the user in your first reply and offer the cleanup. Items from other workspaces need the user to confirm that project is idle first — a concurrent session may be using them. Do NOT kill anything without the user's confirmation."

jq -n --arg ctx "$report" '{
  hookSpecificOutput: {
    hookEventName: "SessionStart",
    additionalContext: $ctx
  }
}'
