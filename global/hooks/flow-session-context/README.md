# flow-session-context

SessionStart context for the flow pack v2, delivered across all three harnesses. Merges what were two separate SessionStart hooks (the old `session-hygiene-context`) plus the flow-pack process map into ONE injection with two independent, self-gating sections:

1. **Flow protocol** — injected ONLY inside a flow workspace (a `_support/PROJECT.md` ledger at or above cwd). A static `<flow-process-protocol>` block: the v2 process chain (brainstorm → spec → plan → execute → deploy/QA) so the model can OFFER the matching `/flow-*` stage when user intent matches. It never forces a stage — flow skills stay user-gated (global CLAUDE.md > Skill Auto-invocation).
2. **Git hygiene** — injected in ANY git repo (not flow-gated). Deterministic backstop for the session-close ritual (`git-mechanics.md > Session close`): most closes are silent, so the ceremony runs at the next fresh seam. Injects pending-hygiene FACTS — local branches fully merged into the integration target, and branches whose upstream is `[gone]` — capped at 8 each, silent when none. State only, never routing instructions.

Both sections empty → no injection at all (exit 0 / no-op). Local git queries only; no fetch, no network. Advisory — never blocks.

## Delivery per harness

| Harness | Channel | File |
|---|---|---|
| Claude Code | `settings.json` hooks block (SessionStart, no matcher; script filters `startup\|clear`) | `flow-session-context.sh` → `~/.claude/hooks/` |
| Codex | `~/.codex/hooks.json` (`SessionStart` matcher `startup\|resume\|clear`) | `flow-session-context.sh` → `~/.codex/hooks/` |
| opencode | auto-loaded local plugin (`experimental.chat.system.transform`, once per session) | `flow-session-context.ts` → `~/.config/opencode/plugins/` |

The `.sh` is shared by Claude Code and Codex (identical stdin JSON schema for SessionStart). The `.ts` reimplements the same two sections for opencode, which has no SessionStart event — it injects on the first system-prompt transform of each session, gated by a per-session `Set` plus a `<flow-process-protocol>` content-marker guard against double injection.

**Codex caveat:** `~/.codex/hooks.json` scripts run only when the hook file's `trusted_hash` matches and hooks are enabled in the Codex config; a changed script must be re-trusted (or hooks re-enabled) before Codex will execute it. Deploy re-registers the hash — a manual copy does not.

Deployed by `/deploy-global` (script → `~/.claude/hooks/` + `~/.codex/hooks/`, block merged into `settings.json` from `settings-config.json`, `.ts` copied to the opencode plugins dir).

## Test payloads

The `.sh` reads SessionStart JSON on stdin and prints either nothing or a `{hookSpecificOutput:{...}}` envelope.

### (i) Flow workspace with pending git hygiene

```bash
# Setup: a dir with _support/PROJECT.md AND a git repo with a merged branch.
tmp=$(mktemp -d)
mkdir -p "$tmp/_support"
printf '| Current phase | specs |\n- Next suggested: /flow-build\n' > "$tmp/_support/PROJECT.md"
git -C "$tmp" init -q && git -C "$tmp" commit -q --allow-empty -m init
git -C "$tmp" branch feature/done   # merged into the trunk

echo "{\"source\":\"startup\",\"cwd\":\"$tmp\"}" | ./flow-session-context.sh
```

Expected: a JSON envelope whose `additionalContext` contains BOTH the `<flow-process-protocol>` block AND a `Pending git hygiene ...: local branches fully merged into master: feature/done ...` line.

### (ii) Plain git repo (no ledger)

```bash
tmp=$(mktemp -d)
git -C "$tmp" init -q && git -C "$tmp" commit -q --allow-empty -m init
git -C "$tmp" branch stale/merged

echo "{\"source\":\"clear\",\"cwd\":\"$tmp\"}" | ./flow-session-context.sh
```

Expected: a JSON envelope with ONLY the git-hygiene section (no `<flow-process-protocol>`).

### (iii) Non-flow, non-git dir → silence

```bash
tmp=$(mktemp -d)
echo "{\"source\":\"startup\",\"cwd\":\"$tmp\"}" | ./flow-session-context.sh
```

Expected: no output, exit 0.

### Source gating

```bash
echo '{"source":"resume","cwd":"/any"}' | ./flow-session-context.sh   # -> nothing (only startup|clear fire)
```
