# Provenance — no_bare_url

Derived 2026-09-26 from `_support/openspec/changes/report-readability/design.md`'s
"URLs sueltas" finding and targeted `jq` field extraction on the source
transcript: only `.message.content[].text` (assistant text blocks) was read,
via a boolean `test("https?://")` and a second boolean `test("\\]\\(https?:
//")` check on that same field — never the field's own content printed
alongside anything else, never a `tool_result` payload, and never a command
that prints a whole transcript line.

## claude-fail.jsonl

Source: Claude Code session `69e9dd88-2845-50ea-82a6-22ea09451a54` in this
project's local transcript history, 2026-09-23, 2437 lines — the session
design.md's own "URLs sueltas" finding names (chosen there specifically to
avoid carrying client URLs, per its own text).

Verified via `jq`, scanning every `type=="assistant"` text block for
`test("https?://")` and flagging which also matched `test("\\]\\(https?://")`
(a Markdown link): six blocks contained a URL, five of them (lines 224, 238,
255, 1035, 2376 — 2376 the only one with a Markdown link) too long or too
compound to use as a minimal fixture; line 548 (174 characters) is short and
self-contained: a status update naming two local development URLs neither
wrapped in a Markdown link nor in any other excluded form.

`claude-fail.jsonl` re-expresses that shape: the same sentence structure
("… sigue parada a la espera de que valides en `<url>` y `<url>`."), with
both real `http://localhost:____` URLs replaced by the fixed placeholder
`https://example.com/one` / `https://example.com/two`, per this change's
sanitization rule (no real URL, not even a harmless local one, is carried
into a fixture verbatim).

## claude-pass.jsonl (constructed)

Not derived from a single transcript line: no session inspected here has one
assistant text block exercising every excluded form design.md lists in one
place. Built directly from the criterion's own contract
(`_support/openspec/changes/report-readability/design.md`, "no_bare_url") to
cover all five: a Markdown link (`[label](url)`), an angle-bracket autolink
(`<url>`), a single-backtick inline-code URL, a URL inside a ` ``` ` fenced
block, and a URL inside a `~~~` fenced block — five distinct
`https://example.com/…` placeholders, one per form, so a broken exclusion for
any one form still leaves the others correctly excluded and is easy to tell
apart from a wholesale regression.

## Wire-format note

Both fixtures use Claude Code's ordinary streamed assistant-text shape
(`{"type":"assistant","message":{"id":…,"content":[{"type":"text","text":…}]}}`),
identical in structure to the pre-existing `question_after_detail/claude-*.jsonl`
fixtures (untouched by this change).
