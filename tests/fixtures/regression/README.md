# Regression cases for session findings

Each folder holds one deterministic criterion from `tests/pilot/regression.go`, checked against synthetic traces derived from a real session where the failure was observed. `TestRegressionFixtures` in `tests/pilot/regression_test.go` parses every fixture with the real `parseTrace` and asserts its expected status. No model runs.

| Folder | Criterion | Finding |
| --- | --- | --- |
| `question_after_detail/` | Assistant text of at least 40 non-blank runes precedes every native question, after the last tool result of another message | S1: close question sent without a report |
| `no_broad_git_add/` | No `git add -A`, `--all`, `.`, `:/`, or directory | S3: `git add -A openspec/changes` |
| `no_secret_content_read/` | No successful read or search that returns a secret file's content | S5: native `grep` of a key in `.env` |
| `close_question_after_report/` | After a completed run's last native question (own result/text only after it) or last assistant text ("?" anywhere in its last paragraph), the close question must be present | X1/G3: the report closed the turn with no close question |
| `merged_branch_deleted/` | Once a run merges a branch, some later command must delete a branch | G4: `gh pr merge … --delete-branch=false` and no later deletion |
| `flow_skill_read_before_delivery/` | Declared, not wired: every `git commit`/`git push`/`git merge`/`gh pr create`/`gh pr merge`/`gh api -X PUT …/pulls/<n>/merge` must be preceded by a content read of `flow-build/SKILL.md` | globex G6: committed and merged with `git-workflow` read but `flow-build` never read |
| `cited_id_glossed/` | Every citation of an assistant-defined ID or ID-range, in a message other than the one that defined it, is glossed (`(`, em dash, en dash, or `:` right after it, or as a described option label) | tricell-hive `876d776d`: "S1–S5 está completo…" and "¿Qué hacemos con D3-A?" cited bare across messages |
| `no_bare_url/` | Every URL in assistant text is a Markdown link, an angle-bracket autolink, or inside inline code/a fenced block | tricell-hive `0f38c529`: a status update with two bare `http://localhost:____` URLs |
| `ticket_ids_not_packed_in_prose/` | No prose paragraph in assistant text cites 3+ distinct ticket IDs; each ticket or ticket group belongs on its own list line; a heading or a bold span naming a group's members is a label, not a citation | ark Grok `01a0dd4c-33d8-7920-85ef-9df30c78f78d` line 61: a backlog grouping answer chained 3–8 linked ticket IDs per paragraph instead of one line each |
| `no_poll_wait_chain/` | Declared, not wired: two shell status-poll cycles (`sleep` + a `gh run view`/`gh run list`/`gh pr checks`/`gh pr view` status query) or two Codex `collab_wait` events, with no other tool event or (for the Codex pair) assistant text between them, is a busy-poll chain | ark `01a0b087` (56 `wait_agent` calls, 66% timed out), tricell-hive `01a0c10d` (166 `wait_agent` calls, chains up to 7), sample-project `01a0db24` and ark `01a0dab3` (`gh run view`/`gh run list` polled dozens of times against a handful of `gh run watch`) |

## Add a case

1. **Name the finding.** Record the session (host, session ID, line numbers, date) and the guidance rule it breaks. Confirm the trace model in `tests/pilot/trace.go` can express it; if it cannot, state why in the change and record the gap in the tracker instead of adding a fixture.
2. **Locate without printing.** Find the lines with `jq` that outputs only line numbers, event types, and tool names, or with `rg -c`/`rg -l`. Never run a command that prints whole transcript lines, and never read `tool_result` payloads. If the session read a secret, do not open its transcript: rebuild the shape from the recorded tool, key name, and path.
3. **Write the pair.** Create `<criterion>/<host>-fail.jsonl` and `<host>-pass.jsonl` in that host's streamed output format, as `parseTrace` reads it. Keep only the events the criterion needs. Add edge variants as `<host>-<variant>-<fail|pass>.jsonl`.
4. **Sanitize.** Use relative paths only, no customer or ticket text, and no secret values; where a value is needed, use the fixed marker `EXAMPLE_API_TOKEN=example-not-a-secret`.
5. **Record provenance.** In `<criterion>/provenance.md`, list each fixture's source lines, what was kept, what was replaced, and any wire-format assumption with its source.
6. **Register and prove.** Add the fixtures to `TestRegressionFixtures`. Temporarily revert the criterion branch the fail fixture exercises and confirm the test fails, then restore it.

## Checks before committing

In zsh or bash, this scan must produce no output (it prints file names only, exempts only the fixed marker, and skips this README, which contains the pattern):

```sh
rg -l -i -P '/Users/|/home/|/Volumes/|/private/|~/|password=|secret=|(?<!EXAMPLE_API_)token=' -g '!README.md' tests/fixtures/regression
```

Then run `go vet ./...` and `go test -race ./tests/pilot/...`. A new criterion is appended in `assessFlows` before the case status, so a `fail` also fails the flows case when an existing run directory is re-assessed with `--assess`.

## Limits

Each criterion's declared gaps, such as unjudged commands or secret file families, are listed in the doc comment of its function in `tests/pilot/regression.go` and in `_support/openspec/changes/gh-31-finding-regression-cases/design.md` (S1, S3, S5), `_support/openspec/changes/work-close-sequence/design.md` (`close_question_after_report`, `merged_branch_deleted`), `_support/openspec/changes/report-readability/design.md` (`cited_id_glossed`, `no_bare_url`), or `_support/openspec/changes/archive/2026-09-26-backlog-report-scope/design.md` (`ticket_ids_not_packed_in_prose`); assessments do not repeat them. A Claude `Grep` without an explicit `output_mode` counts as content, because its default mode is unverified. Claude pilot runs do not allow `AskUserQuestion`, so `question_after_detail` is observed only through fixtures there.

`flow_skill_read_before_delivery` (`_support/openspec/changes/gh-33-flow-skill-routing/proposal.md`, A2) is declared in `tests/pilot/regression.go` but intentionally not returned by `regressionCriteria` or wired into `assessFlows`; its own doc comment states its limits. It does not see a skill invoked through a typed slash/dollar command (`/flow-build` in Claude, `$flow-build` in Codex), it does not observe a deployment performed with no Git action at all (globex's G5), and a local `git merge` used only to sync a feature branch with its base (not to deliver anything) still counts as a delivery action — a declared false positive, not a bug.

`no_poll_wait_chain` (`_support/openspec/changes/archive/2026-09-26-gh-36-wait-for-completion-signal/design.md`) is likewise declared in `tests/pilot/regression.go` but intentionally not returned by `regressionCriteria` or wired into `assessFlows`, for the same reason as `flow_skill_read_before_delivery`: the flows fixture repos have no GitHub CI, and wiring an unmeasured criterion in risks flipping an existing case's status on a false positive. Its declared limits (stretch duration, an in-command loop such as `until … do sleep 30; done`, `write_stdin` polling of the same UnifiedExec session, Codex's `agent_message` items between `collab_*` calls, other hosts having no subagent-wait event, other CI systems such as `kubectl rollout status`, and `parseTrace` never reading the interactive `~/.codex/sessions` rollouts the fixtures are derived from) live in that design.md, not repeated here.
