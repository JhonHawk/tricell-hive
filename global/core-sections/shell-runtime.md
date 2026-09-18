---
order: 80
targets: [claude]
---

## Shell — tool runtime per harness

- **Claude Code, opencode and Pi run bash 5** (`CLAUDE_CODE_SHELL=/opt/homebrew/bin/bash` in Claude settings `env`; opencode `shell` and Pi `shellPath` are deploy-managed; all three validated by the deploy-global preflight) — write plain bash. If zsh-style errors appear (`(eval):N:` prefix, `read-only variable`, `no matches found`), the override has drifted to zsh: flag it and apply the zsh rules below until fixed.
- **Grok and Codex tool shells are zsh via eval** — and so is any shell not confirmed otherwise. For them:
  - **Never assign zsh special names as variables:** `path` (tied to `PATH` — `path=/x` replaces the entire PATH and later commands die with `command not found`), `status` (read-only), `cwd`, `argv`, `pipestatus`, `fpath`, `cdpath`, `manpath`. Use `dir`, `repo_path`, `st`, `exit_status`. Enforcement: prompt-convention here; deterministic deny for `path=`/`status=` in the `bash-policy` hook (Grok scope).
  - **No bash-4isms:** `declare -A`, `${!var}` (zsh form: `${(P)var}`), `${!arr[@]}`, `mapfile`, `read -p`.
  - Unmatched globs ERROR (`nomatch`) — **including as flag values** (`--include=*.ts`): guard the glob or use `find`/`rg`.
  - Unquoted `$var` does NOT word-split: pass file lists via `find -print0 | xargs -0` or arrays — never `cmd $list`.
  - Never unquoted `=word`/`===` as separators or arguments (zsh `=cmd` expansion).
- macOS userland is BSD for every harness (`sed -i ''`, `grep`, `awk` differ from GNU).
- Parsing `gh auth status`: never take the account from `$NF` (the last field is `(keyring)`) — take the token after `account`, or use `gh api user --jq .login`.
- Complex quoting or multiline text **in a command's arguments** → a `python3` heredoc, not heroic shell escaping. File content is authored and edited with the harness's file tools, never through the shell.
- After a state-mutating one-liner, verify the post-state — never trust the success banner (a pipeline can exit 0 having silently no-oped).
