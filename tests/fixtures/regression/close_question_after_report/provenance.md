# Provenance — close_question_after_report

Derived 2026-09-25 from `design.md`'s "Contexto verificado" (X1, G3) and
targeted `jq` field extraction (event `type`/`payload.type`/`payload.role`,
tool-call `name`, `content` length, and boolean `test(...)` checks on the
transcript text — the text itself was never printed). No transcript line
was pasted verbatim into this repository, its provenance, or any brief;
every fixture below is a hand-written, sanitized re-expression of the
observed shape in `parseTrace`'s own streamed wire format (`codex exec
--json` items for Codex, `streaming-messages-json` for Grok), not a copy of
the source line.

## codex-fail.jsonl (X1)

Source: Codex sample-project session `01a0dab1-f01c-78a2-9e36-aa58218c79b2`
(gpt-6-sol medium), rollout dated 2026-09-25 (82 lines total).

Verified via `jq` (`type`/`payload.type`/`payload.role` only):
- line 79: `response_item` / `message`, role `assistant` — the turn's final
  report (ARK-687).
- line 80: `token_usage_record`.
- line 81: `event_msg` / `token_count`.
- line 82: `event_msg` / `task_complete` — the turn ends immediately after
  the report, with no question event of any kind in between.

`codex-fail.jsonl` re-expresses this shape in the `codex exec --json`
streamed format `parseTrace` reads: one `item.completed` `agent_message`
(the report, paraphrased — ARK-687 and its real wording are not
reproduced) immediately followed by `turn.completed`, with no `question`
event and no line of the report text ending in "?".

## codex-pass.jsonl (constructed)

Not derived from a transcript: no available Codex session closes with a
text-only interrogative report line (Codex `exec` 0.157.0 has no
`request_user_input`; see `design.md`'s "Criterios deterministas"). Built
directly from the criterion's own contract to exercise the text-fallback
pass path: one `agent_message` whose report ends in a line ending "?",
followed by `turn.completed`.

## grok-fail.jsonl (G3)

Source: Grok globex session `01a0daf9-ddec-7421-ad4e-7b7e439dd873`
(grok-4.7-build-fast high), `chat_history.jsonl` (557 lines total).

Verified via `jq` (`type`, `tool_calls[].name`, `content` length, and a
boolean `test("\\?\\s*$")` on the rtrimmed report text — the text itself
was never printed):
- line 138: `assistant`, four tool calls, the fourth named
  `ask_user_question` — the session's only native question in the whole
  file, about the deployment's scope, not a close question.
- lines 396, 399, 405: `assistant` messages with a `run_terminal_command`
  tool call each; line 402 calls `get_command_or_subagent_output`; none of
  these commands match a merge or branch-deletion shape (checked with
  `test("branch -d|push.*--delete|push.*:.*|gh pr merge")`, all `false`).
- line 408: `assistant`, `content` length 1,122, zero tool calls, and
  `test("\\?\\s*$")` on the content is `false` — the production
  deployment's final report, closing with no question and no trailing "?".
- A file-wide scan (`awk 'NR>138' … | jq '.tool_calls // [] | map(.name) |
  any(. == "ask_user_question")'`) found no further `ask_user_question`
  call anywhere after line 138 (105 messages checked, all `false`).

`grok-fail.jsonl` re-expresses this shape: an early `ask_user_question`
about scope (paraphrased, not the original wording), its accepted result,
then a closing report with no trailing question mark and no subsequent
question event — matching G3 exactly.

Since D17-A, a fixed intervening `run_terminal_command` call (paraphrased,
not the real deployment command) was added between the `ask_user_question`
tool result and the closing report, so the fixture actually exercises the
"work continued after an early question" shape the finding describes: the
D17-A unanswered-question shortcut (see below) must not fire just because a
native question occurred somewhere earlier in the trace with only assistant
text after the *last* text event; it must see a further tool call before
that final report and fall back to the ordinary check, which still fails
because the report itself has no question of its own. Before this change,
the simplified fixture had no event between the question's tool result and
the closing report, so it could not tell the two shapes apart.

## grok-pass.jsonl (constructed) and grok-text-fallback-pass.jsonl (constructed)

Not derived from a transcript: G3 is the only close-question finding on
file for Grok, and it is a failure. Built directly from the criterion's own
contract (`design.md` → "Criterios deterministas" →
`close_question_after_report`) to exercise its two pass paths:
`grok-pass.jsonl` follows the report text with a native
`ask_user_question` tool call in the same message (evidence kind
`question`); `grok-text-fallback-pass.jsonl` has no question tool call at
all, with the report's own last line ending in "?" (evidence kind `text`,
the fallback path a human should be able to spot and re-check by hand per
`design.md`).

## D17-A: two false negatives found in the 2026-09-26 pilot (T4)

The first `close-sequence` pilot runs (`_support/workspace/2026-09-26-work-close-sequence/runs/`)
found the criterion missed two real passes. Both fixtures below are
paraphrased re-expressions of the observed shape (never the real report
text) in `parseTrace`'s `streaming-messages-json` wire format, located with
`jq` extracting only `type`/`.message.content[].type`/`.name`/`.tool_use_id`
and text *lengths* — the report text itself was never printed. Both
sessions used Grok's direct top-level tool-call form (`ask_user_question`,
`run_terminal_command`), matching the wire-format note below, not the
`use_tool` envelope.

### grok-tool-then-text-fail.jsonl (grok-a-1)

Source: pilot run `grok-a-1` (`close-sequence`, arm A), 29-line
`stdout.jsonl`. Verified via `jq` (`type`, content `type`/`name`, and text
`length`, never the text itself):
- line 26: `assistant`, two tool calls in the same message — `todo_write`
  then `ask_user_question` (the only native question in the file, and the
  run's last tool call).
- line 27: `user`, two `tool_result`s for those same two calls, in order.
- line 28: `assistant`, one `text` block, length 1393, zero tool calls —
  the run's closing report.
- line 29: `result`, `is_error` false, whose own `result` field is also
  1393 characters (Claude/Grok's terminal event duplicates the last
  assistant text; `parseTrace` maps it to `Kind=="final"`, already excluded
  from consideration elsewhere in this criterion).

Before D17-A, `closeQuestionAfterReport` only ever looked *forward* from
the trace's last assistant text for a following `question` event. Here the
question came *before* that last text (the model kept writing after an
unanswered `ask_user_question`), so the old code read this as a report with
no question and failed it — a false negative for a real pass. The fixture
reproduces exactly that shape, including the sibling `todo_write` call
resolved in the same `tool_result` batch, to exercise the fix's
message-based sibling-result matching (not only exact call-ID matching).

#### Reclassification (2026-09-26, finding 7, `/code-review`)

D17-A's shortcut accepted any unanswered native question followed only by
its own result and more text, with no requirement that the question itself
be preceded by any real detail. A code-review finding pointed out this let
"the whole report" pass merely by landing *after* a mid-flow question — and
a field-only `jq` re-check of the real `grok-a-1` trace's earlier lines
(1–25: mostly `read_file`/`run_terminal_command` tool-call turns, with only
short, non-adjacent assistant `text` blocks of 200/179/117/134 characters,
none directly preceding line 26) confirmed this fixture's own shape is an
instance of exactly that bug: the `todo_write`/`ask_user_question` pair at
line 26 has zero non-whitespace runes of assistant text directly preceding
it (per `detailRunesBeforeQuestion`'s window, the same one S1's
`question_after_detail` uses), and the run's real, substantive report
(line 28) is written only *after* the question, never itself asking a close
question. `closeQuestionAfterReport` now requires at least
`minQuestionDetail` (40) runes of detail before the question for the
shortcut to apply; this fixture correctly reclassifies from a false "pass"
to `fail` under that corrected criterion, and is renamed accordingly (no
change to its bytes). This affects the T4 pilot table's Grok arm A count for
this run when re-assessed with `--assess`; see `criterion-assessment-v3.json`
and the change's tasks.md for the corrected reading.

### grok-text-midline-pass.jsonl (grok-b-2)

Source: pilot run `grok-b-2` (`close-sequence`, arm B), 33-line
`stdout.jsonl`. Verified via `jq`: the run's last `assistant` message
(line 32) has one `text` block, length 1437, and no tool call; line 33 is
`result`. The real report's last paragraph asked a question mid-sentence
("Es lo que recomiendo, ¿…? …") and then kept going with more sentences
after the "?", so it never ended in "?". Before D17-A, the fallback only
ever checked whether the *last line* ended in "?", so this read as a fail —
a false negative for a real pass. The fixture paraphrases that shape: a
single paragraph whose "?" sits mid-sentence, with more text after it.

### grok-offer-no-question-fail.jsonl (grok-b-1)

Source: pilot run `grok-b-1` (`close-sequence`, arm B), 31-line
`stdout.jsonl`. Verified via `jq`: the run's last `assistant` message (line
30) has one `text` block, length 1463, and no tool call; line 31 is
`result`. The real report offered to continue without ever asking a
question (no "?" anywhere in it) — a true fail, both before and after
D17-A. Registered here as a fixed regression case for the "offer without a
question mark" shape, distinct from `grok-fail.jsonl`'s "question occurred
earlier, work continued" shape.

## Wire-format note

Grok's CLI in the globex session calls `run_terminal_command` and
`ask_user_question` directly as top-level tool names (not wrapped in a
`use_tool` envelope, unlike the memory-tool examples in
`question_after_detail/provenance.md`). Both forms are accepted by
`parseTrace`'s `grok` branch (`grokToolCall` only unwraps `use_tool`
specifically), so the fixtures use the same direct form observed in this
session.
