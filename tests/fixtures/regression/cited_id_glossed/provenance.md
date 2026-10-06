# Provenance — cited_id_glossed

Derived 2026-09-26 from `_support/openspec/changes/report-readability/design.md`'s
"Contexto verificado" table and targeted `jq` field extraction on the source
Claude Code transcripts: only `.message.content[].text` (assistant text
blocks) and `.message.content[].input.questions[]` (`AskUserQuestion`'s own
question/header/options fields) were ever read, via `jq -c`/`jq -r` filters
naming those exact fields — never a `tool_result` payload, and never a
command that prints a whole transcript line. No transcript line was pasted
verbatim into this repository, its provenance, or any brief; every fixture
below is a hand-written, sanitized paraphrase of the observed shape
(business project names, exact tech choices, exact figures, and exact
wording are removed or replaced), not a copy of the source lines.

## claude-fail.jsonl

Source: Claude Code (`claude-opus-5-5`) session `e8c62372-8843-5c54-b784-b9563e24a9d8`
in this project's local transcript history, 2026-09-25, 3444 lines.

Verified via `jq` reading only `.type`, and, for `type=="assistant"`, the
`AskUserQuestion` tool's own `input.questions[].header`/`.question`/
`.options[].label` fields (never `tool_result`):
- Line 2387: an `AskUserQuestion` call, header "D3 Stack", defines option
  `D3-A` (among `D3-B`/`D3-C`) about a stack-adoption pilot between two
  internal projects — this is the option's own label, a definition per
  design.md.
- Line 2551: a *later*, separate `AskUserQuestion` call (header "D4") whose
  own `question` field asks "¿Qué hacemos con D3-A?" — `D3-A` appears
  immediately before "?", not followed by any of design.md's gloss marks
  (`(`, em dash, en dash, `:`) — a real, observed unglossed cross-message
  citation.

`claude-fail.jsonl` re-expresses this shape in Claude's real streamed
format (`{"type":"assistant","message":{"id":…,"content":[{"type":"tool_use",
"name":"AskUserQuestion","input":{"questions":[…]}}]}}`), paraphrasing the
topic (two unnamed "projects" instead of the real project names, a generic
"stack divergence" instead of the real technology names) and dropping the
unrelated `D7-A`/`D8-A` mentions the real line 2551 also carried, since the
criterion only needs one clean citation to reach `fail` deterministically.

### A definition/citation distinction found while building this fixture

An earlier draft's D4-B option label read "D4-B: mantener D3-A" (mirroring
D4-A's real "D4-A: cerrar D3-A"). Since design.md's definition rule initially
read (in this change's own draft implementation) as "any ID inside an option
label is a definition," `D3-A` appearing inside `D4-A`'s/`D4-B`'s own labels
was wrongly read as *redefining* `D3-A` within message `question@2`, which
then suppressed the real citation in that same message's `question` field
(self-message non-citation) — flipping the expected `fail` to `not_observed`.
Re-reading design.md's definition rule against its own "opens a line" phrasing
for assistant text (a *position*, not "anywhere in the field") resolved this:
an option label defines an ID only when the ID opens that label (`D3-A:
piloto…`), exactly as a assistant-text line only defines an ID that opens it.
An ID cited elsewhere inside a *different* option's label (`D4-A: cerrar
D3-A`) is a citation, glossed via design.md's rule 2 ("an option label whose
description is non-empty") — not a definition. `tests/pilot/regression.go`
implements this (`scanIDOccurrences`'s label branch: `isDef := o.Start == 0`).
This fixture keeps `D4-A: cerrar D3-A (Recomendado)`/`D4-B: mantener D3-A` as
written, since they now correctly exercise that exact distinction end to end
(both are glossed citations via rule 2, so the fixture's single failure is
still, and only, the bare `D3-A` in the `question` field).

## claude-pass.jsonl

Source: Claude Code (`claude-sonnet-5`) session `04192cfb-d9bd-56b1-9825-dd8102dce715`,
2026-09-25.

Verified via `jq` on the same two field families:
- Line 113: an `AskUserQuestion` call with two sub-questions in one call
  (`questions: [...]`), headers "D2 Esfuerzo" and "D3 Portab.", defining
  options `D2-A`/`D2-B` and `D3-A`/`D3-B` about per-role effort
  configuration — a Hive-internal configuration topic, not client business,
  so it is kept close to the original substance (still paraphrased, not
  copied).
- Line 158: a later, separate `AskUserQuestion` call whose `question` field
  reads "¿Cómo seguimos con D2-A (esfuerzo en todos los roles) y D3-A (campo
  effort portable)?" — both citations are immediately followed by " (", per
  design.md's rule 1.

This is the exact "D2-A (esfuerzo en todos los roles)" shape design.md's own
worked example cites. Re-expressed with the real nested `questions` input
shape, paraphrasing field/topic names slightly ("campo de esfuerzo
portable" instead of "campo effort portable").

## claude-question-after-text-fail.jsonl (constructed)

Not derived from a single transcript line, but built to exercise a specific
implementation branch design.md requires: "Un evento question es siempre su
propio mensaje" (a question event is always its own message), even when a
host emits an assistant text block and the following `AskUserQuestion` call
as two content items of the *same* wire-format message (one `message.id`).
This fixture's text block ("- **D5**: mover la validación…") and its
`AskUserQuestion` call share `message.id` "m1" on purpose — real Claude Code
turns commonly interleave commentary text and a tool call this way — so that
reverting the synthetic per-event message key for question events
(`idQuestionMessageKey`) and instead trusting the shared `e.Message` would
wrongly read this as "the same message defines D5, so citing it bare is not
a citation," flipping the expected `fail` to `not_observed`. See T4's
revert-check evidence below for the confirmed effect.

## claude-range-fail.jsonl

Source: same session as `claude-fail.jsonl` (876d776d), a distinct occurrence
covering the range-token and option-description surfaces (both untouched by
`claude-fail.jsonl`, which uses only single IDs and a question's own
`question` field):
- Lines 563 and 627: assistant text table rows (`| **S1** | … |`, `| **S2**
  | … |`, `| **S3** | … |`) define `S1`–`S3` at each row's own start — a
  table-cell definition per design.md.
- Line 1558 (`jq` on `.input.questions[].options[].description` only): a
  *later*, separate `AskUserQuestion` call's option **description** field
  reads "…S1–S6 y la release e4b93bdb7fa4…" — the range citation is followed
  by " y", not one of design.md's gloss marks — unglossed.

`claude-range-fail.jsonl` re-expresses this with a shortened `S1`–`S3` range
(dropping `S4`–`S6`, unnecessary for one clean failing citation) and a
paraphrased option description, in the option-**description** field
specifically (not the `question` field `claude-fail.jsonl` already covers),
for surface variety.

## claude-range-pass.jsonl (constructed)

Not derived from a transcript: no session inspected here shows a glossed
cross-message *range* citation (every real range citation found was bare;
see `claude-range-fail.jsonl` and design.md's own "Fallos observados" table,
which lists no passing range example). Built directly from design.md's own
contract to exercise the range pass path, reusing the same `S1`–`S3`
definitions as `claude-range-fail.jsonl`: a later assistant **text** message
(not a question) reads "El lote S1–S3 (cierre, reparto y git add) queda
desplegado. ¿Seguimos con el resto?" — the range is immediately followed by
" (", glossing it. Using assistant text rather than a question for this
citation also gives the text surface a fixture where it is the *citing*
surface, not only the defining one (see T4's revert-check table).

## Per-branch revert check (T4 verification, 2026-09-26)

Each branch was disabled in isolation directly in `tests/pilot/regression.go`
(a one-line, clearly marked `REVERT-CHECK` edit each time), `go build`
confirmed the tree still compiled, `go test -run TestRegressionFixtures` (and,
for the range check, `TestCitedIDGlossedRangeRequiresMatchingLetters`) was run,
and the edit was restored before moving to the next branch. `diff` against a
pre-check copy of the file confirmed the restore was byte-for-byte exact.

| Branch disabled | How | Fixtures/test that flipped |
| --- | --- | --- |
| Text surface (assistant `Kind=="text"` never scanned) | `case false && e.Kind == "text" && …` | `claude-question-after-text-fail.jsonl` (fail→not_observed), `claude-range-fail.jsonl` (fail→not_observed), `claude-range-pass.jsonl` (pass→not_observed) |
| Question surface (`Kind=="question"` never scanned) | `case false && e.Kind == "question":` | `claude-fail.jsonl`, `claude-pass.jsonl`, `claude-question-after-text-fail.jsonl`, `claude-range-fail.jsonl` all flipped to not_observed (`claude-range-pass.jsonl`, whose citation lives in text, was unaffected) |
| "A question is always its own message" (used `e.Message` instead of the synthetic `idQuestionMessageKey(i)`) | `key := e.Message` | Only `claude-question-after-text-fail.jsonl` flipped (fail→not_observed) — this is the one fixture built specifically to isolate this rule, since its text definition and its question share one real `message.id` |
| Range letter-equality check (`if letterA != letterB { continue }` disabled) | `if false && letterA != letterB` | The five `.jsonl` fixtures were unaffected (a real range's own connecting dash already satisfies the general gloss-mark check for its first endpoint regardless of combining, so combining vs. not does not change their pass/fail here); `TestCitedIDGlossedRangeRequiresMatchingLetters` flipped from pass to fail, which is exactly the case this unit test was written to isolate (design.md's D1–S6/mismatched-letters concern) |

## Wire-format note

All five fixtures use Claude Code's real nested `AskUserQuestion` input,
`{"questions":[{"header","question","options":[{"label","description"}]}]}`,
confirmed against session `876d776d`'s own `input` keys (`jq
'.message.content[]?|select(.type=="tool_use" and .name=="AskUserQuestion")|
.input|keys'` → `["questions"]`), not the flat `{"question":…}` shape the
pre-existing `question_after_detail/claude-*.jsonl` fixtures use (those are
untouched by this change).
