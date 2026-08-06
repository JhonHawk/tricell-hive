# verification-loop-reminder

**Event:** `PostToolUse`, matcher `Bash`. **Non-blocking** — always exits 0.

Deterministic backstop for `global/rules/quality/testing.md > Execution Scope`: "a green full gate is not re-earned per fix". Sessions drift into re-running the full suite to accumulate confidence (`pnpm turbo run test --force`, four times, same green result) — wall-clock spent for no new information. The rule depends on the model noticing its own repetition; this hook counts instead.

## Detection

A command is classified as a **full-suite run** when it matches either shape and is followed only by flags:

- `turbo run test` — including via a package runner (`pnpm turbo run test`, `npx turbo run test`)
- `pnpm test`, `pnpm run test`, `pnpm -r test`, `pnpm -r run test`

**Exempt** — never counted:

- Any command containing `--filter` — that is the sanctioned affected-subset run.
- A run followed by a non-flag argument (`pnpm test messages.spec`) or a scoped script (`pnpm test:unit`).
- Everything else (`pnpm build`, `pnpm vitest related …`, other runners).

The classification is deliberately conservative: false negatives are acceptable, false positives are not. Ambiguous commands stay silent.

## Mechanics

- Keeps a per-session counter at `${TMPDIR:-/tmp}/claude-verification-loop-<session_id>` (a plain integer); a corrupt or missing file resets to 0.
- Increments once per full-suite match. The first run is the legitimate merge-boundary gate and stays silent.
- From the 2nd match onward, emits `hookSpecificOutput.additionalContext` naming the run number and pointing at the rule.

## Enforcement layer

Deterministic non-blocking reminder — it injects a signal and never fails the tool call. `testing.md > Execution Scope` owns the policy (what to re-run, when repeating a green check is legitimate); this hook only reports that repetition happened.

## Known limitations

- Counters live in TMPDIR and are never cleaned up by the hook; the OS purges them.
- Only the pnpm/turbo shapes above are recognized. Other runners (jest, vitest, gradle, pytest) are silent by design — extend the regexes if a stack warrants it.

## Smoke test

```bash
S=smoketest; rm -f "${TMPDIR:-/tmp}/claude-verification-loop-$S"
payload() { jq -n --arg s "$S" --arg c "$1" '{session_id:$s,tool_name:"Bash",tool_input:{command:$c}}'; }

payload "pnpm turbo run test --force" | bash verification-loop-reminder.sh   # silent (run #1)
payload "pnpm turbo run test --force" | bash verification-loop-reminder.sh   # additionalContext, run #2
payload "pnpm turbo run test --filter=web" | bash verification-loop-reminder.sh  # silent, not counted
cat "${TMPDIR:-/tmp}/claude-verification-loop-$S"   # expect 2
```
