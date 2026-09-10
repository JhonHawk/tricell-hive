# post-tool-hub

**Event:** `PostToolUse`, no matcher (every tool call). **Non-blocking** — the canonical
script always exits 0. Pi normalizes its post-tool payload before invoking the script;
the Pi extension handles required-hook failures separately.

One process carrying every post-execution advisory. Absorbs `delegation-reminder`, `verification-loop-reminder`, and `code-search-routing`'s POST mode into three internal sections sharing a single stdin read and a single `jq` emission.

## Sections

Each section self-gates on `tool_name` and returns text; whatever fires is newline-joined into **one** `additionalContext` emission per event.

**0. Plan-capture recovery** (first tool call of a session) — the only section that WRITES rather than advises, and it delegates the write: it runs `flow-plan-capture.sh --from-transcript` (resolved as a sibling of this script) and forwards its text. This section is skipped for `harness: "pi"`; Pi captures the settled plan explicitly through `--from-pi-command`, after its own `/hive-plan approve` gate. Approving a plan with **"Yes, clear context…"** denies `ExitPlanMode`, so `PostToolUse:ExitPlanMode` never fires and the plan is never captured; the fresh session's first user turn carries it in `planContent`, and this is the earliest event that can see it. It lives here rather than in its own registration because this hub already spawns on every tool call — a dedicated entry would double the per-call process cost to catch a once-per-session condition. Rationale and evidence: `flow-plan-capture/README.md > Why a recovery mode exists`.

**1. Delegation counter** (all tools) — every non-delegation tool call increments a per-session counter; a `Task`/`Agent`/`subagent` call resets it to 0; a reminder fires on each multiple of 20. Keyed to the session's first-seen `agent_id` so subagent tool calls (same `session_id`) don't inflate the main thread's count. Policy: `agent-routing.md > Delegation Gates`.

**2. Full-suite run counter** (`Bash` only) — counts repeated whole-suite verification runs (`turbo run test`, `pnpm [-r] [run] test`, followed only by flags). A command carrying `--filter` is the sanctioned affected-subset run and never counts; a non-flag argument (`pnpm test messages.spec`) or a scoped script (`pnpm test:unit`) is treated as scoped. The first run is the legitimate merge-boundary gate and stays silent; run #2 onward reminds. Policy: `testing.md > Execution Scope`.

**4. zsh-signature teacher** (`Bash`/`run_terminal_command`) — fires when a shell tool result carries a zsh dialect error: `read-only variable`, `no matches found`, or `bad substitution` **behind zsh's own `zsh:N:` / `(eval):N:` prefix**, or that prefix alone. Under Grok's zsh runtime it names the fix; under Claude's Bash tool the same signature means the `CLAUDE_CODE_SHELL` override drifted, which IS the alarm. Pi receives the short shell-standards reminder; it does not reuse the Claude-specific shell override wording. Policy: `CLAUDE.md > Shell`.

**5. Git-mode ask advisory** (`Write`/`Edit`/`MultiEdit`, plus it watches `AskUserQuestion`) — the session-mode question (`git-mechanics.md > Commits`) is mandatory at first edit-intent, and real sessions have skipped or shrunk it. Every observed `AskUserQuestion` call touches a sighting marker; the session's **first** `Write`/`Edit`/`MultiEdit` whose `cwd` is inside a git repo consumes a once-per-session evaluation — no prior sighting → one advisory naming the rule; prior sighting, or any later edit → silent. **Claude-shaped payloads only** (snake_case `tool_name`): `AskUserQuestion` is Claude's ask tool, and Grok's equivalent is not observably named here, so firing on Grok payloads would false-positive right after a legitimate ask — Grok stays out of scope, covered by the rule text alone. Pi's structured question coordinator and plan extension own that gate, so PI payloads are skipped here. Heuristic by design (a question the hook could not observe is not counted; the message says so). Policy: `git-mechanics.md > Commits`.

**6. Remote-apply counter** (`Bash` only) — counts remote apply/deploy/migration **attempts**: `terraform … apply`, `gh workflow run`, `gh run rerun`, `aws ecs update-service`, `vercel|flyctl|fly|dokploy deploy`, and a `git push` whose refspec names an environment branch (`qa`, `staging`, `prod`, …). It counts attempts and never failures on purpose — reading failure from output would mean parsing it, and one failed run is polled repeatedly while it is investigated, so a single failure would inflate the count. It fires from attempt **#3**, not #2, because the breaker ends remote execution at the second *failed* attempt and this hook cannot tell success from failure; the message states the condition instead of asserting it. Policy: `debugging.md` (remote-apply breaker); the preventive half is `devops-principles.md > observation path`.

The prefix is load-bearing, not decoration. Matching the bare phrase made any output that merely *contained* it fire — a `cat` or `git diff` of the rules documenting these very signatures was enough, and it happened repeatedly. Reading only `stderr` is not the alternative: Claude's Bash tool merges the command's stderr into `stdout`, so a real failure has no separate channel. The section also collects the response's strings with `[.. | strings] | join("\n")` rather than `tostring` — the latter re-escapes newlines as the two characters `\` `n`, which would make the character before `zsh:` read as alphanumeric and kill the anchor.

## Pi adapter boundary

The Pi extension sends this hub a normalized payload with `harness: "pi"`, the same
`session_id`/`tool_name`/`tool_input` field names used by the canonical script, and the
child's tool result. The hub remains advisory: it can add context or a warning, but it
does not approve plans, change the Pi tool set, or identify a child role. The Pi runtime
owns those decisions through `/hive-plan`, the structured question coordinator, and the
generated role definition.

For Pi, section 0 is skipped because plan capture is an explicit command boundary; section
1 recognizes `subagent` as a delegation reset; section 5 is skipped because the structured
question and Git-mode state are tracked by the extension. A missing, timed-out, non-zero,
or malformed result from a required blocking hook is handled by the runtime as a block. An
advisory hook failure becomes a warning and does not stop the child.

## State files — separate by design

| File | Section |
|---|---|
| `${TMPDIR:-/tmp}/claude-flow-plan-recovery-<session_id>` | 0 (empty; presence = attempted) |
| `${TMPDIR:-/tmp}/claude-delegation-reminder-<session_id>` | 1 (format `<agent_id>\|<count>`) |
| `${TMPDIR:-/tmp}/claude-verification-loop-<session_id>` | 2 (plain integer) |
| `${TMPDIR:-/tmp}/claude-askq-seen-<session_id>` | 5 (empty; presence = an AskUserQuestion was observed) |
| `${TMPDIR:-/tmp}/claude-git-mode-ask-<session_id>` | 5 (empty; presence = the once-per-session evaluation ran) |
| `${TMPDIR:-/tmp}/claude-remote-apply-<session_id>` | 6 (plain integer) |

The counters are never coupled: a `Task` call resetting section 1 must not touch section 2's count. Corrupt or missing counter files reset to 0. Markers live in TMPDIR and are never cleaned up by the hook; the OS purges them.

Section 0's marker is written **before** its attempt, so a failing capture costs one try, not one per tool call. It cannot be the whole guard: a `/resume` replays the same plan under a new session id, so `flow-plan-capture` carries its own content guard.

## Enforcement layer

Deterministic delivery of prompt-convention reminders. The hook injects signals and never fails a tool call; the cited rules own what to do about them.

## Known limitations

- `PostToolUse` may also fire for subagent tool calls within the same session; `agent_id` semantics are undocumented (verified against docs 2026-07). Section 1 mitigates via first-seen-`agent_id` keying plus the reset on delegation. Sections 2 and 4 do not filter by agent — matching their pre-merge behavior. Section 0 is session-scoped on purpose: a subagent's tool call carries the same `session_id` and the same main `transcript_path`, so whichever call comes first is the right one to capture from.
- Section 2 recognizes only the pnpm/turbo shapes above; other runners (jest, vitest, gradle, pytest) are silent by design.
- Section 6 sees only the command shapes above. A bare `git push` from an already-checked-out environment branch, a `gh pr merge` that triggers a deploy, and a console-driven apply are all invisible to it — false negatives it accepts, since a false positive would nag a session whose deploys all succeeded.
- Section 5 is Claude-only and heuristic: it counts only `AskUserQuestion` calls PostToolUse delivered to this hook (which fire after the question is answered). A mode question asked before the hook existed in the session, or on a harness whose ask tool it cannot name (Grok), is invisible to it — hence "does not appear to have been asked" in the message, never a claim that it was not. Pi owns the equivalent state in its structured question coordinator and is intentionally skipped here.

## Deploy

> **WARNING — this consolidation needs the orphan purge that `/deploy-global --apply` performs by default.** The purge removes the absorbed hooks' entries from `~/.claude/settings.json` and their scripts from `~/.claude/hooks/`; an apply run with `--keep-orphans` skips it, leaving `delegation-reminder.sh`, `verification-loop-reminder.sh`, and `code-search-routing.sh` firing alongside this one — duplicate counters and doubled reminders.

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

For the Pi adapter, the same hub accepts normalized fields and does not run plan
recovery or the Claude Git-mode heuristic:

```bash
jq -cn '{harness:"pi",session_id:"pi-smoke",tool_name:"subagent",tool_input:{}}' \
  | bash post-tool-hub.sh
```

The delegation marker should contain `|0` after this call. Plan capture is tested
through `flow-plan-capture.sh --from-pi-command`, not through this advisory hub.
