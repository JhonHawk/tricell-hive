#!/usr/bin/env bash
# ensure-engram-555.sh — keep the Engram #555 detect_project hotfix applied
# across Codex / Claude Code / OpenCode after plugin updates wipe the cache.
#
# Upstream: https://github.com/Gentleman-Programming/engram/issues/555
# Delete this kit when upstream ships the fix natively.
#
# Usage:
#   ./ensure-engram-555.sh              # status (exit 0 = all ok, 1 = drift)
#   ./ensure-engram-555.sh apply        # re-apply missing targets only
#   ./ensure-engram-555.sh install      # install kit + LaunchAgent watcher
#   ./ensure-engram-555.sh uninstall    # remove LaunchAgent (+ optional installed kit)
#
# Marker of a healthy fix: the file contains ENGRAM_PROJECT AND a .engram/config.json walk.

set -euo pipefail

# launchd jobs inherit a minimal PATH (/usr/bin:/bin). Prefer Homebrew when present.
export PATH="/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin${PATH:+:$PATH}"

# Resolve through symlinks (e.g. ~/.local/bin/ensure-engram-555 → kit dir).
# dirname "$0" alone would look for patches next to the symlink and fail.
_resolve_kit_dir() {
  local src="${BASH_SOURCE[0]:-$0}"
  while [ -L "$src" ]; do
    local dir
    dir="$(cd -P "$(dirname "$src")" && pwd)"
    src="$(readlink "$src")"
    case "$src" in
      /*) ;;
      *) src="$dir/$src" ;;
    esac
  done
  cd -P "$(dirname "$src")" && pwd
}
KIT_DIR="$(_resolve_kit_dir)"
MARKER='ENGRAM_PROJECT'
CONFIG_MARKER='.engram/config.json'
INSTALLED_KIT="${ENGRAM_HOTFIX_555_DIR:-$HOME/.engram/hotfix-555}"
LAUNCH_LABEL='com.jmartinez.engram-555-ensure'
LAUNCH_PLIST="$HOME/Library/LaunchAgents/${LAUNCH_LABEL}.plist"
LOG_FILE="${ENGRAM_HOTFIX_555_LOG:-$HOME/.engram/hotfix-555.log}"

PATCH_CODEX='helpers-issue555.patch'
PATCH_CLAUDE='helpers-issue555-claude-code.patch'
PATCH_OPENCODE='engram-ts-issue555-opencode.patch'

log() {
  local msg="[$(date '+%Y-%m-%d %H:%M:%S')] $*"
  printf '%s\n' "$msg"
  mkdir -p "$(dirname "$LOG_FILE")" 2>/dev/null || true
  printf '%s\n' "$msg" >>"$LOG_FILE" 2>/dev/null || true
}

has_fix() {
  local f="$1"
  [[ -f "$f" ]] || return 1
  grep -q "$MARKER" "$f" && grep -qF "$CONFIG_MARKER" "$f"
}

# Collect "patch_file|plugin_root" pairs for every live install we can find.
collect_targets() {
  local d f

  # Codex cache (every version dir) + marketplace source
  for d in "$HOME"/.codex/plugins/cache/engram/engram/*/; do
    [[ -d "$d" ]] || continue
    f="${d}scripts/_helpers.sh"
    [[ -f "$f" ]] && printf '%s|%s\n' "$PATCH_CODEX" "${d%/}"
  done
  f="$HOME/.codex/.tmp/marketplaces/engram/plugin/codex/scripts/_helpers.sh"
  if [[ -f "$f" ]]; then
    printf '%s|%s\n' "$PATCH_CODEX" "$HOME/.codex/.tmp/marketplaces/engram/plugin/codex"
  fi

  # Claude Code cache + marketplace source
  for d in "$HOME"/.claude/plugins/cache/engram/engram/*/; do
    [[ -d "$d" ]] || continue
    f="${d}scripts/_helpers.sh"
    [[ -f "$f" ]] && printf '%s|%s\n' "$PATCH_CLAUDE" "${d%/}"
  done
  f="$HOME/.claude/plugins/marketplaces/engram/plugin/claude-code/scripts/_helpers.sh"
  if [[ -f "$f" ]]; then
    printf '%s|%s\n' "$PATCH_CLAUDE" "$HOME/.claude/plugins/marketplaces/engram/plugin/claude-code"
  fi

  # OpenCode (single file under plugins/)
  f="$HOME/.config/opencode/plugins/engram.ts"
  if [[ -f "$f" ]]; then
    printf '%s|%s\n' "$PATCH_OPENCODE" "$HOME/.config/opencode"
  fi
}

target_file_for() {
  local patch="$1" root="$2"
  case "$patch" in
    "$PATCH_OPENCODE") printf '%s\n' "$root/plugins/engram.ts" ;;
    *) printf '%s\n' "$root/scripts/_helpers.sh" ;;
  esac
}

status_all() {
  local line patch root file status missing=0 total=0
  printf '%-10s  %-8s  %s\n' 'HARNESS' 'STATUS' 'FILE'
  printf '%-10s  %-8s  %s\n' '----------' '--------' '----'
  while IFS='|' read -r patch root; do
    [[ -n "$patch" ]] || continue
    total=$((total + 1))
    file="$(target_file_for "$patch" "$root")"
    case "$patch" in
      "$PATCH_CODEX") harness=Codex ;;
      "$PATCH_CLAUDE") harness=Claude ;;
      "$PATCH_OPENCODE") harness=OpenCode ;;
      *) harness=other ;;
    esac
    if has_fix "$file"; then
      status=OK
    else
      status=MISSING
      missing=$((missing + 1))
    fi
    printf '%-10s  %-8s  %s\n' "$harness" "$status" "$file"
  done < <(collect_targets)

  if [[ "$total" -eq 0 ]]; then
    log "status: no Engram harness plugin files found"
    return 1
  fi
  if [[ "$missing" -gt 0 ]]; then
    log "status: $missing/$total target(s) missing #555 fix"
    return 1
  fi
  log "status: all $total target(s) have #555 fix"
  return 0
}

apply_one() {
  local patch="$1" root="$2"
  local patch_path="$KIT_DIR/$patch"
  local file
  file="$(target_file_for "$patch" "$root")"

  if [[ ! -f "$patch_path" ]]; then
    log "ERROR: patch missing: $patch_path"
    return 1
  fi
  if [[ ! -f "$file" ]]; then
    log "skip (no file): $file"
    return 0
  fi
  if has_fix "$file"; then
    log "already applied: $file"
    return 0
  fi

  # Prefer patch; if upstream rewrote the file, fall back to surgical insert
  # only when the classic detect_project header is still recognizable.
  if patch -d "$root" -p1 --forward --dry-run <"$patch_path" >/dev/null 2>&1; then
    patch -d "$root" -p1 --forward --silent <"$patch_path"
    if has_fix "$file"; then
      log "patched: $file"
      return 0
    fi
  fi

  log "ERROR: could not apply $patch to $root (upstream may have rewritten detect_project — re-derive the patch)"
  return 1
}

apply_all() {
  local line patch root rc=0 applied=0
  while IFS='|' read -r patch root; do
    [[ -n "$patch" ]] || continue
    if apply_one "$patch" "$root"; then
      applied=$((applied + 1))
    else
      rc=1
    fi
  done < <(collect_targets)

  # Functional smoke: if tricell-hive is present, Claude or Codex helpers must resolve project_name.
  local hive="$HOME/Development/projects/tricell/tricell-hive"
  local helper
  for helper in \
    "$HOME"/.claude/plugins/cache/engram/engram/*/scripts/_helpers.sh \
    "$HOME"/.codex/plugins/cache/engram/engram/*/scripts/_helpers.sh
  do
    [[ -f "$helper" ]] || continue
    if [[ -f "$hive/.engram/config.json" ]] && has_fix "$helper"; then
      # shellcheck disable=SC1090
      local got
      got="$(bash -c "source \"$helper\"; detect_project \"$hive\"")"
      if [[ "$got" != "tricell-hive" ]]; then
        log "WARN: detect_project($hive) => '$got' (expected tricell-hive) via $helper"
        rc=1
      else
        log "smoke ok: detect_project => tricell-hive ($helper)"
      fi
      break
    fi
  done

  return "$rc"
}

install_kit_and_watcher() {
  mkdir -p "$INSTALLED_KIT" "$(dirname "$LOG_FILE")" "$HOME/Library/LaunchAgents" "$HOME/.local/bin"

  # Copy the whole kit (script + patches + README) to a stable path outside any
  # plugin cache so LaunchAgent survives repo checkouts and plugin updates.
  rsync -a --delete \
    --exclude '.DS_Store' \
    "$KIT_DIR/" "$INSTALLED_KIT/"
  chmod +x "$INSTALLED_KIT/ensure-engram-555.sh"

  ln -sfn "$INSTALLED_KIT/ensure-engram-555.sh" "$HOME/.local/bin/ensure-engram-555"

  cat >"$LAUNCH_PLIST" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>${LAUNCH_LABEL}</string>
  <key>ProgramArguments</key>
  <array>
    <string>/bin/bash</string>
    <string>${INSTALLED_KIT}/ensure-engram-555.sh</string>
    <string>apply</string>
  </array>
  <key>EnvironmentVariables</key>
  <dict>
    <key>PATH</key>
    <string>/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    <key>HOME</key>
    <string>${HOME}</string>
  </dict>
  <key>WatchPaths</key>
  <array>
    <string>${HOME}/.codex/plugins/cache/engram</string>
    <string>${HOME}/.claude/plugins/cache/engram</string>
    <string>${HOME}/.codex/.tmp/marketplaces/engram</string>
    <string>${HOME}/.claude/plugins/marketplaces/engram</string>
    <string>${HOME}/.config/opencode/plugins</string>
  </array>
  <key>ThrottleInterval</key>
  <integer>30</integer>
  <key>RunAtLoad</key>
  <true/>
  <key>StandardOutPath</key>
  <string>${LOG_FILE}</string>
  <key>StandardErrorPath</key>
  <string>${LOG_FILE}</string>
</dict>
</plist>
EOF

  launchctl bootout "gui/$(id -u)/${LAUNCH_LABEL}" 2>/dev/null || true
  launchctl bootstrap "gui/$(id -u)" "$LAUNCH_PLIST"
  launchctl enable "gui/$(id -u)/${LAUNCH_LABEL}" 2>/dev/null || true
  # Kick once now so the installed copy is healthy.
  "$INSTALLED_KIT/ensure-engram-555.sh" apply
  log "installed kit → $INSTALLED_KIT"
  log "symlink → $HOME/.local/bin/ensure-engram-555"
  log "LaunchAgent → $LAUNCH_PLIST (WatchPaths on engram plugin caches)"
}

uninstall_watcher() {
  launchctl bootout "gui/$(id -u)/${LAUNCH_LABEL}" 2>/dev/null || true
  rm -f "$LAUNCH_PLIST"
  rm -f "$HOME/.local/bin/ensure-engram-555"
  log "uninstalled LaunchAgent and ~/.local/bin/ensure-engram-555"
  log "left kit at $INSTALLED_KIT (remove manually if desired)"
}

cmd="${1:-status}"
case "$cmd" in
  status|"")
    status_all
    ;;
  apply)
    apply_all
    status_all
    ;;
  install)
    # Apply from the calling kit first, then install durable copy + watcher.
    apply_all || true
    install_kit_and_watcher
    status_all
    ;;
  uninstall)
    uninstall_watcher
    ;;
  *)
    echo "usage: $0 [status|apply|install|uninstall]" >&2
    exit 2
    ;;
esac
