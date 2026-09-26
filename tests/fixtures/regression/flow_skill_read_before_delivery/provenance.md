# Provenance — flow_skill_read_before_delivery

Derived 2026-09-26 from `gh-33-flow-skill-routing/proposal.md`'s "Contexto
verificado" (globex G6) and targeted `jq` field extraction (`type`, tool-call
`name`, argument key names, and boolean `test(...)` checks on the command/
path strings only). No transcript line, skill body, or command/report text
was printed or pasted verbatim into this repository, its provenance, or any
brief; every fixture below is a hand-written, sanitized re-expression of the
observed shape in `parseTrace`'s own streamed wire format
(`streaming-messages-json` for Grok, `codex exec --json` items for Codex),
never a copy of the source line.

## grok-fail.jsonl (globex G6)

Source: Grok globex session `01a0daf9-ddec-7421-ad4e-7b7e439dd873`
(grok-4.7-build-fast high), `chat_history.jsonl` (557 lines total; the same
session as `close_question_after_report`'s G3 and `merged_branch_deleted`'s
G4).

Verified via `jq` (`type`, `tool_calls[].name`, argument key names, and a
boolean `test(...)` on the `target_file`/`command` values only):

- line 306: `assistant`, two tool calls — `read_file` (`target_file` ending
  in `git-workflow/SKILL.md`, confirmed with `test("git-workflow/SKILL.md$")`)
  and `run_terminal_command` (unrelated).
- line 307: `tool_result` for the `read_file` call, `is_error` absent
  (success), content length 5,805 — never read.
- line 334: `assistant`, two tool calls — `search_replace` and
  `run_terminal_command`, the latter matching `test("git commit")` — the
  session's first commit.
- lines 335–336: `tool_result`s for both, `is_error` absent (success).
- line 382: `assistant`, two `run_terminal_command` calls; the first matches
  both `test("gh pr merge")` and `test("delete-branch=false")` (already
  documented in `merged_branch_deleted/provenance.md`'s G4 entry, same
  session, same line).

No `read_file`, literal shell `cat`, or native skill payload targeting
`flow-build/SKILL.md` occurs anywhere in the 557-line file (a file-wide `jq`
scan of every `read_file` call's `target_file` for `test("flow-build")`
returned no match) — only `git-workflow` was read, at line 306, and only
after `git-workflow`, never `flow-build`, precedes the commit at line 334.

`grok-fail.jsonl` re-expresses this shape: a `read_file` of
`git-workflow/SKILL.md` (stand-in body, not the real 5,805-character source),
then a `git commit`, then a `gh pr merge … --delete-branch=false` — matching
globex G6 exactly. The commit and merge commands are placeholders (`fix:
example change`, PR `42`); the original command text and repository content
were never read for this fixture beyond the boolean tests above.

## grok-pass.jsonl (`grok-b-2`)

Source: pilot run `grok-b-2` (`close-sequence`, arm B),
`_support/workspace/2026-09-26-work-close-sequence/runs/grok-b-2/stdout.jsonl`
(33 lines). Located with `jq`, extracting only `type`, tool-call `name`,
`target_file`/`command` boolean tests, and `is_error`:

- line 2: `assistant`, two `read_file` calls; the second (`target_file`
  matching `test("flow-build/SKILL.md$")`) reads `flow-build/SKILL.md`.
- line 3: two `tool_result`s, both `is_error: false`.
- lines 4–16: further `read_file`/`run_terminal_command`/`todo_write`/
  `search_replace` calls, none matching `test("git commit|git push|gh pr
  create|gh pr merge")`.
- line 18: `assistant`, one `run_terminal_command` matching
  `test("git commit")` — the run's first Git delivery action.
- line 19: `tool_result`, `is_error: false`.

`flow-build/SKILL.md` is read (line 2) well before the first Git action
(line 18). `grok-pass.jsonl` re-expresses only the two events the criterion
needs: the `read_file` of `flow-build/SKILL.md` (stand-in body) followed by
the `git commit` — matching this pilot run's shape without reproducing its
intervening tool calls or any of its real command/report text.

## codex-pass.jsonl (`codex-b-1`)

Source: pilot run `codex-b-1` (`close-sequence`, arm B, the same run
`merged_branch_deleted/provenance.md` already documents), 46-line
`stdout.jsonl`. Located with `jq`, extracting only `item.id`,
`item.status`/`item.exit_code`, and boolean `test(...)` on `item.command`/
`item.aggregated_output`:

- `item_1` (line 5): `command_execution`, `status: completed`, `exit_code:
  0`; `command` matches `test("cat.*flow-build/SKILL.md")` and
  `aggregated_output` matches `test("name: flow-build")` — a literal shell
  `cat` of `flow-build/SKILL.md` with its front matter in the output.
- `item_17` (line 35): `command_execution`, `status: completed`; `command`
  matches `test("git commit")` — the run's first Git delivery action.
- `item_18` (line 37): `command_execution`, `status: completed`; `command`
  matches `test("git push")`.
- Every `command_execution` between `item_1` and `item_17` (`item_2`…
  `item_15`) was checked and matches none of
  `test("git commit|git push|gh pr create|gh pr merge")`.

`codex-pass.jsonl` re-expresses `item_1`, `item_17` and `item_18` in the
`codex exec --json` `item.completed`/`command_execution` shape `parseTrace`
reads (id, `command`, `status`, `exit_code`, `aggregated_output`). The read
occurs (`item_1`) before the first Git action (`item_17`). The command's real
path (under this pilot's own disposable `shadow-home` fixture directory, not
a customer path) is sanitized to the relative `.agents/skills/flow-build/
SKILL.md`, and `aggregated_output` is replaced with a short stand-in body
carrying `name: flow-build`, `description:` and a `# ` heading, per
`skillContentReads`' literal-read match requirement — not the pilot's real
12,957-byte skill source.

## Declared limits (A2)

Per `proposal.md`'s A2: this criterion does not see a skill invoked through a
typed slash/dollar command (`/flow-build` in Claude, `$flow-build` in Codex —
no fixture case is registered for that shape, since `parseTrace` has no
distinguishable event for it), and it does not observe a deployment performed
with no Git action at all (globex's G5, the blank production page after
deploy) — out of scope for a criterion keyed on Git delivery actions. It is
declared in `tests/pilot/regression.go` but intentionally not returned by
`regressionCriteria` or wired into `assessFlows`.

## Wire-format note

Same session as `close_question_after_report/provenance.md` and
`merged_branch_deleted/provenance.md`: this Grok CLI session calls
`read_file` and `run_terminal_command` directly as top-level tool names, not
wrapped in a `use_tool` envelope.
