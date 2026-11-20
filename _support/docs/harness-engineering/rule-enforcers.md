# Rule enforcers

Which of Hive's rules a mechanism enforces, and which still rest on prose. A rule moves from prose to a mechanism when root `AGENTS.md` «Measurement» marks its watch item `Enforceable:` with that mechanism; this table lists the result. It does not restate any rule: each row links its canonical home.

- **Kind:** `prevents` when the mechanism stops the wrong action or rejects its output before it lands; `detects` when it reports the failure after a session ran, through `go run ./tests/pilot --audit-session`.
- **Status:** `enforced` when the mechanism exists; `judgment` when no mechanism can decide the rule and prose stays its only enforcement.
- `tests/content/rule_enforcers_test.go` fails when a path in the Mechanism column does not exist. Started 2026-11-17 (change `determinism-candidates`).

| Rule | Home | Mechanism | Kind | Status |
| --- | --- | --- | --- | --- |
| A skill or reference stays at or under 16,000 bytes | root `AGENTS.md` «Guidance structure» | `tests/content/skill_size_test.go` | prevents | enforced |
| Internal links and their anchors resolve | `_support/docs/architecture/instruction-resources.md` | `tooling/management/references.go` | prevents | enforced |
| The floor table matches the role profiles | `_support/docs/harness-engineering/model-floor.md` | `tests/content/model_floor_test.go` | prevents | enforced |
| The guidance's reading footprint stays within its budget | root `AGENTS.md` «Guidance structure» | `tests/content/reading_footprint_test.go` | prevents | enforced |
| `HIVE.local.md` entries are dated, short and ignored by Git; the secret location keeps its modes | global guidance «Local instructions» and «Secrets» | `content/skills/harness-audit/scripts/local_instructions.py`, `tooling/cli/doctor_project.go` | prevents | enforced |
| A package is at least 7 days old and free of active advisories | global guidance «Scope and authorization» | `content/skills/flow-build/scripts/check_package.py` | prevents | enforced |
| A commit names its paths, keeps unrelated staged work and passes a secret scan | global guidance «Secrets» | `content/skills/git-workflow/scripts/git_state.py` | prevents | enforced |
| Close deletes only merged branches and clean worktrees | `content/skills/flow-close/references/branch-cleanup.md` | `content/skills/flow-close/scripts/cleanup_branches.py` | prevents | enforced |
| Close fast-forwards the local base only when nothing local is in the way | `content/skills/flow-close/references/branch-cleanup.md` | `content/skills/flow-close/scripts/cleanup_branches.py` (`base_update`) | prevents | enforced |
| An analysis ends with the lettered reply block | `content/skills/hive-rule-report-endings/SKILL.md` | `tests/pilot/session_audit_endings.go` (`analysis_reply_block`) | detects | enforced |
| An admin-merge grant lasts the session within its exclusions | `content/skills/hive-rule-merge-gate/SKILL.md` | none | — | judgment |
| A change folder covers its criteria and its business rules follow their grammar | `content/skills/flow-plan/SKILL.md` | `content/skills/flow-plan/scripts/check_change.py`, `content/skills/flow-plan/scripts/business_rules.py` | prevents | enforced |
| A task's record and the build's checkpoint keep their format | `content/skills/flow-build/SKILL.md` | `content/skills/flow-plan/scripts/tasks.py` | prevents | enforced |
| Deltas merge into the specs before a change folder is archived | `content/skills/flow-close/SKILL.md` step 4 | `content/skills/flow-close/scripts/apply_deltas.py` | prevents | enforced |
| A PRD keeps its sections and rule entries | `content/skills/create-prd/SKILL.md` | `content/skills/create-prd/scripts/check_prd.py` | prevents | enforced |
| A ticket keeps the sections of its type | `content/skills/create-issue/SKILL.md` | `content/skills/create-issue/scripts/check_issue.py` | prevents | enforced |
| A project's verification contract names real commands | `content/skills/hive-rule-project-verification/SKILL.md` | `content/skills/flow-verify/scripts/check_contract.py` | prevents | enforced |
| Session archives move with their records | `content/skills/project-layout/references/archive.md` | `content/skills/project-layout/scripts/archive_sessions.py` | prevents | enforced |
| A project's scripts, data, documents and index follow the layout convention, its `Layout:` line included | `content/skills/project-layout/references/layout.md` | `content/skills/project-layout/scripts/check_layout.py` (LAYOUT-01 to 08); `tooling/cli/doctor_project.go` keeps checking Specs and the ignore of the scripts and secrets folders | detects | enforced |
| What project-layout moves gets the same destination on every run, and its references move with it | `content/skills/project-layout/references/layout.md` | `content/skills/project-layout/scripts/plan_layout.py` (`plan`, `apply`, `verify`, `undo`) | prevents | enforced |
| The code review question offers the host's native review | global guidance «Delivery» | `tests/pilot/session_audit.go` (`review_question_native_option`, Codex) | detects | enforced |
| A workspace repository ignores its children and re-includes them for search | `content/skills/workspace-repo/SKILL.md` | `content/skills/workspace-repo/scripts/workspace_repo.py` | prevents | enforced |
| No broad `git add`, no read of a secret's content, a question after its detail, cited IDs glossed, no bare URLs, ticket IDs not packed in prose, merged branches deleted | global guidance «Secrets» and «Questions and routes»; `content/skills/flow-close/SKILL.md` | `tests/pilot/regression.go` | detects | enforced |
| A round goes through the permitted waiting card, not only in text | global guidance «Questions and routes» | `tests/pilot/session_audit.go` (`text_round_without_card`) | detects | enforced |
| A flow starts from the user's reply to a next-step block or their typed command, never from a reply to a question or proposal | `content/skills/hive-rule-next-step-block/SKILL.md` | `tests/pilot/session_audit_routes.go` (`flow_entered_without_route`: fails on a bare reply or a card's answer, flags a longer reply for review) | detects | enforced |
| A decision round written in text keeps its full format: a `Situation:` line and the starred option | `content/skills/hive-rule-questions/SKILL.md` | `tests/pilot/session_audit_routes.go` (`decision_round_format`) | detects | enforced |
| A choice asked in prose while a waiting card is available is a round: a delivery, review or security choice, offered as alternatives or as a yes or no to confirm, fails; another choice is flagged for review | `content/skills/hive-rule-questions/SKILL.md` | `tests/pilot/session_audit_prose.go` (`prose_choice_without_card`) | detects | enforced |
| A change to a file that renders an interface gets a `hive-review-ux` child | `content/skills/flow-verify/SKILL.md` | `tests/pilot/session_audit_routes.go` (`ui_change_without_ux_review`) | detects | enforced |
| A check whose output is saved runs through a capped helper that closes its log before showing the end | `content/skills/hive-rule-checks/SKILL.md` «Check reuse» | `content/skills/hive-rule-checks/scripts/run_check.py` | prevents | enforced |
| No shell command reads, in a trap or inside the block, the file its own block redirects into | `content/skills/hive-rule-checks/references/run-check.md` | `tests/pilot/session_audit_logs.go` (`self_feeding_log`) | detects | enforced |
| A chosen ticket's assessment ends with a card, the next-step block or the block of replies, never with nothing to answer | `content/skills/hive-mode/SKILL.md` «Assess a chosen ticket» | `tests/pilot/session_audit_assessment.go` (`ticket_assessment_ending`; an open question is flagged for review) | detects | enforced |
| A tracker's states topic found missing is written before the session ends, when the session moves a ticket | `content/skills/hive-rule-trackers/SKILL.md` | `tests/pilot/session_audit_trackers.go` (`tracker_states_cache_written`) | detects | enforced |
| The final report of a verified interface change carries the Try it block | `content/skills/hive-rule-verification-floor/SKILL.md` | `tests/pilot/session_audit_trackers.go` (`try_it_block_in_report`) | detects | enforced |
| Before the first `git add` or commit, a ticket move or a question card, the session reads the leaf that governs it (`hive-rule-git-commit`, `hive-rule-ticket-state`, `hive-rule-questions`) | global guidance «Rule leaves» | `tests/pilot/session_audit.go` (`leaf_read_before_action`; `fail` when an observed action has no read before it, `review` when a withheld shell command could hold it, `not_observed` when the session has no such action; only the first action of each class is checked) | detects | enforced |
| Close deletes the merged branch before its report in the closes that defer no step; at a completion point the report comes first | `content/skills/flow-close/SKILL.md` step 5 | `tests/pilot/session_audit.go` (`close_before_branch_cleanup`, flags for review only) | detects | judgment |
| A reason states its effect, not the instruction that requires it | global guidance «Questions and routes» | `tests/pilot/session_audit.go` (`skill_named_as_reason`, flags for review only) | detects | judgment |
| Choose the smallest sufficient investigation, plan and verification | global guidance «Proportionality» | none | — | judgment |
| Ask only for a choice that is the user's: take and list the rest, ask only what criteria a–f send | `content/skills/hive-rule-questions/SKILL.md` «What to ask» | none | — | judgment |
