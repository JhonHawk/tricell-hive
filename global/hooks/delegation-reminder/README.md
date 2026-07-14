# delegation-reminder

**Event:** `PostToolUse`, no matcher (every tool call). **Non-blocking** — always exits 0.

Deterministic backstop for the delegation gate in `global/rules/workflow/agent-routing.md > Delegation Gates`: the "~20 tool calls without delegating → re-plan" rule depends on the model observing itself, which degrades in long sessions. This hook counts instead.

## Mechanics

- Keeps a per-session counter at `${TMPDIR:-/tmp}/claude-delegation-reminder-<session_id>` (format `<agent_id>|<count>`).
- Every non-delegation tool call increments it; a `Task`/`Agent` tool call resets it to 0.
- On crossing each multiple of 20, emits `hookSpecificOutput.additionalContext` reminding the model to delegate the remainder or justify staying inline — mirroring the gate's escape hatch.

## Known limitations

- PostToolUse may also fire for tool calls made by subagents within the same session; `agent_id` semantics are undocumented (verified against docs 2026-07). Two mitigations: the counter only accepts events matching the session's first-seen `agent_id`, and the reset on `Task`/`Agent` prevents residual noise from accumulating across delegations. Worst case is an early reminder — acceptable for an advisory hook.
- Markers live in TMPDIR and are never cleaned up by the hook; the OS purges them.

## Smoke test

```bash
S=smoketest; rm -f "${TMPDIR:-/tmp}/claude-delegation-reminder-$S"
for i in $(seq 1 20); do
  out=$(echo "{\"session_id\":\"$S\",\"tool_name\":\"Read\"}" | bash delegation-reminder.sh)
done
echo "$out"   # expect additionalContext JSON at call 20
echo "{\"session_id\":\"$S\",\"tool_name\":\"Task\"}" | bash delegation-reminder.sh
cat "${TMPDIR:-/tmp}/claude-delegation-reminder-$S"   # expect main|0
```
