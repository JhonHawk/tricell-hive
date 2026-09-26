# Provenance — no_poll_wait_chain

Derived 2026-09-26 from `_support/openspec/changes/archive/2026-09-26-gh-36-wait-for-completion-signal/design.md`'s
"Contexto verificado" and "Fixtures" sections. No interactive rollout line was
opened, pasted, or read verbatim: every fixture below is reconstructed from
the counts and command shapes design.md already records (from `jq` field
extraction over the Codex CLI's own session-log directory, run during that
change's own research) and from the `codex exec --json` wire schema at `rust-v0.157.1`
(`codex-rs/exec/src/exec_events.rs:122-124,210-257`,
`event_processor_with_jsonl_output.rs:236-257`). All ids, run numbers and
branch/agent names are synthetic.

## Wire-format assumption

The source sessions are interactive Codex rollouts, not `codex exec --json`
output: design.md states this format gap explicitly (N4, "Formato de
traza"). Every fixture here re-expresses the counted shapes in the
`item.completed` / `collab_tool_call` and `command_execution` form
`parseTrace`'s `codex` branch reads, per this change's own T2 parser work,
not in the interactive rollout's own event shape. `agents_states` and its
`{status, message}` entries follow the schema fields named above; which
exact agents a real `wait` handler lists is unverified (design.md, round 2
N1), so the two/one-agent shapes below are illustrative, not measured.

## codex-ci-fail.jsonl / codex-ci-bare-sleep-fail.jsonl / codex-ci-single-pass.jsonl / codex-ci-log-failed-pass.jsonl (sample-project `01a0db24`, ark `01a0dab3`)

Source: design.md's own counts — sample-project session `01a0db24` (`gh run view`
×87, `gh run list` ×13, against `gh run watch` ×4) and ark session
`01a0dab3` (`gh run view` ×145 against `gh run watch` ×1, median interval
55.7s) — no line of either transcript was read for this fixture beyond
those already-recorded counts.

- `codex-ci-fail.jsonl`: three repeats of a single `sleep 60 && gh run view
  <id> --json status` shell command, matching the "sleep-then-view" shape
  both sessions show repeated dozens of times. Neither session's real flag
  list was inspected field-by-field, so this exact flag combination is
  constructed to exercise the criterion, not copied from a transcript.
- `codex-ci-bare-sleep-fail.jsonl`: the same two-cycle shape split across
  four shell events (`sleep 60`, then `gh run view <id> --json status`,
  twice), covering the "bare sleep, next event is the query" cycle form
  design.md's criterion also recognizes.
- `codex-ci-single-pass.jsonl`: one polling cycle, a local `go test ./...`
  (standing in for the sessions' own local work between polls), then one
  more isolated cycle — the "otra herramienta corta la cadena" case.
- `codex-ci-log-failed-pass.jsonl`: one polling cycle followed immediately
  by `sleep 10 && gh run view <id> --json status --log-failed` — otherwise a qualifying query, but design.md's criterion
  excludes any command carrying `--log-failed` from counting as a status
  query (diagnostic flag), so only one real cycle is ever observed.

## codex-ci-pass.jsonl (sample-project `01a0db24`)

Not a repeated shape from the transcript counts (those sessions rarely used
`gh run watch`): built directly from the criterion's own contract to
exercise the correct-wait pass path design.md names — `gh run watch <id>
--exit-status` — the blocking-wait form both sessions used only a handful
of times against dozens of `gh run view` polls.

## codex-wait-fail.jsonl / codex-wait-pass.jsonl / codex-wait-failed-pass.jsonl / codex-wait-result-pass.jsonl (ark `01a0b087`, tricell-hive `01a0c10d`)

Source: design.md's own counts — ark session `01a0b087` (56 `wait_agent`
calls, 60000ms×45 and 10000ms×11, 66% timed out) and tricell-hive session
`01a0c10d` (166 `wait_agent` calls, 54% timed out, chains up to 7 in a row).
No transcript line was read; only the call counts and timeout parameters
design.md already records were used.

- `codex-wait-fail.jsonl`: three consecutive `collab_tool_call` `wait`
  items with one agent staying `running` throughout and no `agent_message`
  between them, standing in for tricell-hive `01a0c10d`'s up-to-7-long
  wait chains.
- `codex-wait-pass.jsonl`: the same two-wait shape with one short
  `agent_message` between them, exercising the criterion's text-break case
  — Codex's multi-agent prompt already asks the model to post a status line
  between waits (design.md, "Contexto verificado":
  `multi_agent_instructions.rs:10`).
- `codex-wait-failed-pass.jsonl`: a `wait` item with `status: failed`
  followed by a real `wait`, exercising "a failed wait does not count as a
  wait or break the chain" — not derived from either session's own failure
  shape (neither transcript's `wait_agent` failures were inspected beyond
  the aggregate timeout percentages above); constructed from the
  criterion's own contract.
- `codex-wait-result-pass.jsonl`: a `spawn_agent` item showing two agents
  `running`, then a `wait` that returns one of them `completed`, then a
  second `wait` with no text between — the "first wait brought a result"
  exception design.md's re-review round added. Constructed from the
  criterion's own contract; the two-agent shape is illustrative only, per
  the wire-format assumption above.

## Sanitization

Every fixture uses a synthetic run number (`4200`), synthetic thread ids
(`a1`, `b1`), and generic prompt/status text. No customer name, ticket
reference, path outside the fixture, or secret value appears in any file.
