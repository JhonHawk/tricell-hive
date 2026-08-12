#!/usr/bin/env bash
# bash-policy.sh — PreToolUse (shell + codegraph/jbcontext MCP), DENY-FIRST.
#
# Dual-runtime: Claude Code (Bash, tool_name/tool_input) and Grok Build
# (run_terminal_command, toolName/toolInput; MCP as server__tool). Keep the
# field fallbacks in sync with post-tool-hub.sh.
#
# One process for every pre-execution shell policy gate. The order is
# load-bearing: every DENY runs first (exit 2 + human reason on stderr;
# Grok-safe JSON decision on stdout), and only then may the advisory emit JSON.
#
#   Denies:   (a) repo-scoped index routing (map: bash-policy.json)
#             (b) pip ban — uv is the only Python package manager
#             (c) package-manager mixing vs the nearest lockfile
#             (d) lockfile deletion
#   Advisory: (e) pre-push local-quality-gate reminder
#
# Non-shell (MCP index) tools see (a) only; (b)-(e) are shell-command policy.
#
# `permissionDecision: "allow"` is emitted ONLY inside the push advisory — a
# global allow would auto-approve arbitrary shell.
#
# Policy owners: rules/tools/code-search.md, CLAUDE.md > Python Dependency
# Management, CLAUDE.md > Package Manager, CLAUDE.md > Build & Lint.

set -uo pipefail

MAP="$HOME/.claude/hooks/bash-policy.json"

# Deny for both harnesses: Claude honors exit 2 + stderr; Grok honors
# {"decision":"deny"} and/or exit 2. Always print both.
deny() {
  local reason="$1"
  printf '%s\n' "$reason" >&2
  jq -n --arg r "$reason" '{decision: "deny", reason: $r}' 2>/dev/null || true
  exit 2
}

input=$(cat)

# Dual-runtime field normalization (Claude snake_case | Grok camelCase).
tool_name=$(printf '%s' "$input" | jq -r '.tool_name // .toolName // empty' 2>/dev/null)
[ -n "$tool_name" ] || exit 0

command=""
case "$tool_name" in
  Bash|run_terminal_command)
    command=$(printf '%s' "$input" | jq -r '.tool_input.command // .toolInput.command // empty' 2>/dev/null)
    ;;
esac
cwd=$(printf '%s' "$input" | jq -r '.cwd // .workspaceRoot // empty' 2>/dev/null)

is_shell=0
case "$tool_name" in
  Bash|run_terminal_command) is_shell=1 ;;
esac

# ---------------------------------------------------------------------------
# (a) DENY — repo-scoped index-tool routing.
# Which index tool (if any) does this call involve?
# ---------------------------------------------------------------------------
tool=""
haystack=""
case "$tool_name" in
  Bash|run_terminal_command)
    haystack="$command"
    if printf '%s' "$command" | grep -qE '(^|[^[:alnum:]_-])codegraph([^[:alnum:]_-]|$)'; then tool="codegraph"; fi
    if printf '%s' "$command" | grep -qE '(^|[^[:alnum:]_-])jbcontext([^[:alnum:]_-]|$)'; then tool="${tool:+$tool }jbcontext"; fi
    ;;
  mcp__codegraph__*|codegraph__*)
    tool="codegraph"
    haystack=$(printf '%s' "$input" | jq -r '.tool_input.projectPath // .toolInput.projectPath // .cwd // .workspaceRoot // empty' 2>/dev/null)
    ;;
  mcp__jbcontext__*|jbcontext__*)
    tool="jbcontext"
    haystack=$(printf '%s' "$input" | jq -r '.tool_input.pathFilter // .toolInput.pathFilter // .cwd // .workspaceRoot // empty' 2>/dev/null)
    ;;
esac

if [ -n "$tool" ] && [ -f "$MAP" ]; then
  haystack="$haystack $cwd"
  # Map rows: {"repos": ["substr", ...], "deny": ["codegraph"], "reason": "..."}
  n=$(jq 'length' "$MAP" 2>/dev/null || echo 0)
  i=0
  while [ "$i" -lt "$n" ]; do
    row=$(jq -c ".[$i]" "$MAP")
    i=$((i + 1))
    for t in $tool; do
      jq -e --arg t "$t" '.deny // [] | index($t)' >/dev/null 2>&1 <<<"$row" || continue
      while IFS= read -r repo; do
        [ -n "$repo" ] || continue
        if printf '%s' "$haystack" | grep -qF "$repo"; then
          reason=$(jq -r '.reason // "contraindicated here"' <<<"$row")
          deny "bash-policy: '$t' is DENIED in repo '$repo' — $reason. Use the routed alternative (rules/tools/code-search.md)."
        fi
      done < <(jq -r '.repos[]?' <<<"$row")
    done
  done
fi

# Everything below is shell-command policy.
[ "$is_shell" -eq 1 ] || exit 0
[ -n "$command" ] || exit 0

# ---------------------------------------------------------------------------
# (b) DENY — pip ban. `uv pip …` is the sanctioned form and is stripped first.
# ---------------------------------------------------------------------------
pip_re='(^|[;&|]|sudo[[:space:]]+)[[:space:]]*pip3?([[:space:]]|$)'
pip_stripped=${command//uv pip/ }
if [[ "$command" == *"--break-system-packages"* ]] || [[ "$pip_stripped" =~ $pip_re ]]; then
  deny "bash-policy: pip is banned — use uv (uv pip install / uv run). See CLAUDE.md > Python Dependency Management."
fi

# ---------------------------------------------------------------------------
# (b2) DENY — zsh special-variable assignment. Grok's tool shell is zsh via
# eval, where `path` is tied to PATH (assigning it replaces the ENTIRE PATH —
# every later command dies with `command not found`) and `status` is read-only.
# Scoped to run_terminal_command: Claude Code's Bash tool runs bash 5 via
# CLAUDE_CODE_SHELL, where these names are ordinary variables — a deny there
# would only produce false positives. Word-boundary pattern so repo_path= and
# file_path= pass; covers plain, +=, and typeset/local/export/declare forms.
# ---------------------------------------------------------------------------
if [ "$tool_name" = "run_terminal_command" ]; then
  zsh_special_re='(^|[;&|({[:space:]])((typeset|local|export|declare)[[:space:]]+(-[A-Za-z]+[[:space:]]+)*)?(path|status)\+?='
  if [[ "$command" =~ $zsh_special_re ]]; then
    deny "bash-policy: this tool shell is zsh — 'path' is tied to PATH (assigning it wipes the entire PATH) and 'status' is read-only. Rename the variable (repo_path, dir, st, exit_status) and retry. See CLAUDE.md > Shell."
  fi
fi

# ---------------------------------------------------------------------------
# (c) DENY — package-manager mixing. Nearest lockfile decides the manager;
# only MUTATING verbs count, so read-only calls (npm view, pnpm why) pass.
# ---------------------------------------------------------------------------
lockfile_manager() {
  local dir="$1"
  while [ -n "$dir" ] && [ "$dir" != "/" ] && [ "$dir" != "." ]; do
    [ -f "$dir/pnpm-lock.yaml" ]    && { printf 'pnpm|%s/pnpm-lock.yaml' "$dir"; return 0; }
    [ -f "$dir/yarn.lock" ]         && { printf 'yarn|%s/yarn.lock' "$dir"; return 0; }
    [ -f "$dir/package-lock.json" ] && { printf 'npm|%s/package-lock.json' "$dir"; return 0; }
    [ -f "$dir/bun.lockb" ]         && { printf 'bun|%s/bun.lockb' "$dir"; return 0; }
    dir=$(dirname "$dir")
  done
  return 1
}

mutating_manager() {
  local cmd="$1"
  [[ "$cmd" =~ (^|[;&|]|[[:space:]])npm[[:space:]]+(install|i|add|update|remove)([[:space:]]|$) ]]  && { printf 'npm';  return 0; }
  [[ "$cmd" =~ (^|[;&|]|[[:space:]])yarn[[:space:]]+(add|install|remove|upgrade)([[:space:]]|$) ]] && { printf 'yarn'; return 0; }
  [[ "$cmd" =~ (^|[;&|]|[[:space:]])pnpm[[:space:]]+(add|install|update|remove)([[:space:]]|$) ]]  && { printf 'pnpm'; return 0; }
  [[ "$cmd" =~ (^|[;&|]|[[:space:]])bun[[:space:]]+(add|install)([[:space:]]|$) ]]                 && { printf 'bun';  return 0; }
  return 1
}

if cmd_pm=$(mutating_manager "$command") && [ -n "$cwd" ]; then
  if lock_hit=$(lockfile_manager "$cwd"); then
    lock_pm=${lock_hit%%|*}
    lock_path=${lock_hit#*|}
    if [ "$cmd_pm" != "$lock_pm" ]; then
      deny "bash-policy: never mix package managers — '$cmd_pm' denied, this tree is '$lock_pm' ($lock_path). Use $lock_pm. See CLAUDE.md > Package Manager."
    fi
  fi
fi

# ---------------------------------------------------------------------------
# (d) DENY — lockfile deletion. Evaluated per command segment so an unrelated
# `rm` in one segment cannot pair with a lockfile named in another.
# ---------------------------------------------------------------------------
rm_re='(^|[[:space:]])(sudo[[:space:]]+)?((git[[:space:]]+)?rm|unlink)[[:space:]]'
lock_re='(pnpm-lock\.yaml|package-lock\.json|yarn\.lock|bun\.lockb|uv\.lock|poetry\.lock|Cargo\.lock|Gemfile\.lock)'
while IFS= read -r segment; do
  [[ "$segment" =~ $rm_re ]] || continue
  [[ "$segment" =~ $lock_re ]] || continue
  deny "bash-policy: never delete or regenerate lockfiles unless the user explicitly asks (${BASH_REMATCH[1]}) — surface the problem instead. See CLAUDE.md > Package Manager."
done < <(printf '%s\n' "$command" | tr ';&|' '\n')

# ---------------------------------------------------------------------------
# (e) ADVISORY — pre-push local-quality-gate reminder. NEVER blocks.
#
# Stays silent for repos with no detectable Node quality gate (no package.json,
# or no lint/test script), and for repos where an INSTALLED lefthook pre-push
# hook already blocks deterministically.
#
# Scope note: detection is Node/package.json only (the dominant flow-build push
# stack). Python/Java repos get no reminder by design; extend here if needed.
# ---------------------------------------------------------------------------
case "$command" in
  *"git push"*) ;;
  *) exit 0 ;;
esac

repo_root=$(git rev-parse --show-toplevel 2>/dev/null || true)
root="${repo_root:-.}"
pkg="$root/package.json"
[ -f "$pkg" ] || exit 0   # no Node manifest -> nothing to remind about

# Lefthook covers this push? Both conditions must hold: the git pre-push hook is
# installed and belongs to lefthook (honoring core.hooksPath and worktrees), AND
# some lefthook config declares a top-level `pre-push:` section (1.x `commands:`
# and 2.x `jobs:` both nest under it). Config present but never `lefthook install`ed
# -> no active gate -> keep reminding.
hooks_dir=$(git -C "$root" config core.hooksPath 2>/dev/null || true)
[ -n "$hooks_dir" ] || hooks_dir=$(git -C "$root" rev-parse --git-path hooks 2>/dev/null || true)
case "$hooks_dir" in
  /*) ;;                                # absolute -> use as-is
  ?*) hooks_dir="$root/$hooks_dir" ;;   # relative -> anchor to repo root
esac
if [[ -n "$hooks_dir" && -f "$hooks_dir/pre-push" ]] \
   && grep -qi 'lefthook' "$hooks_dir/pre-push" 2>/dev/null; then
  for lh_cfg in lefthook.yml lefthook.yaml .lefthook.yml .lefthook.yaml \
                lefthook-local.yml lefthook-local.yaml; do
    if [[ -f "$root/$lh_cfg" ]] && grep -qE '^pre-push:' "$root/$lh_cfg" 2>/dev/null; then
      exit 0   # deterministic gate active -> reminder redundant
    fi
  done
fi

# Detect package manager (mirrors the global pnpm-default rule).
pm=pnpm
if   [ -f "$root/pnpm-lock.yaml" ];   then pm=pnpm
elif [ -f "$root/yarn.lock" ];        then pm=yarn
elif [ -f "$root/package-lock.json" ]; then pm=npm
fi

# Only mention gates that actually exist; bail if neither does.
has_lint=$(jq -r '.scripts.lint // empty' "$pkg" 2>/dev/null)
has_test=$(jq -r '.scripts.test // empty' "$pkg" 2>/dev/null)
[ -n "$has_lint$has_test" ] || exit 0

cmds=""
[ -n "$has_lint" ] && cmds="\`$pm lint\`"
[ -n "$has_test" ] && cmds="${cmds:+$cmds, }\`$pm test\`"

reminder="Pre-push reminder (non-blocking): you are about to \`git push\`. Confirm the local quality gate ran on the AFFECTED subset SINCE your last code edit — any edit after the last run (e.g. a post-push CI fix) invalidates it, and CI will then catch what a local check would have. Detected for this repo: ${cmds}. If you edited code since the last run, re-run them now, then push. This does not block the push."

jq -n --arg ctx "$reminder" '{
  hookSpecificOutput: {
    hookEventName: "PreToolUse",
    permissionDecision: "allow",
    additionalContext: $ctx
  }
}'
