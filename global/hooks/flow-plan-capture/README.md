# flow-plan-capture

PostToolUse hook on `ExitPlanMode` (Claude Code) / `exit_plan_mode` (Grok) — the organic bridge of the flow
pack ("Artifact-Attached Flow" design, 2026-07-10). When a native plan-mode plan is
approved inside a flow workspace (ledger `_support/PROJECT.md` at or above cwd), the plan
is captured to the session-capture layer (`sessions/YYYY-MM-DD-<slug>/<slug>-plan.md`,
`Status: planned`, sessions-index row) and the model is told the canonical path is the
working plan from there. `/flow-build` adopts that file (ADOPT step). opencode has no
equivalent event — it gets the same convention as instructions via the workspace
`AGENTS.md` template in `flow-core/templates/workspace-agents.md`.

**Three entry points, one capture** — same slug rules, same `Session: no` opt-out, same
index row, same idempotency:

| Flag | Harness / event | Covers |
|---|---|---|
| *(none)* | Claude Code `PostToolUse:ExitPlanMode` · Grok `PostToolUse:exit_plan_mode` | the plain approval, on either harness — told apart by the payload, see below |
| `--from-transcript` | Claude Code, called by `post-tool-hub` | the approval that clears the context — it denies the tool, so no PostToolUse fires |
| `--from-codex-prompt` | Codex `UserPromptSubmit` | **both** Codex approval buttons |

Design decisions (user-approved 2026-07-10):
- **No size threshold** — entering plan mode is the proportionality filter.
- **Opt-out rides the plan** (`Session: no` line) — the skip is explicit in the approved
  text, never a silent heuristic.
- **Idempotent by slug**: overwrite while `Status: planned`; `-2` suffix once advanced.
- **Breadcrumb log** at `$TMPDIR/claude-flow-plan-capture.log` — every write/skip auditable
  (`write(hook)` for the PostToolUse path, `write(recovery)` for the one below; a
  `[grok]` tag marks the harness when the payload was Grok's).
- Non-blocking: exit 0 always; silent outside flow workspaces.

## Grok: same entry point, different payload — and no context injection

Grok reads `~/.claude/settings.json` through its Claude compat layer, so it runs this very
hook — but two things differ, both verified against its hooks doc and a real session
payload (2026-08-21):

- **The matcher needs both names.** Grok aliases Claude tool names in matchers (`Bash` →
  `run_terminal_command`, `Task` → `spawn_subagent`), but has no alias for this one, so a
  matcher of `ExitPlanMode` alone never fires there. The block ships the regex
  `ExitPlanMode|exit_plan_mode`.
- **The payload is camelCase and `toolResult` is a plain STRING**, not Claude's structured
  `tool_response` object: a banner (`Your plan has been saved at: <path>`), then a
  `## Plan:` line and the plan text. The script prefers that file — the string field is
  subject to Grok's free-text clipping — and falls back to parsing after `## Plan:` when
  the banner wording changes. Detection is by payload shape, never by a flag: one command
  serves both harnesses, so there is nothing to pass.

**What Grok does NOT get: the `additionalContext` injection.** This hook is `PostToolUse`;
the 1.0.13 bundled guide still ignores stdout on that event (`harness/grok/README.md`
injection map). A 2026-08-21 probe on an older CLI saw injection only on
`Stop`/`SubagentStop`; PreToolUse post-call notes are a different event and do not
announce this capture. The durable half still works — the plan lands in the sessions
layer, which is the point (day-2 continuity is a repo file, not harness state, and
`/flow-build` adopts it) — but on Grok the model is not told. What reaches the model
there is the rules layer it already loads, not this hook.

## Why a recovery mode exists (`--from-transcript`)

The plan-approval dialog offers **"Yes, clear context…"** (setting
`showClearContextOnPlanAccept`; Codex has the same option). That path does NOT approve the
tool: it resolves `ExitPlanMode` as a permission **deny**, wipes the conversation, and
re-injects the plan as the first user turn of a fresh session. Consequences, verified in
the 2.1.237 binary and in real transcripts:

- **`PostToolUse` never fires** — the transcript records `toolDenialKind: "user-rejected"`
  on a plan the user approved, so the hook path above never sees it.
- **`PermissionDenied` does not fire either** — that event is gated to auto-mode classifier
  denials, not user decisions.
- **`UserPromptSubmit` does not fire** for the re-injected turn (`origin.kind:
  "auto-continuation"`), so no prompt hook can compensate.
- `SessionStart` DOES fire with `source: "clear"` and a **new session id + transcript**.
  It is not a compaction: `flow-session-context` handles that case separately.

Measured cost before the fix: sample-project, 2026-08-19 — a 171-line approved plan existed only
in the transcript; no session folder, no breadcrumb, and `flow-context` went on offering
`/flow-build` for older plans.

The signal that closes it is exact, not heuristic: the fresh session's first user row
carries a dedicated **`planContent`** field holding the plan verbatim, and that field
exists on no other path. `post-tool-hub.sh` (already spawned on every tool call) fires this
script with `--from-transcript` **once per session**, on the first tool call, and the same
capture logic runs — same slug rules, same `Session: no` opt-out, same index row. In that
mode the script prints bare text instead of a hook envelope: the caller owns the envelope.

A `/resume` of a cleared session replays the same `planContent` under a new session id, so
the once-per-session marker cannot stop it — a content guard does: an existing plan file
whose body already equals this plan ends the run, whatever its date or `Status`.

**Known limitation — retroactive capture.** The trigger is the transcript, not the moment
of approval, so resuming a session whose plan was approved long ago captures it *now*: with
`Status: planned` even if the work is done, and in a folder dated today rather than the day
of the plan. `flow-context` then offers `/flow-build` for it — a false pending. The content
guard only helps once a copy exists. This bit at deploy time (2026-08-19): five live
sessions across three client repos would each have written a stale plan on their next tool
call; they were neutralised by pre-seeding their markers
(`: > $TMPDIR/claude-flow-plan-recovery-<session_id>`), which is the manual escape hatch
whenever a recovery is not wanted.

Two things bound the damage. The recovered file **says so itself** — its `Captured:` line
adds `recovered after the fact — Status is unverified against git`, so a reader who never
runs `/flow-build` still sees the caveat (the ExitPlanMode and Codex paths capture at
approval time and carry no such line). And `/flow-build`'s reconciler treats `Status` as a
claim and git as ground truth, so a stale plan costs one offer, never re-executed work.

## Codex: one hook, both approval paths (`--from-codex-prompt`)

Codex has no `ExitPlanMode` tool and no plan-approval event. What it does have — measured
live on 0.148.0 with an isolated `CODEX_HOME` logging all 11 events — is the thing Claude
Code lacks: **`UserPromptSubmit` fires for the message the CLI injects itself**. Both
approval buttons reach the model as a prompt, so both reach this hook:

| Button | Events | Where the plan is |
|---|---|---|
| *Yes, clear context and implement* | `SessionStart` `source:"clear"` (new rollout) + `UserPromptSubmit` | **in the prompt**, behind the `A previous agent produced the plan below…` paragraph |
| *Yes, implement this plan* | `UserPromptSubmit` only (context kept, no new session) | prompt is exactly `Implement the plan.`; the plan is in the rollout as `item.type == "Plan"` |

The keep-context marker is a **structured field**, not prose: the rollout records the plan
as `event_msg` → `item_completed` → `item: {type: "Plan", text: …}`, and the hook takes the
LAST such item. Only the clear-context branch depends on English text, and only for the
preamble it strips.

Also measured, and worth knowing before trusting a Codex hook: `SessionStart` does **not**
fire when the TUI launches — it fires at the first model request, with `source:"startup"`.

Deployed by `/deploy-global` (script → `~/.claude/hooks/`, block merged into
`settings.json` from `settings-config.json`; the presence of `codex-hooks.json` also copies
the script to `~/.codex/hooks/` and merges that block into `~/.codex/hooks.json`). The
Claude recovery path needs no settings entry — it is reached through `post-tool-hub.sh`,
which resolves it as a sibling of its own directory.

**Codex trust caveat:** a new or changed hook script stays inert until it is re-accepted at
Codex's *"Hooks need review → Trust all and continue"* prompt. Editing this script means
the next Codex session asks again.

## Test cases

```bash
# Recovery: a synthetic transcript with planContent, a workspace with a ledger.
ws=$(mktemp -d); mkdir -p "$ws/_support" "$ws/x-specs/sessions"
printf 'x\n' > "$ws/_support/PROJECT.md"
python3 - "$ws/t.jsonl" <<'PY'
import json,sys
plan="Session: yes\n\n# Plan de prueba\n\ncuerpo\n"
open(sys.argv[1],"w").write(json.dumps({"type":"user","planContent":plan,
  "origin":{"kind":"auto-continuation"},"message":{"role":"user","content":"x"}})+"\n")
PY
jq -cn --arg cwd "$ws" --arg tr "$ws/t.jsonl" \
  '{session_id:"t1",tool_name:"Read",transcript_path:$tr,cwd:$cwd}' \
  | ./flow-plan-capture.sh --from-transcript
find "$ws/x-specs/sessions" -name '*-plan.md'
```

Expected: bare context text naming the captured path, and one plan file with
`Status: planned`. Re-running it (any session id) prints nothing and creates nothing.

```bash
# Grok approval: camelCase payload, toolResult a string, plan file on disk.
ws=$(mktemp -d); mkdir -p "$ws/_support"; printf 'x\n' > "$ws/_support/PROJECT.md"
printf '# Plan Grok\n\nStatus: planned\n' > "$ws/plan.md"
jq -cn --arg cwd "$ws" --arg res "Your plan has been approved.

Your plan has been saved at: $ws/plan.md

## Plan:
# Plan Grok

Status: planned" '{hookEventName:"PostToolUse",toolName:"exit_plan_mode",cwd:$cwd,toolResult:$res}' \
  | ./flow-plan-capture.sh
find "$ws/_support/sessions" -name '*-plan.md'   # -> captured; stdout ignored BY grok, emitted anyway

# No planContent -> silence, and NO breadcrumb (this is every ordinary session).
printf '{"type":"user","message":{"role":"user","content":"hola"}}\n' > "$ws/plain.jsonl"
jq -cn --arg cwd "$ws" --arg tr "$ws/plain.jsonl" \
  '{session_id:"t2",tool_name:"Read",transcript_path:$tr,cwd:$cwd}' \
  | ./flow-plan-capture.sh --from-transcript
```

```bash
# Codex, keep-context approval: the plan comes from the LAST `Plan` item.
python3 - "$ws/rollout.jsonl" <<'PY'
import json,sys
rows=[{"type":"event_msg","payload":{"type":"item_completed",
        "item":{"type":"Plan","id":"p1","text":"# Plan de prueba\n\ncuerpo\n"}}}]
open(sys.argv[1],"w").write("\n".join(json.dumps(r) for r in rows)+"\n")
PY
jq -cn --arg cwd "$ws" --arg tr "$ws/rollout.jsonl" \
  '{prompt:"Implement the plan.",cwd:$cwd,transcript_path:$tr}' \
  | ./flow-plan-capture.sh --from-codex-prompt

# Codex, clear-context approval: the plan is the prompt minus its first paragraph.
jq -cn --arg cwd "$ws" --arg p 'A previous agent produced the plan below to accomplish the user'"'"'s task.

# Otro plan

cuerpo
' '{prompt:$p,cwd:$cwd}' | ./flow-plan-capture.sh --from-codex-prompt
```

Expected: `hookEventName: "UserPromptSubmit"` and one plan file each. Any other prompt
(every ordinary user turn) exits 0 with no output and no breadcrumb.
