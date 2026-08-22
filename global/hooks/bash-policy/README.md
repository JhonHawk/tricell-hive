# bash-policy

**Event:** `PreToolUse`, matcher `Bash|run_terminal_command`.

One process carrying every pre-execution shell policy gate. Absorbs the former `code-search-routing` (PRE mode) and `pre-push-lint-reminder`, and adds three universal denies. Consolidation goal: a Bash call spawns 2 hook processes instead of 4-5, and the stdin/`jq` boilerplate is paid once.

## Sections — order is load-bearing

Every DENY runs before any advisory, so no JSON is ever emitted before a block. Each deny ends with `exit 2` plus a one-line stderr naming the rule and the fix.

| # | Section | Semantics | Applies to |
|---|---|---|---|
| a | Repo-scoped index routing | deny | Bash + MCP index tools |
| b | pip ban | deny | Bash |
| c | Package-manager mixing | deny | Bash |
| d | Lockfile deletion | deny | Bash |
| e | Protected environment branches (production, qa) | deny | Bash |
| f | Pre-push quality-gate reminder | advisory | Bash |

Non-Bash (MCP index) tools see section (a) only — (b)-(f) are shell-command policy.

**(a) Repo-scoped index routing.** Ported verbatim from `code-search-routing.sh`, including its JSON map (renamed `bash-policy.json`, same schema: `{"repos": [substr…], "deny": [tool…], "reason": "…"}`). Denies an index tool in repos where it is contraindicated; extend the map with evidence, not intuition. Policy: `rules/tools/code-search.md`.

**(b) pip ban.** Denies `pip`/`pip3` in command position (start, or after `;`, `&&`, `|`, `sudo`) and any `--break-system-packages`. `uv pip …` is stripped before the check — it is the sanctioned form. Policy: `CLAUDE.md > Python Dependency Management`.

**(c) Package-manager mixing.** Walks up from the hook input's `.cwd` to the nearest directory holding `pnpm-lock.yaml` / `yarn.lock` / `package-lock.json` / `bun.lockb`, and denies a MUTATING verb of a *different* manager (`npm install|i|add|update|remove`, `yarn add|install|remove|upgrade`, `pnpm add|install|update|remove`, `bun add|install`). Read-only calls (`npm view`, `npm ls`, `npm audit`, `yarn info`, `pnpm why`) are not in the verb list and never deny. No lockfile found → allow. Policy: `CLAUDE.md > Package Manager`.

**(d) Lockfile deletion.** Denies `rm` / `git rm` / `unlink` naming any of `pnpm-lock.yaml`, `package-lock.json`, `yarn.lock`, `bun.lockb`, `uv.lock`, `poetry.lock`, `Cargo.lock`, `Gemfile.lock`. Evaluated **per command segment** (split on `;`, `&`, `|`) so an unrelated `rm` in one segment cannot pair with a lockfile merely named in another — `rm dist/x && cat pnpm-lock.yaml` passes, `rm pnpm-lock.yaml` does not.

**(e) Protected environment branches.** Denies `git commit` while the affected repo's current branch is `production` or `qa`, and `git push` whose destination is one of them — explicit refspec (`git push origin production`, `git push origin HEAD:qa`) or a bare push while standing on the branch. The repo resolves from `git -C <path>` when present, else from the hook cwd (with per-segment `cd` tracking); the branch from `git symbolic-ref --short HEAD` — detached HEAD and non-repo paths allow. `master`/`main` are deliberately OUT of the set: trunk-direct repos are a declared workflow, and that half stays prompt-convention. An explicit non-protected refspec pushed while standing on `qa` is allowed — the commit lands where the refspec says. This is the deterministic backstop of the promotion confirm-gate: the deny message routes the agent to a work branch or to the USER's confirmation (`git-workflow.md > Safety gates`); the agent never self-confirms.

**(f) Pre-push reminder.** Ported verbatim from `pre-push-lint-reminder.sh` — same `git push` filter, same lefthook-aware suppression, same package-manager detection, same message.

## Scoped allow — why `permissionDecision` appears only in (f)

The push advisory emits `permissionDecision: "allow"` because that is what it always did, and it is safe there: the branch is only reachable for a command containing `git push`. Emitting that field anywhere higher in the script would auto-approve **arbitrary Bash**, silently disabling the permission prompt for every shell command in the session. The field must stay inside the push branch.

## Deny regexes

```
pip        (^|[;&|]|sudo[[:space:]]+)[[:space:]]*pip3?([[:space:]]|$)     (after stripping `uv pip`)
npm        (^|[;&|]|[[:space:]])npm[[:space:]]+(install|i|add|update|remove)([[:space:]]|$)
yarn       (^|[;&|]|[[:space:]])yarn[[:space:]]+(add|install|remove|upgrade)([[:space:]]|$)
pnpm       (^|[;&|]|[[:space:]])pnpm[[:space:]]+(add|install|update|remove)([[:space:]]|$)
bun        (^|[;&|]|[[:space:]])bun[[:space:]]+(add|install)([[:space:]]|$)
rm         (^|[[:space:]])(sudo[[:space:]]+)?((git[[:space:]]+)?rm|unlink)[[:space:]]
lockfiles  (pnpm-lock\.yaml|package-lock\.json|yarn\.lock|bun\.lockb|uv\.lock|poetry\.lock|Cargo\.lock|Gemfile\.lock)
```

The leading `(^|[;&|]|[[:space:]])` guard is what keeps `pnpm install` from matching the `npm` pattern.

## Known gaps (deliberate — extend only with evidence)

- `python -m pip install …` is not denied; only the direct `pip`/`pip3` invocation forms are.
- `npm ci` is not in the mutating-verb list, so it does not trip section (c).
- Section (e) covers `production`/`qa` only; a project-declared custom protected branch is
  confirm-gated by `git-workflow.md`, not by this hook. `git merge` into a protected branch
  while standing on it is caught only via the commit/push it implies, not as a verb.
- Section (f) detects Node quality gates only (`package.json` scripts); Python/Java repos get no reminder.

## Enforcement layers

Sections (a)-(e) are **deterministic** (hook `exit 2` blocks the call). Section (f) is a deterministic delivery of a prompt-convention reminder — it never blocks.

**Harness reach:** Claude Code (settings hook) and Grok (settings compat merge; the script
parses both payload shapes). **Codex runs no hooks at all** (`codex exec` has no hook
runtime), and Cursor only what its hook surface supports — there the same gates are
prompt-convention plus the sandbox/permission layer of that harness.

## State

None. This hook is stateless; the only external input is the map at `~/.claude/hooks/bash-policy.json`.

## Deploy

> **WARNING — this consolidation needs `/deploy-global --apply --delete-orphans`.** A plain apply merges the new block but leaves the absorbed hooks' entries in `~/.claude/settings.json` and their scripts in `~/.claude/hooks/`, so `code-search-routing.sh` and `pre-push-lint-reminder.sh` would keep firing alongside this one — double denies and duplicate reminders. `--delete-orphans` purges both.

## Smoke test

```bash
t() { jq -n --arg c "$1" --arg d "${2:-$PWD}" '{tool_name:"Bash",cwd:$d,tool_input:{command:$c}}' \
        | bash bash-policy.sh; echo "rc=$?"; }

t "pip install requests"          # rc=2, pip ban
t "uv pip install requests"       # rc=0
t "npm install"  /path/to/pnpm-repo   # rc=2, manager mixing
t "pnpm why react" /path/to/pnpm-repo # rc=0, read-only
t "rm pnpm-lock.yaml"             # rc=2, lockfile deletion
t "rm -rf node_modules && pnpm install"  # rc=0
```
