# reviewer-guard — deterministic read-only backstop for review/audit agents

PreToolUse hook on `Bash`, **agent-scoped**: it rides the frontmatter `hooks:` block of
each agent in the roster below, never `settings.json` (there is deliberately no
`settings-config.json` in this directory — a global registration would fire on every
session's Bash). Deploy is automatic: `/deploy-global` copies every `*.sh` under
`global/hooks/**` flat to `~/.claude/hooks/`, so the frontmatter references
`$HOME/.claude/hooks/reviewer-guard.sh`.

## Why

The `tools:` allowlist of a review agent omits `Write`/`Edit`, but unrestricted `Bash`
keeps a mutation path open, and `permissionMode: plan` is ignored when the parent session
runs in auto mode (verified 2026-08-15: a `review-code` subagent created a file via
`touch` unblocked). This hook closes that gap deterministically for the agents whose
doctrine is investigate-and-report.

## Roster (strict — carries the hook)

`review-code`, `review-security`, `sdd-product-critic`, `sdd-spec-reviewer`,
`sdd-explore`, `workspace-custodian`.

**Exempt (execute-to-observe by doctrine — no hook):** `review-refuter` (runs tests and
repro commands), `sdd-verify` and `review-ux` (drive a real browser and its
session files), `state-fetcher` (executes approved tracker writes). Their read-only-ness
stays `tools:`-allowlist + prompt-convention. Rationale recorded in
`_support/docs/enforcement-layers.md`.

## What it denies

- git mutations: `commit`, `push`, `merge`, `rebase`, `reset`, `checkout`, `switch`,
  `restore`, `clean`, `stash`, `cherry-pick`, `revert`, `rm`, `mv`, `am`, `apply`,
  `pull`, `worktree` (including `git -C <path>` forms).
- Deleters anywhere: `rm`, `rmdir`, `unlink`, `shred`, `truncate`, `find -delete`/
  `-exec rm`.
- In-place edits: `sed -i`, `perl -i`.
- Package installs: npm/pnpm/yarn/bun/uv/poetry/pip/pipx/cargo/gem/brew/apt install-class
  verbs.
- Writes landing outside temp space: `>`/`>>` redirections, `tee`, `cp` (destination),
  `mv`, `touch`, `mkdir`, `ln`, `chmod`, `chown`. Temp space stays writable —
  `$TMPDIR`, `/tmp`, `/var/folders`, `/dev/null`, the session scratchpad — because
  investigation output needs a home.

Everything else passes: `rg`/`grep`/`find` (read-only), `git log/diff/show/status/fetch`,
`gh … view`, registry reads (`npm view`), interpreter runs that only read.

## Honest limits

- **Guardrail, not a sandbox.** An interpreter one-liner (`python3 -c "open('x','w')"`)
  or an unlisted binary can still write; that residue is prompt-convention (the agents'
  never-mutate clause), same layer as before — the hook removes the COMMON paths, not all
  paths.
- **Claude Code only.** `hooks:` is a Claude-only frontmatter field: `convert-agents.py`
  strips it from the generated Codex/opencode/Grok agents, and `codex exec` runs no hooks
  at all. On those harnesses the read-only doctrine remains `tools:`/permissions +
  prompt-convention (see `harness/codex/README.md`).

## Tests

`bash global/hooks/reviewer-guard/test-reviewer-guard.sh` — deny and allow cases; keep it
green and extend it with every new deny path. Named distinctly because the flat hooks
deploy would collide a second `test-dual-runtime.sh` with bash-policy's.
