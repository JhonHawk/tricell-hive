# post-tool-hub

**Event:** `PostToolUse`, no matcher (every tool call). **Non-blocking** — always exits 0.

One process carrying every post-execution advisory. Absorbs `delegation-reminder`, `verification-loop-reminder`, and `code-search-routing`'s POST mode into three internal sections sharing a single stdin read and a single `jq` emission.

## Sections

Each section self-gates on `tool_name` and returns text; whatever fires is newline-joined into **one** `additionalContext` emission per event.

**0. Plan-capture recovery** (first tool call of a session) — the only section that WRITES rather than advises, and it delegates the write: it runs `flow-plan-capture.sh --from-transcript` (resolved as a sibling of this script) and forwards its text. Approving a plan with **"Yes, clear context…"** denies `ExitPlanMode`, so `PostToolUse:ExitPlanMode` never fires and the plan is never captured; the fresh session's first user turn carries it in `planContent`, and this is the earliest event that can see it. It lives here rather than in its own registration because this hub already spawns on every tool call — a dedicated entry would double the per-call process cost to catch a once-per-session condition. Rationale and evidence: `flow-plan-capture/README.md > Why a recovery mode exists`.

**1. Delegation counter** (all tools) — every non-delegation tool call increments a per-session counter; a `Task`/`Agent` call resets it to 0; a reminder fires on each multiple of 20. Keyed to the session's first-seen `agent_id` so subagent tool calls (same `session_id`) don't inflate the main thread's count. Policy: `agent-routing.md > Delegation Gates`.

**2. Full-suite run counter** (`Bash` only) — counts repeated whole-suite verification runs (`turbo run test`, `pnpm [-r] [run] test`, followed only by flags). A command carrying `--filter` is the sanctioned affected-subset run and never counts; a non-flag argument (`pnpm test messages.spec`) or a scoped script (`pnpm test:unit`) is treated as scoped. The first run is the legitimate merge-boundary gate and stays silent; run #2 onward reminds. Policy: `testing.md > Execution Scope`.

**4. zsh-signature teacher** (`Bash`/`run_terminal_command`) — fires when a shell tool result carries a zsh dialect error: `read-only variable`, `no matches found`, or `bad substitution` **behind zsh's own `zsh:N:` / `(eval):N:` prefix**, or that prefix alone. Under Grok's zsh runtime it names the fix; under Claude's Bash tool the same signature means the `CLAUDE_CODE_SHELL` override drifted, which IS the alarm. Policy: `CLAUDE.md > Shell`.

The prefix is load-bearing, not decoration. Matching the bare phrase made any output that merely *contained* it fire — a `cat` or `git diff` of the rules documenting these very signatures was enough, and it happened repeatedly. Reading only `stderr` is not the alternative: Claude's Bash tool merges the command's stderr into `stdout`, so a real failure has no separate channel. The section also collects the response's strings with `[.. | strings] | join("\n")` rather than `tostring` — the latter re-escapes newlines as the two characters `\` `n`, which would make the character before `zsh:` read as alphanumeric and kill the anchor.

## State files — separate by design

| File | Section |
|---|---|
| `${TMPDIR:-/tmp}/claude-flow-plan-recovery-<session_id>` | 0 (empty; presence = attempted) |
| `${TMPDIR:-/tmp}/claude-delegation-reminder-<session_id>` | 1 (format `<agent_id>\|<count>`) |
| `${TMPDIR:-/tmp}/claude-verification-loop-<session_id>` | 2 (plain integer) |

The counters are never coupled: a `Task` call resetting section 1 must not touch section 2's count. Corrupt or missing counter files reset to 0. Markers live in TMPDIR and are never cleaned up by the hook; the OS purges them.

Section 0's marker is written **before** its attempt, so a failing capture costs one try, not one per tool call. It cannot be the whole guard: a `/resume` replays the same plan under a new session id, so `flow-plan-capture` carries its own content guard.

## Enforcement layer

Deterministic delivery of prompt-convention reminders. The hook injects signals and never fails a tool call; the cited rules own what to do about them.

## Known limitations

- `PostToolUse` may also fire for subagent tool calls within the same session; `agent_id` semantics are undocumented (verified against docs 2026-07). Section 1 mitigates via first-seen-`agent_id` keying plus the reset on delegation. Sections 2 and 4 do not filter by agent — matching their pre-merge behavior. Section 0 is session-scoped on purpose: a subagent's tool call carries the same `session_id` and the same main `transcript_path`, so whichever call comes first is the right one to capture from.
- Section 2 recognizes only the pnpm/turbo shapes above; other runners (jest, vitest, gradle, pytest) are silent by design.

## Deploy

> **WARNING — this consolidation needs `/deploy-global --apply --delete-orphans`.** A plain apply merges the new block but leaves the absorbed hooks' entries in `~/.claude/settings.json` and their scripts in `~/.claude/hooks/`, so `delegation-reminder.sh`, `verification-loop-reminder.sh`, and `code-search-routing.sh` would keep firing alongside this one — duplicate counters and doubled reminders. `--delete-orphans` purges both.

## Smoke test

```bash
S=smoketest; T=${TMPDIR:-/tmp}; rm -f "$T"/claude-delegation-reminder-$S "$T"/claude-verification-loop-$S
hub() { jq -n --arg s "$S" --arg t "$1" --arg c "${2:-}" \
          '{session_id:$s,tool_name:$t,tool_input:{command:$c}}' | bash post-tool-hub.sh; }

for i in $(seq 1 20); do out=$(hub Read); done; echo "$out"  # delegation reminder at 20
hub Task > /dev/null; cat "$T/claude-delegation-reminder-$S" # expect main|0
hub Bash "pnpm test" > /dev/null                             # silent (run #1)
hub Bash "pnpm test"                                         # full-suite reminder, run #2
```
