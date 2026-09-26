# Provenance — merged_branch_deleted

Derived 2026-09-25 from `design.md`'s "Contexto verificado" (G4) and
targeted `jq` field extraction (tool-call `name` and a boolean
`test(...)` on the command string only — no command or result text was
printed). No transcript line was pasted verbatim into this repository, its
provenance, or any brief.

## grok-fail.jsonl (G4)

Source: Grok globex session `01a0daf9-ddec-7421-ad4e-7b7e439dd873`
(grok-4.7-build-fast high), `chat_history.jsonl` (557 lines total; the
same session as `close_question_after_report`'s G3).

Verified via `jq` (`tool_calls[].name` and boolean `test(...)` on the
command string only):
- line 382: `assistant`, two `run_terminal_command` calls. The first
  matches `test("gh pr merge")` and `test("delete-branch=false")`, both
  `true` — `gh pr merge <n> --merge --delete-branch=false`; the second
  matches neither pattern.
- lines 383–408 (`awk 'NR>382 && NR<=408' … | jq '.tool_calls // [] |
  map(.arguments | fromjson? | .command // "" | test("branch
  -d|push.*--delete|push.*:.*|gh pr merge"))'`): every command tested
  `false` — no further merge or branch-deletion command follows before the
  session's final report at line 408 (see
  `close_question_after_report/provenance.md`).

`grok-fail.jsonl` re-expresses this shape in the `streaming-messages-json`
format `parseTrace` reads: a single successful shell call running
`gh pr merge <n> --merge --delete-branch=false`, with no later deletion
command — matching G4 exactly (the report's missing cleanup line and the
`git bundle` design.md also names are outside this criterion's trace model
and are not reproduced here).

## grok-pass.jsonl (constructed)

Not derived from a transcript: G4 is the only branch-cleanup finding on
file, and it is a failure. Built directly from the criterion's own contract
(`design.md` → "Criterios deterministas" → `merged_branch_deleted`) to
exercise the pass path: a successful `gh pr merge <n> --merge` (no delete
flag) followed by a successful `git push origin --delete <branch>`, so a
deletion is observed after the last merge.

## codex-pass.jsonl and codex-fail.jsonl (codex-b-1)

Source: pilot run `codex-b-1` (`close-sequence`, arm B), 46-line
`stdout.jsonl` (`_support/workspace/2026-09-26-work-close-sequence/runs/codex-b-1/stdout.jsonl`).
Located with `jq`, extracting only `item.status`/`item.exit_code` and a
boolean `test("merge|branch|push|delete")` on the command first, then the
command text itself for the matching, successful (`status=="completed"`)
items only — no other field or transcript line was read:

- `item_18`: `git push fixture fix/greeting-punctuation …` — publishing the
  fix branch, not itself a merge or deletion.
- `item_19`: `git switch main && git merge --ff-only fix/greeting-punctuation
  && npm test` — the local merge.
- `item_21`: `git push fixture main …` — the confirming push to base.
- `item_22`: `git push fixture --delete fix/greeting-punctuation && git
  branch -d fix/greeting-punctuation …` — the branch deletion, both remote
  and local, in the same command.

`codex-pass.jsonl` re-expresses `item_19`, `item_21` and `item_22` in the
`codex exec --json` `item.completed`/`command_execution` shape `parseTrace`
reads (id, `command`, `status`, `exit_code`; `aggregated_output` empty since
this criterion never reads it). `codex-fail.jsonl` is the identical shape
with `item_22` (the deletion) removed — a merge with no qualifying deletion
after it, per finding 10 (`/code-review` on `work-close-sequence`). Field
names and command text are the pilot fixture's own generic `fix/<slug>`
branch and `npm test` commands from `tests/fixtures/flows/cases.json`'s
`close-sequence` case, not any real project's business content.

## Wire-format note

Same as `close_question_after_report/provenance.md`: this session's Grok
CLI calls `run_terminal_command` directly, not wrapped in `use_tool`.
