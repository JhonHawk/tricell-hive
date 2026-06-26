#!/usr/bin/env bash
# pre-push-lint-reminder.sh — PreToolUse (Bash) hook, NON-BLOCKING.
#
# When the agent is about to run `git push`, injects a one-shot reminder to
# re-run the local quality gate (lint + affected tests) on any code edited
# SINCE the last run — the exact failure mode where a post-push fix gets
# pushed un-linted and CI catches what a 3-second local check would have.
#
# It NEVER blocks the push (permissionDecision: "allow", exit 0). It stays
# silent for every other command and for repos with no detectable Node
# quality gate (no package.json, or no lint/test script) — so it is noise-free
# in config-only or non-Node repos.
#
# Scope note: detection is Node/package.json only (the dominant flow-build push
# stack). Python/Java repos get no reminder by design; extend here if needed.

set -uo pipefail

input=$(cat)
command=$(printf '%s' "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)

# React only to git push; silent for everything else.
case "$command" in
  *"git push"*) ;;
  *) exit 0 ;;
esac

repo_root=$(git rev-parse --show-toplevel 2>/dev/null || true)
root="${repo_root:-.}"
pkg="$root/package.json"
[ -f "$pkg" ] || exit 0   # no Node manifest -> nothing to remind about

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
