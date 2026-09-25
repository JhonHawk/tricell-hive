# Provenance — question_after_detail (S1)

Derived 2026-09-25 from `design.md`'s "Contexto verificado" and targeted `jq`
field extraction (tool name, input, assistant-text length only; no
`tool_result` payload was read). No transcript line was pasted verbatim into
this repository or into any brief; every fixture below is a hand-written,
sanitized re-expression of the observed shape in the pilot's own trace wire
format, not a copy of the source line.

## claude-fail.jsonl / grok-fail.jsonl

Source: ark Grok session
`01a0d734-a562-7462-afac-258ac95b8788`, lines 545–548 (ARK-706 batch close).

Verified via `jq` (type, tool name, assistant-text length only):
- line 545: `assistant` message, tool call `use_tool` wrapping
  `engram__mem_session_summary`, assistant text length 0.
- line 546: `tool_result` (265 chars; payload not read).
- line 547: `reasoning` (not part of the pilot's trace model).
- line 548: `assistant` message, tool call `ask_user_question`, assistant
  text length 0.

This confirms the case design.md names: an empty-text `mem_session_summary`
call followed by an `ask_user_question` call with no assistant detail. The
`claude-fail.jsonl` fixture re-expresses the same shape (a memory tool call,
its result, then a question with no detail) in Claude's wire format, since
the source session is Grok; `grok-fail.jsonl` mirrors it in Grok's own wire
format with the same `use_tool`/`ask_user_question` native tool names
confirmed against this session's tool-call name distribution.

## claude-pass.jsonl / grok-pass.jsonl

Source: sample-project Grok session
`01a0d735-762c-7821-ae46-a394c13f0dbb`, line 398.

Verified via `jq` (type, tool name, assistant-text length only): `assistant`
message, tool call `ask_user_question`, assistant text length 1293 — i.e.
substantial detail text and the question in the same message. The fixtures
re-express this shape with a short, generic Entrega/Limpieza/Recordatorio
paragraph instead of the original 1293-character report, which is not
reproduced here.

## pi-fail.jsonl / pi-pass.jsonl / pi-toolresult-fail.jsonl / claude-usertext-fail.jsonl

Not derived from a transcript: Pi has no native question tool (design.md
notes it comes from the user extension `@juicesharp/rpiv-ask-user-question`)
and no ark/sample-project Pi session with this failure was available. These four
fixtures are synthetic, built directly from the criterion's own contract
(design.md → Criterios → S1) to exercise:
- `pi-fail.jsonl`: a question with no preceding assistant detail at all.
- `pi-pass.jsonl`: a message with sufficient detail text and a sibling tool
  call (`mem_save`) whose result is observed before the question — the
  sibling does not cut the window because it shares the question's Message.
- `pi-toolresult-fail.jsonl`: a `toolResult`-role `message_end` with ≥40
  non-whitespace characters, to confirm that role, not length, gates S1.
- `claude-usertext-fail.jsonl`: a synthetic Claude `user`-role text of ≥40
  characters before the question, to confirm `user`-role text does not
  count either.

## Pi wire format

Pi 0.87.1 carries a turn's tool calls as `{"type":"toolCall","id","name","arguments"}`
blocks inside the assistant `message.content` (`ToolCall` in
`@earendil-works/pi-ai/dist/types.d.ts`), and the plan review traced the
assistant `message_end` before the matching `tool_execution_start`
(`pi-agent-core/dist/agent-loop.js:141-166`). Both come from the installed
sources; no recorded Pi trace with this failure exists, so the fixtures are
constructed. Pi's `ask_user_question` comes from the user extension
`@juicesharp/rpiv-ask-user-question@2.9.0`, not from Pi itself.
