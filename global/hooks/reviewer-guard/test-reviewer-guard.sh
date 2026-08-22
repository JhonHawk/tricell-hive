#!/usr/bin/env bash
# Checks for reviewer-guard.sh (not deployed as a hook — the flat deploy copies
# every *.sh, so the distinct name avoids colliding with bash-policy's test).
# Run from anywhere: bash global/hooks/reviewer-guard/test-reviewer-guard.sh
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$ROOT/reviewer-guard.sh"
fail=0

run_case() {
  local name="$1" expect_exit="$2" cmd="$3"
  local out rc payload
  payload=$(jq -n --arg c "$cmd" '{tool_name: "Bash", tool_input: {command: $c}, cwd: "/tmp"}')
  out=$(printf '%s' "$payload" | bash "$SCRIPT" 2>&1) || rc=$?
  rc=${rc:-0}
  if [ "$rc" -ne "$expect_exit" ]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$name" "$expect_exit" "$rc" "$out"
    fail=$((fail + 1))
    return
  fi
  printf 'ok   %s (exit %s)\n' "$name" "$rc"
}

# --- denied ---
run_case "git-commit"            2 'git commit -m "fix it"'
run_case "git-push"              2 'git push origin feature'
run_case "git-checkout-dot"      2 'git checkout .'
run_case "git-restore"           2 'git restore src/app.ts'
run_case "git-dash-C-reset"      2 'git -C /some/repo reset --hard HEAD~1'
run_case "rm"                    2 'rm -rf build/'
run_case "xargs-rm"              2 'find . -name "*.tmp" | xargs rm'
run_case "find-delete"           2 'find . -name "*.log" -delete'
run_case "sed-in-place"          2 'sed -i "" "s/foo/bar/" src/app.ts'
run_case "perl-in-place"         2 'perl -pi -e "s/foo/bar/" src/app.ts'
run_case "pnpm-add"              2 'pnpm add lodash'
run_case "npm-install-pkg"       2 'npm install left-pad'
run_case "brew-install"          2 'brew install jq'
run_case "redirect-repo-file"    2 'rg -n "TODO" src/ > findings.txt'
run_case "append-repo-file"      2 'echo done >> notes.md'
run_case "tee-repo-file"         2 'rg -n auth src/ | tee findings.txt'
run_case "cp-into-repo"          2 'cp /tmp/patch.diff src/app.ts'
run_case "mv-repo-file"          2 'mv src/app.ts src/app.bak'
run_case "touch-repo-file"       2 'touch marker.txt'

# --- allowed ---
run_case "rg"                    0 'rg -n "auth" src/'
run_case "git-log"               0 'git log --oneline -20'
run_case "git-diff"              0 'git --no-pager diff HEAD~1 -- src/'
run_case "git-show"              0 'git show origin/main:src/app.ts'
run_case "git-fetch"             0 'git fetch origin'
run_case "git-status"            0 'git status --short'
run_case "redirect-tmp"          0 'rg -n "TODO" src/ > /tmp/findings.txt'
run_case "redirect-var-folders"  0 'rg -n "TODO" src/ > /var/folders/ab/x/T/out.txt'
run_case "redirect-scratchpad"   0 'rg -n "TODO" src/ > /private/tmp/claude-501/x/scratchpad/out.txt'
run_case "redirect-devnull"      0 'npm view lodash version 2>/dev/null'
run_case "stderr-merge"          0 'pnpm why lodash 2>&1'
run_case "tee-tmp"               0 'rg -n auth src/ | tee /tmp/findings.txt'
run_case "cp-to-tmp"             0 'cp src/app.ts /tmp/app-copy.ts'
run_case "mkdir-tmp"             0 'mkdir -p /tmp/review-workdir'
run_case "npm-view"              0 'npm view lodash versions'
run_case "gh-pr-view"            0 'gh pr view 42 --json files'
run_case "heredoc-read"          0 'python3 - <<EOF
print("analysis only")
EOF'

if [ "$fail" -ne 0 ]; then
  printf '\n%d case(s) failed\n' "$fail"
  exit 1
fi
printf '\nall reviewer-guard cases passed\n'
