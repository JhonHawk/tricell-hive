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
| A change folder covers its criteria and its business rules follow their grammar | `content/skills/flow-plan/SKILL.md` | `content/skills/flow-plan/scripts/check_change.py`, `content/skills/flow-plan/scripts/business_rules.py` | prevents | enforced |
| A task's record and the build's checkpoint keep their format | `content/skills/flow-build/SKILL.md` | `content/skills/flow-plan/scripts/tasks.py` | prevents | enforced |
| Deltas merge into the specs before a change folder is archived | `content/skills/flow-close/SKILL.md` step 4 | `content/skills/flow-close/scripts/apply_deltas.py` | prevents | enforced |
| A PRD keeps its sections and rule entries | `content/skills/create-prd/SKILL.md` | `content/skills/create-prd/scripts/check_prd.py` | prevents | enforced |
| A ticket keeps the sections of its type | `content/skills/create-issue/SKILL.md` | `content/skills/create-issue/scripts/check_issue.py` | prevents | enforced |
| A project's verification contract names real commands | `content/skills/flow-verify/references/project-verification.md` | `content/skills/flow-verify/scripts/check_contract.py` | prevents | enforced |
| Session archives move with their records | `content/skills/docs-cleanup/references/archive.md` | `content/skills/docs-cleanup/scripts/archive_sessions.py` | prevents | enforced |
| No broad `git add`, no read of a secret's content, a question after its detail, cited IDs glossed, no bare URLs, ticket IDs not packed in prose, merged branches deleted | global guidance «Secrets» and «Questions and routes»; `content/skills/flow-close/SKILL.md` | `tests/pilot/regression.go` | detects | enforced |
| A round goes through the permitted waiting card, not only in text | global guidance «Questions and routes» | `tests/pilot/session_audit.go` (`text_round_without_card`) | detects | enforced |
| Close deletes the merged branch before its report in the closes that defer no step; at a completion point the report comes first | `content/skills/flow-close/SKILL.md` step 5 | `tests/pilot/session_audit.go` (`close_before_branch_cleanup`, flags for review only) | detects | judgment |
| A reason states its effect, not the instruction that requires it | global guidance «Questions and routes» | `tests/pilot/session_audit.go` (`skill_named_as_reason`, flags for review only) | detects | judgment |
| Choose the smallest sufficient investigation, plan and verification | global guidance «Proportionality» | none | — | judgment |
| Ask only for a choice that is the user's | global guidance «Questions and routes» | none | — | judgment |
