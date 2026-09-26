# Provenance — ticket_ids_not_packed_in_prose

Derived 2026-09-26 from ark Grok session `01a0dd4c-33d8-7920-85ef-9df30c78f78d`,
line 61 of that session's `chat_history.jsonl` (assistant text, 8463 chars),
2026-09-26. The session was not opened directly here: the shape below was
handed down already sanitized by the main thread (the finding's own
directive was "Do NOT open the session transcript; build the fixture from
this sanitized shape"), so no `jq` extraction was run against the original
transcript for this criterion.

## What was kept

- The paragraph structure: an opening status line, then several paragraphs
  each opening with a bold module-style label (`**Templates.**`,
  `**CI and tests.**`), then a short recommendation paragraph.
- The chaining pattern inside each failing paragraph: 3–8 Markdown-linked
  ticket IDs run together in one prose paragraph, several with a parenthetical
  gloss after the link (`(validation)`, `(sync button)`).
- The link-per-ticket convention: each ID is its own `[ID](url)` Markdown
  link, with the tracker path repeating the ID (`/issue/ABC-101`).
- The overall counts: `grok-fail.jsonl`'s Templates paragraph has 4 distinct
  IDs, its CI paragraph has 3 — both at or above
  `ticketProsePackThreshold` (3).

## What was replaced

- Every real ticket ID, replaced by the synthetic `ABC-1xx`/`ABC-2xx` series
  (design.md's own convention for a sanitized, non-persisted example ID).
- Every ticket title/description, replaced by short generic phrases ("rate
  limit on submit", "sync button") carrying no client or product detail.
- Every tracker URL, replaced by `https://tracker.example.com/issue/<ID>`
  (no real host, no real project key).
- The real backlog count, replaced by an arbitrary placeholder number (46).

## Wire-format assumption

Grok's streamed assistant-text shape, identical in structure to the
pre-existing `question_after_detail/grok-*.jsonl` and `no_bare_url` fixtures'
Claude shape (`parseTrace` treats `claude` and `grok` identically for a plain
`type=="assistant"` message with a `text` content block; there is no
`tool_use`/`use_tool` wrapping involved here since these fixtures carry only
assistant text, no tool calls):

```json
{"type":"assistant","message":{"id":"m1","content":[{"type":"text","text":"…"}]}}
```

Source file for this assumption: `tests/fixtures/regression/question_after_detail/grok-fail.jsonl`
(same envelope, verified working with `parseTrace("grok", …)` in
`TestRegressionFixtures`).

## Fixtures

### `grok-fail.jsonl`

The sanitized shape from the finding almost verbatim (four paragraphs
separated by blank lines): the Templates paragraph cites four IDs
(ABC-101–104) in prose, glossed by parentheses or plain apposition but never
put on their own list line; the CI paragraph cites three (ABC-201–203) the
same way. The criterion returns `fail` on the first paragraph over threshold
in trace order — the Templates paragraph — with evidence
`stdout.jsonl:1 text   ABC-101,ABC-102,ABC-103,ABC-104` (the two blank fields
after `text` are the event's empty Path and Tool, always blank for a plain
assistant text event; verified by direct print of `ticketIDsNotPackedInProse`'s
output against this fixture).

### `grok-pass.jsonl` (constructed)

Same ticket set and module labels, restructured per the fixed guidance: each
ticket or ticket group is its own list line under a bold module heading
(`**Templates**` / `**CI and tests**`), so `proseLinesOf` drops every list
line and its wrapped ID pairs entirely — none of those lines is scanned as
prose. Two prose (non-list) paragraphs remain: one citing two IDs
(`ABC-301`/`ABC-302`, under threshold) and the closing recommendation citing
one (`ABC-101`). Constructed directly from the criterion's own contract, not
derived from a second transcript line — no single real session paragraph in
scope exercises every boundary (list exclusion, a two-ID prose paragraph, a
one-ID prose paragraph) at once.

### `grok-list-continuation-pass.jsonl` (constructed)

Exercises the continuation-line rule specifically: a list item's ID-bearing
line wraps onto an indented second line before the next `- ` item. Without
`proseLinesOf` treating that indented line as part of the same excluded list
item, the block's surviving prose would still be empty either way here (the
continuation line carries no ID itself) — its purpose is instead to prove the
revert-proof case below: with list exclusion disabled outright, the whole
block (heading + both list lines + continuation) is read as one prose
paragraph with four distinct IDs and fails.

### `grok-links-dedup-fail.jsonl` (constructed)

A prose (non-list) paragraph where three Markdown links each repeat their own
ID in both the visible label and the URL path
(`[ABC-401](.../issue/ABC-401)`). Not derived from a transcript line;
constructed to prove `maskForTicketProse`'s URL-masking does not corrupt or
drop the link's label text — the distinct-ID count from the visible labels
alone (ABC-401, ABC-402, ABC-403) must still reach the threshold and fail,
exactly as it would if the URLs were absent.

## Revert-proof (recorded 2026-09-26)

Two branches were each disabled in turn in a local, uncommitted edit to
`tests/pilot/regression.go`, `go test ./tests/pilot/... -run
TestRegressionFixtures/ticket_ids_not_packed_in_prose -v` was run, and the
edit was then reverted (confirmed clean via `git diff --stat` and a final
green `go test -race -count=1 ./tests/pilot/...`).

1. **List-exclusion branch disabled** (`proseLinesOf` replaced with `return
   block`, bypassing list/table exclusion entirely): `grok-pass.jsonl` and
   `grok-list-continuation-pass.jsonl` both flipped from `pass` to `fail`
   (`got fail want pass`), because their list-line ID pairs were then read as
   raw prose. `grok-fail.jsonl` and `grok-links-dedup-fail.jsonl`, which carry
   no list lines, were unaffected.
2. **Distinct-ID threshold branch disabled** (the `len(seen) >=
   ticketProsePackThreshold` check short-circuited to `false`):
   `grok-fail.jsonl` and `grok-links-dedup-fail.jsonl` both flipped from
   `fail` to `pass` (`got pass want fail`), since no paragraph could ever
   reach "packed" status. `grok-pass.jsonl` and
   `grok-list-continuation-pass.jsonl`, which have no paragraph at or above
   threshold either way, were unaffected.

Each fixture pair proves the branch it targets: disabling either one flips
exactly the fixtures that depend on it, and restoring the code returns all
four to their declared status.

## Follow-up: heading and bold-label masking (2026-09-26)

Coordinator-reported false positive, verified by running the criterion over
four independently good backlog-grouping answers (synthetic backlog, grouped
by work session, every ticket already on its own list line): all four failed
because a group's own label line — a heading or a bold span naming that
group's members — was scanned as ordinary prose. Three label shapes were
observed, described (not derived from a specific transcript line — these are
a generalization across the four runs, not one session):

1. A heading naming the group: `## G2: API Keys (ABC-220, ABC-221, ABC-222)`,
   followed directly by a bold note and the group's own list lines.
2. A bold label with trailing text after the closing marker:
   `**S2: API Keys (ABC-220, ABC-221, ABC-222)** — pequeña/mediana`, followed
   directly by the group's own list lines.
3. A bare bold line naming only the IDs, under a heading:
   `**ABC-220 · ABC-221 · ABC-222**`, followed directly by the group's own
   list lines.

### Change

`proseLinesOf` now drops a heading line (`isHeadingLine`, trimmed start `#`)
outright, the same way it already dropped a list-marker or table line.
`maskForTicketProse` now also blanks the entire content of a bold span
(`maskBoldSpans`, `**...**` or `__...__`) before scanning, so an ID cited only
inside a bold label is never counted, wherever that label sits in a line.

### New fixtures (all constructed; no transcript line — see above)

- `grok-heading-label-pass.jsonl`: shape 1. Isolates heading exclusion: its
  bold line (`**Grouped for one session**`) carries no ID, so only the
  heading-drop mechanism keeps this fixture from failing.
- `grok-bold-label-pass.jsonl`: shape 2. Isolates bold-span masking: no
  heading is present, so only the bold-masking mechanism keeps this fixture
  from failing.
- `grok-bold-ids-only-pass.jsonl`: shape 3. Exercises bold-span masking on a
  line that is *entirely* a bold span (blanks to nothing), under a heading
  that itself carries no ID (so heading exclusion is present but inert here).
- `grok-bold-label-then-bare-ids-fail.jsonl` (new fail fixture, per the
  contract's own requirement that this stay a failure): a bold label
  (`**Follow-up needed:**`) is immediately followed, on the same line but
  outside the bold markers, by three IDs in plain prose
  (`ABC-501, ABC-502 and ABC-503`). Bold masking only blanks "Follow-up
  needed:"; the three IDs stay visible and still fail — proving label masking
  does not swallow a citation merely adjacent to a label.

Each pass fixture also keeps a short, separate "Recommendation: start with
ABC-220." prose line so the overall result is an explicit `pass` (an ID was
found and judged, and stayed under threshold), not merely `not_observed`.

The three pre-existing fail fixtures (`grok-fail.jsonl`,
`grok-links-dedup-fail.jsonl`) were rerun unchanged and still fail: their bold
spans (`**Templates.**`, `**CI and tests.**`) carry no IDs of their own, so
masking them removes nothing relevant, and their packed IDs sit in plain
prose outside any bold span or heading.

### Revert-proof (recorded 2026-09-26)

Each new masking mechanism was disabled in turn in a local, uncommitted edit,
`go test ./tests/pilot/... -run
TestRegressionFixtures/ticket_ids_not_packed_in_prose -v` was run, then the
edit was reverted (confirmed via `grep -n REVERT-PROOF-DISABLED
tests/pilot/regression.go` printing nothing, and a final green `go test
-race -count=1 ./tests/pilot/...`).

1. **Heading exclusion disabled** (`isHeadingLine(line) ||` short-circuited
   to `false &&` in `proseLinesOf`): only `grok-heading-label-pass.jsonl`
   flipped from `pass` to `fail` (`got fail want pass`) — its heading's own
   three IDs were then read as prose. Every other fixture, including the two
   other new pass fixtures (protected by bold masking, not heading
   exclusion), was unaffected.
2. **Bold-span masking disabled** (the `maskBoldSpans` call in
   `maskForTicketProse` commented out): `grok-bold-label-pass.jsonl` and
   `grok-bold-ids-only-pass.jsonl` both flipped from `pass` to `fail` (`got
   fail want pass`) — their bold labels' IDs were then read as prose.
   `grok-heading-label-pass.jsonl` (protected by heading exclusion, not bold
   masking, since its own bold line carries no ID) and
   `grok-bold-label-then-bare-ids-fail.jsonl` (already failing on IDs outside
   any bold span) were both unaffected.

Each fixture pair proves the mechanism it targets: disabling either one flips
exactly the fixtures that depend on it, and restoring the code returns every
fixture in this folder to its declared status.
