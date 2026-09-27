# Provenance — task_marked_after_verdict

Derived 2026-09-26 from OpenCode session `ses_f20143eacffe8d44SvMbTnjkr7`
(the CLI's local `opencode.db`, table `session_message`, read with
`sqlite3 -readonly` and `json_extract`/`json_each`, field by field, per
`README.md`'s "Locate without printing"). No transcript line, prompt text or
tool-result body was ever printed; only field names, tool names, role
strings, boolean flags, path suffixes, and line-based checkbox counts were
extracted, matching `_support/openspec/changes/per-task-verification/design.md`'s
own "Contexto verificado" account of this same session (H4, "Marcas sin
veredicto").

## Role substitution

The source session launched `sdd-verify` and `review-ux` in the same batch
(seq 1848, two `subagent` tool calls, `state.input.agent` = `"review-ux"` and
`"sdd-verify"`, both `state.input.background` = `true`, both
`state.status="completed"` already on that same line — the launch's own
acknowledgment, not the child's real completion). `review-task` did not exist
in that session (per-task-verification's own role, added by this change);
every fixture below substitutes it for whichever of those two names the
underlying shape came from. The three `tasks.md` marks (seq 1861, 1876,
1889, each ~9-13 s apart from the launch) and the real completion signal
(seq 1992, `type="synthetic"`, `metadata.source="subagent"`,
`metadata.agent="sdd-verify"`, `metadata.state="completed"`, ~33 min after
the launch) are the shape `opencode-fail.jsonl` and `opencode-pass.jsonl`
reconstruct, substituting `review-task` for `sdd-verify`.

Extracted, never printed:
- seq 1848: two `session_message.data.content[]` blocks of `type="tool"`,
  `name="subagent"`, `state.status="completed"`, `state.input` keys
  `agent`/`background`/`description`/`prompt`, values `agent="review-ux"`/
  `"sdd-verify"`, `background=1` (true) for both.
- seq 1861/1876/1889: each has exactly one `content[]` block of
  `type="tool"`, `name="edit"`, `state.input` keys `path`/`oldString`/
  `newString`; `path` ends `.../<change-id>/tasks.md`
  in all three. Counted (never printed) via
  `(length(x) - length(replace(x, marker, ''))) / length(marker)` for each
  marker in `{"- [x]", "- [ ]", "- [!]", "- [?]", "- [/]"}`: every one of the
  three edits has exactly one `"- [ ]"` in `oldString` going to zero, and
  zero `"- [x]"` in `oldString` going to exactly one in `newString` — a
  single-task mark each, none a batch or a reject.
- seq 1992: top-level keys `metadata`/`time`/`text`/`description`;
  `metadata` keys `source`/`childID`/`agent`/`state`; values
  `source="subagent"`, `agent="sdd-verify"`, `state="completed"`.
- `time_created` deltas from seq 1848: +22.7s (1861), +32.4s (1876), +39.3s
  (1889), +1994.4s / ~33.2 min (1992) — consistent with design.md's
  independently-recorded "9, 16 y 30 s" (measured from the launch's own ack,
  not from seq 1848 itself) and "unos 33 min después".

## Wire-format assumption

The source is the OpenCode V2 database's own internal message shape
(`session_message.data`, `type="synthetic"`, a top-level `metadata` object,
no `part` wrapper), not `opencode run --format json` output: design.md
states this gap explicitly ("Supuesto de formato, más fuerte de lo que
parece"). Every OpenCode fixture below re-expresses the launch and its
foreground completion in the streamed `tool_use`/`part.tool`/`part.state`
form `parseTrace`'s `opencode` branch already reads (matching the DB's own
`tool`/`state.input`/`state.status` field names, renamed to the stream's
`part.tool`/`part.state.input`/`part.state.status`), and invents a
`type="synthetic"` top-level stream event, carrying the DB's own
`metadata.source`/`metadata.agent`/`metadata.state` unchanged, for the
background completion signal — this exact stream shape is not observed from
a live `opencode run --format json` invocation, only inferred from the DB
row above, per design.md's own admission.

Claude's `origin.kind="task-notification"` shape was not derived from this
OpenCode session at all — Claude was never the host of the source session.
Per design.md, it is "solo se observó en transcripciones interactivas de
Claude Code 2.1.283, no en `claude -p --output-format stream-json`", so
`claude-background-ack-fail.jsonl` and `claude-background-pass.jsonl` mirror
the OpenCode shape's timing (ack immediately after launch, real signal much
later) using Claude's own launch tool names (`Task`, `subagent_type`,
`run_in_background`) and the declared `origin.kind`/`subagent_type` shape,
not a session capture.

## Fixtures

- `opencode-fail.jsonl`: launch (`background=true`, ack `completed`), then
  the mark, then the real `synthetic` end arriving after — reproduces H4's
  bug shape (mark before the real end) with `review-task` substituted for
  `sdd-verify`.
- `opencode-no-launch-fail.jsonl`: a mark with no launch or end anywhere in
  the trace. Not derived from the session (which always launches before
  marking); built directly from the criterion's own contract to exercise the
  base "no end available" case.
- `opencode-double-mark-fail.jsonl` / `opencode-batch-mark-fail.jsonl` /
  `opencode-reject-then-mark-fail.jsonl` / `opencode-pass.jsonl` /
  `opencode-accepted-pass.jsonl`: not derived from the session (which never
  exhibits a batch mark, a reject, or a D9-A acceptance); each is built
  directly from the criterion's own contract in design.md to exercise one
  specific branch (see `tests/fixtures/regression/README.md`'s "Add a case"
  and the branch-reversal table in design.md), using the session's own
  foreground-launch shape (`background=false`, a single line carrying both
  the launch and its completed status) for the "end" side once a fixture
  needs one.
- `claude-background-ack-fail.jsonl` / `claude-background-pass.jsonl`: not
  derived from any session; built from design.md's stated Claude shape
  (`Task`/`subagent_type`/`run_in_background`, `origin.kind=
  "task-notification"`) to exercise the same ack-vs-real-end ordering as the
  OpenCode pair, per design.md's own declared symmetry between the two
  hosts' background forms.
- `claude-foreground-failed-fail.jsonl` / `claude-foreground-pass.jsonl`:
  synthetic, no source session — the source session's two `subagent` calls
  (seq 1848) were both `background=true`, so it never exercises a foreground
  launch at all. Added in T5 fix round 1 (Codex review, AC6 `not met`): the
  first fixture's launch tool_result carries `"is_error":true` (Claude's own
  failed-result shape, `parseTrace`'s `message()` reading `is_error` via
  `truth(c["is_error"])` into `Success=false`), proving a foreground
  review-task launch that itself errored must never satisfy a later mark;
  the second is the same shape with a successful tool_result (no
  `is_error`), so the foreground path is exercised both ways. Neither
  fixture asks `run_in_background` at all, matching Claude's own foreground
  default per design.md.
- `claude-failed-mark-retry-pass.jsonl`: synthetic, no source session. Added
  after the change's `/code-review` found that a failed `tasks.md` edit
  counted as a mark and consumed the only verdict, so a correct retry was
  reported as `fail`. One review-task end, one `Edit` whose tool_result
  carries `"is_error":true` ("old_string not found"), then a successful
  retry; the criterion must pass. Removing the failed-edit exclusion in
  `isTasksMarkdownEdit` makes this fixture fail.

## Sanitization

Every fixture uses the generic path `openspec/changes/example-change/
tasks.md` (never the source session's real
`<specs>/.../<change-id>/tasks.md`), generic
task labels ("T1 - Example task", "T2 - Another example task"), synthetic
session/call/child ids (`ses-1`, `launch-1`, `edit-1`, `edit-2`, `child-1`),
and no customer, ticket, or secret text. No `oldString`/`newString`/
`old_string`/`new_string` value in any fixture reproduces text read from the
source session; every line's wording is authored fresh for this fixture set.
