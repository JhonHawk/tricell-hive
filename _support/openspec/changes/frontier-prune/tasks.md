# Tasks

Base `development` at `34395b92`, branch `refactor/frontier-prune`. T1, T2, T7 and T8 run in the main thread: they change the floor, the repository rules, the tests every other task depends on and the public and architecture text. T3–T6 run in `hive-write-docs` children side by side, on disjoint files, each briefed with this folder; a child edits no test file and no file under `site/` or `_support/docs/` unless its task names it, and returns the pinned phrases it changed or removed, each Default or Scaffolding removal of a behavior's last statement, each link it needs changed in another task's file, and its byte count before and after. Each task's checks run in the main thread. The per-task byte targets remove at least 168.4 KB together (T3 22,715, T4 23,199, T5 49,644, T6 72,558, T7 909), the reduction the first size criterion needed.

### T1 — Floor and public pages

- [x] The floor names Opus 5.5 and gpt-6.2-sol as the writing target, with one tier up and down, Grok 4.7 and `grok-4.7-build-fast`; README, llms.txt and the site home state it.

**Closes:** AC1, AC6

**Depends on:** none

**Locations:** `_support/docs/harness-engineering/model-floor.md`, `tests/content/model_floor_test.go`, `tests/pilot/qualify_run_test.go`, `_support/docs/harness-engineering/rule-enforcers.md`, `_support/docs/harness-engineering/session-review-rubric.md` (only if its main-session clause changes), `README.md`, `llms.txt`, `site/src/content/docs/index.mdx`, `_support/openspec/changes/frontier-prune/specs/model-qualification/spec.md`

**Execution:** main thread

**Test approach:** check — `TestModelFloor` and `TestModelFloorDetectsDrift` own the floor contract; their drift cases change as `design.md` «Floor» sets. `qualify_run_test.go`'s `newQREnv` takes a fixed floor document, as `qualify_score_test.go` does with `qsFloorDoc`, since its OpenCode effort cases have no row left on the new floor.

**Changes:** Rows as `design.md` «Floor» sets, with Fable's and Astra's IDs and aliases read from current vendor documentation and recorded in the verified context; `model-floor.md:3,8,10,16,38` rewritten for the new floor; superseded and small-model rows go. `integrations/agent-profiles.json` does not change. The pilot runtime code does not change.

**Verification:** `go test ./tests/content -run 'TestModelFloor' -v` and `go test ./tests/pilot -run 'Qualify'` pass; `rg -c 'gpt-6-astra|gpt-6.1-sol|grok-4.6|claude-haiku|luna' _support/docs/harness-engineering/model-floor.md` prints nothing; `rg -c 'claude-fable-5-2|gpt-6.2-astra' _support/docs/harness-engineering/model-floor.md` prints at least 2; `rg -n -i 'haiku|flash-class' README.md llms.txt site/src/content/docs/index.mdx` prints nothing; `rg -c 'Opus 5.5' README.md llms.txt site/src/content/docs/index.mdx` prints a count for each file.

### T2 — Measurement and authoring rules

- [x] «Measurement» accepts vendor guidance with the user's decision as evidence to relax or remove a rule; «Guidance structure» drops the two sentences AC2 names.

**Closes:** AC2

**Depends on:** none

**Locations:** `AGENTS.md`, `_support/openspec/changes/frontier-prune/specs/model-qualification/spec.md`

**Execution:** main thread

**Changes:** `AGENTS.md:22` names the writing target as the floor document's «Two sets» and keeps "a plausible misreading by a smaller model is not enough"; its sentence on vendor guides points to «Measurement»; `:21` reason becomes "since models apply instructions literally and act on what they can check"; «Measurement» adds the third basis and keeps the restatement clause. The repository-wide search of AGENTS.md «Guidance structure» (global guidance, skills, roles, `site/`, `_support/docs/architecture/`, `tests/pilot/`) runs for these sentences.

**Verification:** `rg -n "alone does not justify removing a rule|literal and smaller models" AGENTS.md` prints nothing; `rg -n "prompting guidance" AGENTS.md` prints the new «Measurement» line.

### T3 — Prune `hive-core` part A (build, messages, git, plan, delegation)

- [x] The five domains lose the inventory's restatements, tautologies, default and scaffolding text, with the «Keep» list intact.

**Closes:** AC7

**Depends on:** none

**Locations:** `content/skills/hive-core/{build,messages,git,plan,delegation}/**/*.md`

**Execution:** `hive-write-docs`

**Changes:** As `design.md` «Inventory» part A. Does not change scripts, `assets/`, the first «Read…» sentence of any leaf, or the global index.

**Verification:** `find content/skills/hive-core/build content/skills/hive-core/messages content/skills/hive-core/git content/skills/hive-core/plan content/skills/hive-core/delegation -name '*.md' -exec cat {} + | wc -c` at most 80,000 (base 102,715); `go test ./tests/content/...` after T8.

### T4 — Prune `hive-core` part B (review, business, close, local, tracker, ui, principles)

- [x] The seven domains lose the inventory's restatements, tautologies, default and scaffolding text, with the «Keep» list, the principle leaves and both review leaves intact.

**Closes:** AC7

**Depends on:** none

**Locations:** `content/skills/hive-core/{review,business,close,local,tracker,ui,principles}/**/*.md`

**Execution:** `hive-write-docs`

**Changes:** As `design.md` «Inventory» part B. Does not change scripts, `assets/`, the first «Read…» sentence of any leaf, or the global index.

**Verification:** `find content/skills/hive-core/review content/skills/hive-core/business content/skills/hive-core/close content/skills/hive-core/local content/skills/hive-core/tracker content/skills/hive-core/ui content/skills/hive-core/principles -name '*.md' -exec cat {} + | wc -c` at most 81,000 (base 104,199); `go test ./tests/content/...` after T8.

### T5 — Prune the flow skills, `hive-mode` and its playbooks

- [x] Seven playbooks are gone and routed by `hive-mode`'s table; the flows lose tautological reply and packet text, restatements and verbose paragraphs, and keep their leaf links.

**Closes:** AC7, AC4

**Depends on:** none

**Locations:** `content/skills/{hive-mode,flow-plan,flow-build,flow-verify,flow-close,flow-deploy,git-workflow,unattended-delegation}/**/*.md`

**Execution:** `hive-write-docs`

**Changes:** As `design.md` «Inventory» flow section. Keeps each `Reply:` line, the packet fallbacks and "holds" list, the flow-step leaf links and the `Playbook` field.

**Verification:** `ls content/skills/hive-mode/playbooks/` lists six files; `rg -n "playbooks/(investigate|bugfix|feature|hotfix|infra-ci|pr-review-deploy|deliverable)\.md" content` prints nothing; `rg -c "^\| \`(feature|ui-change|bugfix|hotfix|data-change|dependency-security|infra-ci|investigate|ticket-status|deliverable|pr-review-deploy|refactor|perf)\`" content/skills/hive-mode/SKILL.md` prints 13; the byte measure over these skills at most 143,000 (base 192,644).

### T6 — Prune the other skills and the roles

- [x] The other skills and all 21 roles lose the inventory's text; the six generic role bodies are at most 900 bytes; retired `HA-*` rules are marked retired in the catalog.

**Closes:** AC7, AC8

**Depends on:** none

**Locations:** `content/skills/{harness-audit,create-design,project-layout,create-prd,create-report,create-issue,test-audit,workspace-repo,hive-typescript,comment-audit}/**/*.md`, `content/agents/**/*.md`, `_support/docs/harness-engineering/harness-audit-rules.md`

**Execution:** `hive-write-docs`

**Changes:** As `design.md` «Inventory» other skills and roles. Role names, frontmatter and profiles unchanged; each shortened footer still starts with «Stay within the parent», and a footer size of 0 in the check fails it; HTML skeletons and scripts unchanged; HA-IF-12 and HA-IF-16 kept.

**Verification:** the byte measure over these skills and `content/agents` at most 300,000 (base 372,558); `python3 -I -c "import sys; [print(len(t.partition('\\n# ')[2].partition('Stay within the parent')[0].encode()), len(t.partition('Stay within the parent')[1].encode()+t.partition('Stay within the parent')[2].encode()), f) for f in sys.argv[1:] for t in [open(f).read()]]" content/agents/*/*.md` prints a body of at most 900 for the six generic roles and, for every role, a footer at most 65% of its base value in `proposal.md` «Verified context»; the footer keeps its first words «Stay within the parent»; `python3 -m unittest discover -s tests/skills -p '*_test.py'` passes.

### T7 — Global core

- [x] `global.md` loses its verbose sentences without losing a condition.

**Closes:** AC7

**Depends on:** none

**Locations:** `content/guidance/global.md`

**Execution:** main thread

**Changes:** As `design.md` «Global core». «Preservation» and every invariant sentence keep every condition; the index keeps one line per leaf.

**Verification:** `wc -c content/guidance/global.md` at most 25,000 (base 25,909); `go test ./tests/content -run 'TestGlobalGuidance|TestIndexTriggers' -v` passes after T8.

### T8 — Reconcile pinned tests and documentation, run the suite

- [x] Every content test passes against the pruned text, each pinned phrase change is deliberate, and the public and architecture text describes the playbooks, packets and roles as they now are.

**Closes:** AC7, AC4, AC8

**Depends on:** T1, T2, T3, T4, T5, T6, T7

**Locations:** `tests/content/*_test.go` (`single_home`, `leaf_skill`, `leaf_links`, `stage_manifest`, `agent_boundary`, `reading_footprint`, `index_trigger`, `reply_line`, `cited_targets`, `rule_enforcers`), `README.md:31-33,429-430`, `llms.txt:36`, `site/src/content/docs/skills.md`, `site/src/content/docs/flows/hive-mode.md` and other `site/` pages that link a deleted playbook, `_support/docs/architecture/instruction-resources.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check — the content tests own the text contracts; a row changes only where a child reported the phrase moved or removed.

**Changes:** Drop `leaf_skill_test.go`'s «Reporting» check and the deleted playbook kinds' file checks; update `agent_boundary_test.go` constants to the shorter footer; update the packet wording `stage_manifest_test.go` asserts and `instruction-resources.md` states; remove the playbook «Reporting» row of `rule-enforcers.md`; add the `[Unreleased]` entry. Run the repository-wide search of AGENTS.md «Guidance structure» for every rule the prune removes or moves and list the places checked in the report.

**Verification:** `go vet ./... && go test ./... && python3 -m unittest discover -s tests/skills -p '*_test.py'` passes; `rg -n "playbooks/(investigate|bugfix|feature|hotfix|infra-ci|pr-review-deploy|deliverable)\.md|one of 13|13 .*playbook" content site _support/docs README.md llms.txt` prints nothing; `find content -name '*.md' -exec cat {} + | wc -c` at most 715000; each «Keep» item of `design.md` found by `rg` in its home.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./... && go test ./... && python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Size and «Keep» list | Final candidate, locally | `find content -name '*.md' -exec cat {} + \| wc -c` ≤ 715000; one `rg` per «Keep» item | Covered by the implementation |
| Code review | Final candidate, before the result is reported | Claude Code `/code-review` | The review answer |

Delivery is `hold`: the verified working tree is the result. If the user later integrates it, the maintainer refresh deploys the content (AGENTS.md «Refreshing a local installation»), and the first three real sessions that run the pruned guidance are checked with `go run ./tests/pilot --audit-session`; a failure there is recorded as a watch item.

## Checkpoint

- Updated: 2026-11-22
- Stage: close · hold pause close
- Last commit: 34395b92
- Open findings: none
- Next: user picks a delivery mode to integrate; then full close (archive, branch and worktree cleanup)
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/frontier-prune · refactor/frontier-prune
- Unwritten: none: every decision is in proposal.md; only the delivery choice is pending with the user
- Rules: flow-verify/stage-entry.md
- Principles: none

### Evidence

- before captures: none (no UI)
- go test ./tests/content -run 'TestModelFloor|TestLeafSkill|RuleEnforcer' · ok · uncommitted worktree
- AC walk: AC1 TestModelFloor ok, 8 rows; AC2 rg old sentences 0, new Measurement basis line 86; AC4 6 playbooks, 13 rows; AC6 Opus 5.5 in 3 pages, no haiku/flash-class; AC7 712928; AC8 six bodies 557-899, footers 71-78% · uncommitted worktree
- /code-review: 10 findings (6 important, 4 nit), all confirmed against base and fixed in one round: report-endings continuation-read sentence, global critical-effect clause, questions 'no tool call between', infra verify/no-bypass step, project-verification cleanup ownership, tests negative-case paragraph and inventory item, hive-mode read playbook before route, site hive-mode page, instruction-resources line 30 and rule-enforcers row, AGENTS.md parenthetical
- final suite: go vet ok; go test ./... ok; tests/content ok after last edit; python 412 OK; AC7 713749; AC8 infra body 897 · uncommitted worktree

## Review status and progress

2026-11-22 — Plan written; revised after the plan review (see `proposal.md`).
2026-11-22 — Revision after the build (D10 A): the first size criteria → AC7 (≤ 715,000) and AC8 (footers ≤ 80%); the per-task byte targets stay as recorded misses.
2026-11-22 — T1 · uncommitted · unverified: none
2026-11-22 — T2 · uncommitted · unverified: none
2026-11-22 — T7 · uncommitted · unverified: go test ./tests/content passes only after T8 reconciles the children's pinned phrases
2026-11-22 — T3 · uncommitted · unverified: byte target missed: 92,760 vs 80,000 (cut 9,955) without cutting rule conditions
2026-11-22 — T4 · uncommitted · unverified: byte target missed: 95,081 vs 81,000 (cut 9,118)
2026-11-22 — T5 · uncommitted · unverified: byte target missed: ~159 KB vs 143,000 (cut ~33.8 KB); main thread restored findings-retention and branch-naming defaults, added security-boundaries to the infra-ci row
2026-11-22 — T6 · uncommitted · unverified: byte target missed: 335,262 vs 300,000; role footers at 71-78% of base, not 65%; six bodies <= 900
2026-11-22 — T8 · uncommitted · unverified: first size criterion missed: content Markdown 712,928 vs 635,000; python suite last run before the final two md edits; project-settings example block restored for tooling/cli doctor tests
