# Design — frontier-prune

## Approach

One sensible way: prune by domain with build children working side by side on disjoint files, each from the inventory below, while the main thread changes the floor, `AGENTS.md`, the spec, the global core and every test that pins text. A child edits no test file, and no file under `site/` or `_support/docs/` except where its task names one: `tests/content/single_home_test.go` and its siblings are shared, so the main thread reconciles them after the children return, from each child's list of pinned phrases it changed or removed. Doing the whole prune in the main thread does not fit one context (about 150 files).

## Pruning test

For each paragraph, ask whether Opus 5.5 or gpt-6.2-sol would get the work wrong without it. Cut it when the answer is no, and in these classes:

- **Restatement**: it repeats a rule whose home is another file (the global index, a leaf, the questions leaf, the Delivery section). Replace with nothing, or a pointer when the reader needs the moment.
- **Tautology**: it restates its heading, the frontmatter, the table row that routes to it, or says nothing checkable.
- **Default**: generic engineering practice a frontier agent does unprompted (read before editing, check facts, inspect the stack, write clear tickets, a reviewer looks for bugs), and verification instructions the Opus 5 guide says cause over-verification.
- **Scaffolding**: exhaustive enumerations of obvious cases, fixed sequences where order does not matter, fallback chains for misreadings, documentation of a script's flags or validations that `--help` or the script's errors already give.
- **Verbose**: necessary but more than 40% longer than needed: rewrite shorter, keeping every condition and exception.

Never cut, whatever the class: authorization and scope of effects, secrets, irreversible effects, preservation, the delivery modes, and the «Keep» list below. A rule whose «never» or «always» guards an invariant keeps it; another rule rewritten in this change states scope and reason instead (AGENTS.md «Guidance structure», D5 A of 2026-11-22). A cut that removes the last statement of a behavior in its home is a removal, allowed by D3 B only for the Default and Scaffolding classes; a child reports each such removal.

## Keep (looks prunable, is not)

- `messages/reasons-and-blockers.md`: never give "a rule requires it" as the reason (core sentence).
- `messages/questions.md`: the decide-versus-ask matrix (a–f), the round and card format, the channel table.
- `build/prose.md`: the AI-tic list.
- `git/git-commit.md` and `git/merge-gate.md`: no `git add -A`; the human reviewer's pending approval holds the merge; never bypass branch protection.
- `build/defect-fix.md`: a guard at the symptom is not a fix.
- `plan/loops.md`: the baseline run before the first cycle; never relax what judges the loop.
- `delegation/delegation.md`: launch all children in one response; `delegation/reviewer-brief.md`: no producer reasoning in a reviewer's brief.
- `build/checks.md`: the pipe exit-status rule; `build/checks/run-check.md`: the disk-filling trap.
- `ui/viewer/host-session-identity.md` table; `tracker/trackers/jira.md` acli limits.
- `flow-plan`: a ticket's solution hint is a claim to verify; a mock or stub closes no criterion.
- `flow-build`: between tasks, continue with the next tool call.
- `flow-verify`: a sign-in page or empty list captured as evidence is a setup failure.
- `flow-deploy/references/deployment-checks.md`: prove the served identity on the provider.
- `flow-close/references/batch-archive.md`: wait for the next UTC minute before the second archive.
- `hive-mode/references/code-audit.md`: the `git cat-file -e` and `git diff --stat` suppression test.
- `git-workflow`: say whether a revert fully undoes the merge; `references/pr-captures.md`: an image on a public repository stays visible.
- `harness-audit`: host loading facts (`references/hierarchy.md` table, `references/memory.md`), `references/stack-skills.md` supply-chain check.
- `test-audit`: keepers before deletions and the mutation proof.
- `create-issue/references/existing-ticket.md`: do not edit a ticket the account did not author.
- `agents/hive-build-owner.md` (E2E authority), `agents/hive-research.md` return contract, `agents/hive-review-code.md` finding threshold and P2 for unrequested changes, `agents/hive-build-infra.md` migration-count gate, `agents/hive-build-data.md` EXPLAIN ANALYZE effect warning.
- The principle leaves, the «Candidates» section of `close-report`, the `create-report` HTML skeletons, the global «Rule leaves» index, the flow-step links to the leaves read at that moment, and the two review leaves `review-evidence` and `review-findings` as separate files (decisions in `proposal.md`).
- `global.md` «Preservation»: "Inspect relevant state before changing it" (an invariant).
- `build/project-verification-format.md:24`: commands are opaque data interpreted only after the flow's authorization checks; never embed credentials.
- `git/merge-gate.md:5` stays the home of the CI wait (one blocking command, never redirect to a file, one message when it ends); `build/background-wait.md` keeps its pointer to it.
- `messages/reasons-and-blockers.md:9`: the manual-action instructions and "offer that check as an option"; only the try-it part is a restatement.
- `harness-audit/references/instruction-files.md` HA-IF-12 (HTML comments cost tokens in `AGENTS.md`) and HA-IF-16 (`@path` imports): host-loading facts.
- The stage-packet entry's fallbacks: missing packet, end line, missing heading.

## Inventory by domain

Audit line numbers are from `32bef8c8`; reread the current file. Estimates in bytes.

### `hive-core` part A — build, messages, git, plan, delegation (~25 KB)

- About 25 leaf preambles: keep the first «Read…» sentence (pinned by `index_trigger_test`) and any definition; delete the "It applies…/Also read…" restatements.
- `build/project-verification-format.md:5-26`: validations `check_contract.py check` enforces; keep the field table, the `verified` rule, line 24 and the phrase `single_home_test.go` pins at line 26.
- `build/background-wait.md:5-8`: Verbose parts; keep the deadline, wake-in-stretches and blocking-command rules and the pointer to `merge-gate` for CI.
- `build/security-boundaries.md:5-7`: Default; keep the trigger and the `hive-review-security` pointer.
- `build/tests.md:19-22`: halve the rejected-test list; the sleep bullet goes. `build/tests/refactor.md` merges into `tests.md`.
- `build/prose.md:17-29`: "Write the sentence" and "Pick the mode" go; keep the AI-tic list.
- `build/checks.md:7`: one sentence; `build/tests-no-runner.md`, `build/stored-data.md`: Default parts.
- `messages/reasons-and-blockers.md:5`: rewrite the 1,350-byte bullet in one sentence keeping the pinned phrase; on line 9 only the try-it clause restates the verification floor.
- `messages/questions/unanswered.md:9-16`: the five-way ladder becomes one paragraph; keep the 120 s timeout line.
- `messages/report-endings.md:5,51,70-81`: tautological intro; compress the compaction reply and the close-footer list.
- `messages/questions.md:17,50-54`: one sentence per rule; the matrix and format stay.
- `git/merge-gate.md:13`: restructure the admin-merge paragraph keeping its semantics; line 5 stays the CI-wait home, shortened only.
- `git/delivery-ask.md`, `git/git-commit.md`, `git/pr-reviews.md`, `git/foreign-commits.md`: Verbose and Tautology parts.
- `plan/checkpoint.md:5,11,13`: point to `tasks.py --help` for flags; keep `--unwritten`, `--rules` and the temp-file checkpoint; one-sentence reread order.
- `plan/reuse-before-writing.md:7-8,11`, `plan/scope-changes.md:14`, `plan/loops.md:12-14`, `plan/project-settings.md:9-19` (example repeats line 5), `plan/external-facts.md`, `plan/change-folder.md`, `plan/external-effects.md`.
- `delegation/child-brief.md:7,9`: line 7 as a short list; line 9 restates the wait rule. `delegation/delegation.md:24-28`: one sentence. `delegation/reviewer-brief.md:11`: link the criteria reference for states. `delegation/unattended-effects.md`.

### `hive-core` part B — review, business, close, local, tracker, ui, principles (~25 KB)

- About 15 leaf preambles, as in part A.
- `review/blast-radius.md:7-23`: drop the five-rung ladder; keep "name the one fact the safety rests on; run code that fails loudly".
- `review/no-comments.md:5-17`: one paragraph keeping the Constraint-to-enforcer idea and the `comment-audit` lint handling merged at `ec51a107`.
- `review/review-findings.md:5` becomes five bullets; `review/review-evidence.md` is trimmed in place. `review/suppression-review.md:3,7`, `review/interrogate.md`.
- `business/*`: Verbose parts only.
- `close/close-report.md`, `close/close-modes.md:13-21`, `close/close-hygiene.md`: Verbose; «Candidates» stays.
- `local/local-record.md:5-6`: drop the memory-tool fallback chain; `local/local-write.md:5-7`: point to the helper's `--help`; `local/local-script.md`.
- `tracker/trackers.md:9-13`, `tracker/trackers/linear.md`, `tracker/trackers/github.md`: keep the priority order, the state cache and the categories; drop the tool-name tables; `jira.md` stays. `tracker/ticket-state.md`, `tracker/trackers/time.md`.
- `ui/ui-verification/device-tool.md:6`, `ui/a11y-audit.md:5`, `ui/viewer-open.md:8`, `ui/ui-plan.md`, `ui/design-stage.md`, `ui/ui-example.md`, `ui/walkthrough.md`, `ui/viewer.md`, `ui/viewer-close.md`, `ui/before-captures.md`: Verbose and Default parts.
- `principles/index.md`: Tautology parts only; the leaves stay.

### Flow skills and `hive-mode` (~55 KB)

- Every flow `SKILL.md` entry paragraph: one shorter sentence that keeps the missing-packet, end-line, missing-heading and cut fallbacks and the "it holds the leaves …" list (`stage_manifest_test`, `tests/pilot/session_audit_packets.go`).
- Delete `hive-mode/playbooks/{investigate,bugfix,feature,hotfix,infra-ci,pr-review-deploy,deliverable}.md`; carry what is specific to each into its table row in `hive-mode/SKILL.md` (for example bugfix → `defect-fix` before the first edit, hotfix → urgent fixes reference, investigate → `code-audit` for an audit, deliverable → the Figma, design-review and PRD routes now repeated at `SKILL.md:16,29`). Every link to a deleted file is repointed. `hive-mode/SKILL.md:51` and the entry lines of `flow-build/SKILL.md:10` and `flow-verify/SKILL.md:16` say that a kind without a file is read from the row of `hive-mode/SKILL.md#pick-the-playbook`.
- The six remaining playbooks: delete the «Reporting», «Leaves» and delivery-mode paragraphs (the «Reporting» removal is listed in `proposal.md`); reduce «Handoffs» to one line.
- Each flow's `Reply:` line: one short line linking its report row (`reply_line_test`).
- Flow-step leaf links stay; only prose around them that restates the leaf's content goes.
- `hive-mode/SKILL.md:23,25,27-30,69-73`; `hive-mode/references/code-audit.md:7,22-30,38-42,60` (keep the Hive-specific clauses and the heading anchors).
- `flow-plan/SKILL.md:14,23,27`; `references/plan-handoff.md:3,28`; `references/plan-review.md:19,31`; `references/execution-record.md:9,58-60,64,66,96` (keep the manifest example, the verifier-independence rule and a `--help` pointer).
- `flow-build/SKILL.md:19,24,26`; `references/e2e-coordination.md:3,9,21,27` (delete only what global «Delegation» states).
- `flow-verify/SKILL.md:18,25`; `flow-close/SKILL.md:14`; `references/branch-cleanup.md:23-27`; `batch-archive.md`; `support-material.md`.
- `flow-deploy/SKILL.md:22,31`; `references/deployment-checks.md:7,9-11,21-30` (one home for rollback triggers); `urgent-fixes.md:11`.
- `git-workflow/SKILL.md:12,14,15,16,21,24`; `references/pr-captures.md:29-31`.
- `unattended-delegation/SKILL.md:16,17-18` (keep every authorization exception).

### Other skills (~55 KB)

- `harness-audit`: `SKILL.md:39-59` outcomes and severity tightened, `:67-79` «Ask» points to the questions leaf; `references/instruction-files.md` HA-IF-02, 04, 06, 07 merged into HA-IF-01, HA-IF-11 dropped, HA-IF-12 and HA-IF-16 kept, HA-IF-08 and HA-IF-17 shortened; `references/hierarchy.md:5-13,52-65`; `references/skills-and-agents.md:7,9,12,18-21` (keep HA-SK-02, SK-04, AG-01, AG-05); `references/native-tools.md:5-10`; `references/local-instructions.md:44-53`; `references/personal-rules.md:19-31`; `references/transcript-sources.md:1-13`. Each retired `HA-*` rule is marked retired in `_support/docs/harness-engineering/harness-audit-rules.md` (`tests/skills/harness_audit_rules_test.py`).
- `create-design`: `references/refinement.md` deleted (link from `hive-design-ui` repointed); `references/sections.md:5-48` cut by about half; `references/design-library.md:9-24` to one sentence of precedence; `references/design-record.md:5-18`; the unattended paragraph kept only in `references/special-cases.md`; `SKILL.md:23-34` step 5 shortened.
- `project-layout`: `references/layout.md:52-57`; `SKILL.md:18-36` (keep the undo rule).
- `create-prd`: `references/documents.md:19-30`; Verbose parts elsewhere.
- `create-issue`: `references/ticket-types.md:5-26`, `references/trackers.md:9-24` (keep the hint rules).
- `create-report`: `references/diagram-grammar.md` and `references/diagram-svg-construction.md` trimmed to about 8 KB together; skeletons unchanged.
- `hive-typescript`: rules 4–9 cut to their non-default parts (keep brands, no `as`, honest guards); keep `paths` and model invocation (`leaf_skill_test`).
- `test-audit`, `workspace-repo`, `comment-audit`: Verbose parts only.

### Roles (~25 KB)

- The boundary footer in all 21 roles: at most 65% of its base size, still starting with «Stay within the parent», one sentence per topic; `tests/content/agent_boundary_test.go` constants change with it (T8).
- `hive-build-backend`, `-data`, `-infra`, `hive-write-tests`, `hive-tune-performance`, `hive-design-architecture`: body to its non-default lines and leaf links (AC8).
- `hive-review-ux.md:10` frame-review paragraph moves to `create-design/references/review.md` with a pointer; `:14` compressed. `hive-review-code.md:12`, `hive-verify-task.md:14`, `hive-read-state.md:11-15`, `hive-review-plan.md:11`, `hive-build-frontend.md:13,17`, `hive-build-mobile.md:11,13`, `hive-research.md:10-12`.

### Global core (~3 KB)

- «Scope and authorization»: shorten the ticket-moves bullet and the critical-effect sentence keeping every condition. «Preservation» keeps every sentence. The index keeps one line per leaf.

## Floor

| Row | Writing target | Origin |
| --- | --- | --- |
| claude `opus` (`claude-opus-5-5`) | yes | `reasoning`, `verifier` |
| codex `gpt-6.2-sol` | yes | `execution`, `reasoning`, `verifier` |
| claude `sonnet` (`claude-sonnet-5-5`) | no | `execution`, `focused`; tier below |
| codex `gpt-6.1-terra` | no | `focused`; tier below |
| claude `claude-fable-5-2` | no | tier above |
| codex `gpt-6.2-astra` | no | tier above |
| grok `grok-4.7` | no | `verifier` (D6 A) |
| grok `grok-4.7-build-fast` | no | user's main session |

`model_floor_test.go` changes from "every profile model is a writing target and every target names a profile" to: every profile model has a floor row; the writing-target rows are exactly the backquoted Model values the «Writing target» bullet of «Two sets» lists, a bullet that carries no other backquoted text; and the flash-class list no longer feeds the expected rows. Its drift cases replace the two that die ("flash model marked as writing target", "stale writing-target row") with "profile model without a row" and "target row «Two sets» does not list". «Two sets» keeps the clause that adds a main-session model on the user's yes, except a Haiku, Luna or flash-class model. `tests/pilot/qualify_run_test.go` copies the real floor document: `newQREnv` takes a fixed floor document as `qualify_score_test.go` does, since its OpenCode cases have no row left.
