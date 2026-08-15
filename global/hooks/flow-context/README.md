# flow-context

UserPromptSubmit context for the flow pack, delivered in flow workspaces (a `_support/PROJECT.md` ledger at or above cwd). Merges the two former UserPromptSubmit hooks — **replaces `flow-phase-context` + `flow-plan-injector`** — into one script that makes TWO INDEPENDENT emit decisions on every prompt.

## The two sections (and why they must NOT collapse)

| Section | Gate | Marker | Re-fire semantics |
|---|---|---|---|
| **Phase** | flow workspace | `…/claude-flow-context-phase-<session>` holds the ledger MTIME | Once per session, **re-fires when the ledger changes** — a phase transition mid-session must refresh the injected `Current phase` / `Next suggested` state, or the model offers the wrong `/flow-*` stage. |
| **Plan** | `permission_mode == "plan"` | `…/claude-flow-context-plan-<session>` (empty touch) | **Once per session** — a plan's artifact conventions don't change mid-plan; injecting them repeatedly is noise. |

A single shared marker would couple the two: the once-per-session plan gate would suppress the mtime-driven phase re-fire (or the phase re-fire would reset the plan gate). Two markers keep each section's cadence independent. Each section is gated, computed, and emitted separately; both can fire on the same prompt (phase first), either can fire alone, and when neither passes its gate the script emits nothing (exit 0).

The `permission_mode` read preserves both key fallbacks (`.permission_mode // .permissionMode`); absent field → the plan section stays silent and the phase section is unaffected.

Payload is STATE, not routing instructions (the retired flow-route-reminder proved per-prompt routing payloads fail). Advisory — never blocks (exit 0 always; stdout becomes model context).

## Delivery

Claude Code only (UserPromptSubmit + native plan mode are Claude Code concepts). Deployed by `/deploy-global` (script → `~/.claude/hooks/`, block merged into `settings.json` from `settings-config.json`). Codex/opencode get the flow conventions through the deployed global core (`~/.codex/AGENTS.md` / `~/.config/opencode/AGENTS.md`) plus the `flow-session-context` hook — NOT through the workspace `AGENTS.md`, which their git-root-bounded discovery never reaches from a child-repo session (it loads only when the session opens at the workspace root).

## Test payloads

The script reads UserPromptSubmit JSON on stdin and prints plain text (no JSON envelope for UserPromptSubmit).

### (i) Flow workspace, normal prompt → phase section only

```bash
tmp=$(mktemp -d)
mkdir -p "$tmp/_support"
printf '| Current phase | specs |\n- Next suggested: /flow-build\n' > "$tmp/_support/PROJECT.md"

echo "{\"cwd\":\"$tmp\",\"session_id\":\"s1\"}" | ./flow-context.sh
```

Expected: the `Flow workspace — state from the ledger …` block with `Current phase: specs` and `Next suggested step: /flow-build`. No plan section.

### (ii) Flow workspace, plan mode → both sections

```bash
echo "{\"cwd\":\"$tmp\",\"session_id\":\"s2\",\"permission_mode\":\"plan\"}" | ./flow-context.sh
```

Expected: the phase block first, then the `Flow workspace, plan mode. Conventions …` block.

### (iii) Ledger-mtime change re-fires the phase section

```bash
sid="s3"
echo "{\"cwd\":\"$tmp\",\"session_id\":\"$sid\"}" | ./flow-context.sh   # fires
echo "{\"cwd\":\"$tmp\",\"session_id\":\"$sid\"}" | ./flow-context.sh   # SILENT (same mtime)
printf '| Current stage | desarrollo |\n- Next suggested: /flow-build\n' > "$tmp/_support/PROJECT.md"
echo "{\"cwd\":\"$tmp\",\"session_id\":\"$sid\"}" | ./flow-context.sh   # RE-FIRES (mtime changed)
```

Expected: first call emits, second is silent, third emits again with the new `Current stage: desarrollo`. (The hook accepts both `Current stage` — v2 template — and `Current phase` — pre-v2 ledgers.)

### (iv) Non-flow workspace → silence

```bash
tmp2=$(mktemp -d)
echo "{\"cwd\":\"$tmp2\",\"session_id\":\"s4\"}" | ./flow-context.sh
```

Expected: no output, exit 0.
