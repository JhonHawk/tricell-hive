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

proc_findings=""
finding_keys=""   # stable identities (pids/ports, not ages) for the cooldown fingerprint

# --- 1. Leaked browser-automation processes ---------------------------------
browser_count=0
browser_oldest=0
# shellcheck disable=SC2009  # pgrep returns only pids; this needs pid+etime+command in one pass, and the
# worst case of an argv false positive is one spurious advisory line in a hygiene report.
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
  proc_findings+="- ${browser_count} agent-browser/Chrome-for-Testing process(es) older than $((AGE_MIN / 60))h (oldest: $(fmt_age "$browser_oldest")) — likely leaked by a previous agent session. Cleanup: review \`agent-browser session list\` with the user and close only sessions confirmed stale. NEVER \`close --all\` as a reflex — it also kills sessions belonging to other projects' concurrently active agents."$'\n'
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
      proc_findings+="- Dev server on port ${port} (PID ${pid}, up $(fmt_age "$mins"), cwd: ${cwd}) — THIS workspace's orphan; verify it is still wanted and offer to stop it."$'\n' ;;
    other-workspace)
      proc_findings+="- Dev server on port ${port} (PID ${pid}, up $(fmt_age "$mins"), cwd: ${cwd}) — from ANOTHER workspace: could be a leak or a concurrent session's active server; offer cleanup, but ask the user to confirm that project is idle before stopping it."$'\n' ;;
    *)
      proc_findings+="- Dev server on port ${port} (PID ${pid}, up $(fmt_age "$mins"), cwd: ${cwd:-unknown}) — workspace unknown; verify with the user before any cleanup."$'\n' ;;
  esac
done < <(lsof -nP -iTCP -sTCP:LISTEN 2>/dev/null | awk 'NR > 1 {print $1, $2, $9}')

# --- 3. Agent-shell drift (Claude Code only) ---------------------------------
# CLAUDE_CODE_SHELL keeps the Bash tool on bash 5 (CLAUDE.md > Shell). It can
# degrade silently (update dropping the variable, brew moving the binary, a
# session that failed to inherit the settings env), so verify it every fresh
# session. Report-only, and exempt from the cooldown: a drifted shell affects
# every command.
shell_drift=""
if [ "${CLAUDECODE:-}" = "1" ]; then
  cfg_shell=$(jq -r '.env.CLAUDE_CODE_SHELL // empty' "$HOME/.claude/settings.json" 2>/dev/null)
  if [ -z "$cfg_shell" ]; then
    shell_drift="CLAUDE_CODE_SHELL is no longer set in settings.json env — the Bash tool is running zsh. The zsh rules in CLAUDE.md > Shell apply until it is restored."
  elif [ ! -x "$cfg_shell" ]; then
    shell_drift="CLAUDE_CODE_SHELL points to '$cfg_shell', which is missing or not executable (brew upgrade moved it?) — Claude fell back to zsh. Fix the path or reinstall bash; zsh rules apply meanwhile."
  else
    # shellcheck disable=SC2016  # single quotes intentional: BASH_VERSINFO must expand in the probed child bash, not in this hook
    bmajor=$("$cfg_shell" -c 'echo "${BASH_VERSINFO[0]:-0}"' 2>/dev/null)
    case "$bmajor" in ''|*[!0-9]*) bmajor=0 ;; esac
    if [ "$bmajor" -lt 5 ]; then
      shell_drift="CLAUDE_CODE_SHELL resolves to bash ${bmajor} (<5) — pre-5 bash breaks the snapshot's BASHPID branch. Point it to Homebrew bash 5."
    elif [ -z "${CLAUDE_CODE_SHELL:-}" ]; then
      # Hooks inherit settings env (live-reloaded per docs), so with the setting
      # present on disk its absence HERE is anomalous — an old version or a
      # failed reload. Soft warning: the message instructs in-band verification.
      # (Never infer drift from ~/.claude/shell-snapshots: snapshots are created
      # lazily on the FIRST Bash call — after SessionStart hooks run — and are
      # deleted on clean exit but survive crashes. Absence is the normal
      # fresh-session state and presence may be stale: both edges are noise.)
      shell_drift="settings declare bash 5 but CLAUDE_CODE_SHELL is absent from this hook's environment — the session may not have loaded it. Verify on the first Bash call: echo \$BASH_VERSION → 5.x means bash is fine and this warning is stale; empty means the Bash tool is on zsh (zsh rules in CLAUDE.md > Shell apply; a full Claude Code restart restores bash)."
    fi
  fi
  if [ -n "$shell_drift" ]; then
    finding_keys+="shell-drift"$'\n'
  fi
fi

# --- 4. Stale hive-profile block (report-only) --------------------------------
# A repo carrying a compiled hive profile (harness/hive-compile.py) stamps the
# hive SHA it was generated from. When the hive has moved past that SHA
# touching global/rules/ or the classifier itself, the block may assert stale
# facts — ONE advisory line, never a mutation. Skipped silently when the hive
# checkout is absent (other machines) or the cwd repo carries no profile.
hive_profile_stale=""
hive_stamp=""
HIVE_REPO="${HIVE_REPO:-$HOME/Development/projects/tricell/tricell-hive}"
if [ -n "$session_cwd" ] && [ -d "$HIVE_REPO/.git" ]; then
  profile_root=$(git -C "$session_cwd" rev-parse --show-toplevel 2>/dev/null || true)
  if [ -n "$profile_root" ] && [ -f "$profile_root/AGENTS.md" ] \
     && grep -q 'hive-profile:start' "$profile_root/AGENTS.md" 2>/dev/null; then
    hive_stamp=$(grep -oE 'hive@[0-9a-f]+' "$profile_root/AGENTS.md" 2>/dev/null | head -1 | cut -d@ -f2)
    hive_head=$(git -C "$HIVE_REPO" rev-parse --short HEAD 2>/dev/null || true)
    if [ -n "$hive_stamp" ] && [ -n "$hive_head" ] && [ "$hive_stamp" != "$hive_head" ]; then
      if ! git -C "$HIVE_REPO" rev-parse --verify --quiet "${hive_stamp}^{commit}" >/dev/null 2>&1; then
        hive_profile_stale="hive-profile stale (stamp hive@${hive_stamp} unknown to the hive checkout); run harness/hive-compile.py"
      elif [ -n "$(git -C "$HIVE_REPO" log --name-only "${hive_stamp}..HEAD" -- global/rules/ harness/hive-compile.py 2>/dev/null)" ]; then
        hive_profile_stale="hive-profile stale (hive moved ${hive_stamp}→${hive_head} touching rules/); run harness/hive-compile.py"
      fi
    fi
  fi
fi
if [ -n "$hive_profile_stale" ]; then
  finding_keys+="hive-profile:${hive_stamp}"$'\n'
fi

# --- Emit ---------------------------------------------------------------------
findings="$proc_findings"
if [ -n "$shell_drift" ]; then
  findings+="- Shell drift: ${shell_drift}"$'\n'
fi
if [ -n "$hive_profile_stale" ]; then
  findings+="- ${hive_profile_stale}"$'\n'
fi
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
    # Shell drift always reports — it degrades every command, cooldown or not.
    [ "$new_key" -eq 1 ] || [ -n "$shell_drift" ] || exit 0
  fi
fi

mkdir -p "$(dirname "$STATE_FILE")" 2>/dev/null
{ date +%s; printf '%s' "$finding_keys"; } > "${STATE_FILE}.tmp" 2>/dev/null \
  && mv "${STATE_FILE}.tmp" "$STATE_FILE" 2>/dev/null

# Header and instructions match the findings actually present — a drift-only
# report must not claim leaked processes were detected.
intro=""
outro=""
if [ -n "$proc_findings" ]; then
  intro="processes likely leaked by previous agent sessions were detected"
  outro="Mention this to the user in your first reply and offer the cleanup. Items from other workspaces need the user to confirm that project is idle first — a concurrent session may be using them. Do NOT kill anything without the user's confirmation."
fi
if [ -n "$shell_drift" ]; then
  if [ -n "$intro" ]; then
    intro+="; the agent-shell (bash 5) check also raised a warning"
  else
    intro="the agent-shell (bash 5) check raised a warning"
  fi
  outro="${outro:+$outro }Mention the shell warning in your first reply and verify it before acting on it."
fi
if [ -n "$hive_profile_stale" ]; then
  if [ -n "$intro" ]; then
    intro+="; this repo's compiled hive profile is stale"
  else
    intro="this repo's compiled hive profile is stale"
  fi
  outro="${outro:+$outro }Offer to regenerate the hive profile; never regenerate it unasked."
fi

report="Session-hygiene report (report-only): ${intro}.
${findings}${outro}"

jq -n --arg ctx "$report" '{
  hookSpecificOutput: {
    hookEventName: "SessionStart",
    additionalContext: $ctx
  }
}'
