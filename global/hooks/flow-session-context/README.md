# flow-session-context

SessionStart context for the flow pack v2, delivered across all three harnesses. Merges what were two separate SessionStart hooks (the old `session-hygiene-context`) plus the flow-pack process map into ONE injection. On a **fresh context** (`startup|clear`) it emits two independent, self-gating sections; on **`compact`** it emits a third, narrower payload instead.

1. **Flow protocol** — injected ONLY inside a flow workspace (a `_support/PROJECT.md` ledger at or above cwd). A static `<flow-process-protocol>` block: the intent→playbook map (idea → spec → plan → execute → deploy/QA) so the model can work the matching playbook, or OFFER `/flow-build`, when user intent matches. It never forces a stage — `/flow-build` stays user-gated (global CLAUDE.md > Skill Auto-invocation).
2. **Git hygiene** — injected in ANY git repo (not flow-gated). Deterministic backstop for the end-of-work hygiene ritual (`git-mechanics.md > End-of-work hygiene`): most closes are silent, so the ceremony runs at the next fresh seam. Injects pending-hygiene FACTS — local branches fully merged into the integration target, and branches whose upstream is `[gone]` — capped at 8 each, silent when none. State only, never routing instructions.
3. **Post-compaction recovery** (`source == compact`, flow workspaces only) — see below.

Nothing applies → no injection at all (exit 0 / no-op). Local git queries only; no fetch, no network. Advisory — never blocks.

## Why `compact` fires at all

The process map is the **only trigger the SPEC step has**. Unlike EXECUTE — which is anchored to a detectable artifact (a captured plan with `Status: planned|building`, scanned by `flow-context`) — a *missing* spec leaves no trace on disk, so nothing but this injected map tells the model that a decided idea should be formalized. A compaction that drops the block silently removes that trigger for the rest of the session.

`flow-context` (UserPromptSubmit) does not compensate: its phase marker is keyed on `session_id` + a `cksum` of the pending-plan set. A compaction changes neither, so `phase_fire=0` and the pending-plan state is never re-emitted either. **Both signals are lost at once** — hence this hook clears both `flow-context` markers (`claude-flow-context-phase-<session>` and `claude-flow-context-plan-<session>`) so that state re-emits on the next prompt, with no duplicated scan logic here.

The compact payload is **deliberately not the startup payload**. It is recovery, not a reload: the condensed intent→playbook map only. The bootstrap / migration / hygiene playbook catalogue and the git-hygiene section stay out — a end-of-work-hygiene backstop is noise mid-task, and re-paying the full map costs context exactly when it is scarcest. Pattern borrowed from the Engram plugin, which routes `compact` to a separate recovery payload rather than replaying its session-start load.

## Delivery per harness

| Harness | Channel | File |
|---|---|---|
| Claude Code | `settings.json` hooks block (SessionStart, no matcher; script filters `startup\|clear\|compact`) | `flow-session-context.sh` → `~/.claude/hooks/` |
| Codex | `~/.codex/hooks.json` (`SessionStart` matcher `startup\|clear\|compact`) | `flow-session-context.sh` → `~/.codex/hooks/` |
| opencode | auto-loaded local plugin (`experimental.chat.system.transform`, once per session) | `flow-session-context.ts` → `~/.config/opencode/plugins/` |

One settings entry, not two: `deploy-global`'s merge keys on the inner `.command`, so both topologies merge cleanly — branching inside the script keeps the ledger walk and the git section on a single path. Whether Codex emits a `compact` source is **unverified**; adding it to the matcher is inert if it never does.

**opencode has no post-compaction recovery.** Its plugin gates on a per-session `Set`, and opencode exposes no compaction signal to `experimental.chat.system.transform`. Inverting the gate to presence-based (re-inject whenever the `<flow-process-protocol>` marker is absent from the system array) would cover it organically, but only if opencode reliably passes transformed system arrays back through — unverified, and re-injecting on every step is the failure mode if it does not. Left as-is deliberately; revisit with a verified answer.

The `.sh` is shared by Claude Code and Codex (identical stdin JSON schema for SessionStart). The `.ts` reimplements sections 1–2 for opencode, gated by a per-session `Set` plus a `<flow-process-protocol>` content-marker guard against double injection.

**Codex caveat:** `~/.codex/hooks.json` scripts run only when the hook file's `trusted_hash` matches and hooks are enabled in the Codex config; a changed script must be re-trusted (or hooks re-enabled) before Codex will execute it. Deploy re-registers the hash — a manual copy does not.

Deployed by `/deploy-global` (script → `~/.claude/hooks/` + `~/.codex/hooks/`, block merged into `settings.json` from `settings-config.json`, `.ts` copied to the opencode plugins dir).

## Test payloads

The `.sh` reads SessionStart JSON on stdin and prints either nothing or a `{hookSpecificOutput:{...}}` envelope.

### (i) Flow workspace with pending git hygiene

```bash
# Setup: a dir with _support/PROJECT.md AND a git repo with a merged branch.
tmp=$(mktemp -d)
mkdir -p "$tmp/_support"
printf '| Current phase | specs |\n' > "$tmp/_support/PROJECT.md"
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

### (iii) Post-compaction in a flow workspace

```bash
tmp=$(mktemp -d); mkdir -p "$tmp/_support"; printf 'x\n' > "$tmp/_support/PROJECT.md"
: > "${TMPDIR:-/tmp}/claude-flow-context-phase-testS"
: > "${TMPDIR:-/tmp}/claude-flow-context-plan-testS"

echo "{\"source\":\"compact\",\"cwd\":\"$tmp\",\"session_id\":\"testS\"}" | ./flow-session-context.sh
ls "${TMPDIR:-/tmp}"/claude-flow-context-*-testS 2>/dev/null   # -> both gone
```

Expected: the `<flow-process-protocol source="post-compaction">` block ONLY (no git-hygiene section), and both `flow-context` markers deleted.

### (iv) Post-compaction outside a flow workspace → silence

```bash
echo '{"source":"compact","cwd":"/tmp","session_id":"x"}' | ./flow-session-context.sh
```

Expected: no output, exit 0 — the recovery payload is flow-gated.

### (v) Non-flow, non-git dir → silence

```bash
tmp=$(mktemp -d)
echo "{\"source\":\"startup\",\"cwd\":\"$tmp\"}" | ./flow-session-context.sh
```

Expected: no output, exit 0.

### Source gating

```bash
echo '{"source":"resume","cwd":"/any"}' | ./flow-session-context.sh   # -> nothing (resume: context already has it)
```
