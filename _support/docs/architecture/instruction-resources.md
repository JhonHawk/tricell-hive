# Instruction resources

## Authoring

Keep each procedure and resource in one owning skill. Its `SKILL.md` states the essential procedure and links situational resources with the condition that requires reading them. A reference may link further resources needed for that situation; do not preload every reference for every task.

Use Markdown links for packaged dependencies, including scripts and assets. Within a skill, use paths relative to the containing file, such as `[UI verification](references/ui-verification.md)`. Between skills, use relative links such as `[plan handoff](../flow-plan/references/plan-handoff.md)`. Put fictional template links in fenced examples. Runtime project inputs are supplied or discovered in the task; they are not bundled dependencies.

Native agent renderers can flatten or relocate role files. In role instructions, identify a skill resource with `[plan handoff](skill:flow-plan/references/plan-handoff.md)` instead of a path relative to the source role. This is an authoring locator: the owner is `flow-plan` and the path within it is `references/plan-handoff.md`. It does not create a URI handler, invoke a tool, or preload a skill. Do not embed a developer's absolute installation path.

## Runtime resolution

The essential runtime boundaries and read conditions remain in `content/guidance/global.md`; the delegation procedure has its canonical home in `content/skills/flow-build/references/delegation.md`, read from the global guidance's «Proportionality» and from each flow that delegates. Resolve the named skill through the host's available catalog/loader or an explicit task path, then read the resource relative to that skill directory. Reading a reference does not invoke its owner's complete workflow; a resource that is read is read whole, under the resource rule in `content/guidance/global.md`, and `tests/content/skill_size_test.go` keeps each reference at or under 16,000 bytes (issue #204). Native discovery, parent context inheritance, and automatic reference reading are different capabilities; do not assume one establishes another.

Canonical procedural homes under `content/`:

- `skills/flow-verify/references/walkthrough.md`: publish UI captures and their iteration to the viewer (`hive view work capture`), or deliver a UI walkthrough, also one the user asks for, through the supported fallback; the close of a work with captures follows its `viewer-publication.md`. Its consumers link to it instead of repeating publication steps; `hive view work capture --help` owns the capture command and `hive view work contract` the advanced semantic publisher interface and `hive view contract` the advanced machine-readable protocol.
- `skills/hive-ui-criteria/SKILL.md`: the precedence of UI sources (the project's accessibility requirement, recorded decisions, customization layer, library defaults, then the skill's defaults), the `DESIGN.md` design record and the UI criteria with their measured checks. `flow-plan`, the design, build and UX review roles and `flow-verify` link to it instead of restating them.
- `skills/flow-figma-design/SKILL.md`: the drawing of screen examples in Figma, the business rules the frames decide, the choice, the record of the approved example and the request whose deliverable is the design; `hive-mode` only recommends the stage in its next-step block. `flow-plan`, `flow-build`, `flow-verify`, `hive-review-ux` and the site's flow pages link to it or compare against the approved example it records instead of restating it; the brief of the `hive-review-ux` child that reviews its frames is owned by `skills/flow-build/references/delegation.md` «Reviewer brief»; the `**Design file:**` and `**Design library:**` fields it reads are defined by `skills/hive-ui-criteria/references/design-record.md`, and how the stage asks for, writes and delivers them is owned by its `references/design-library.md`.
- `skills/flow-figma-design/references/design-library.md`: the sibling records the stage reads, which Figma library binds the frames (the conversation, the design record or detection), the inventory written before the first drawing, the pieces reused from the bound library and from the siblings, and the question, the write and the delivery that record the design file and the design library in the app's `DESIGN.md`; owned by `flow-figma-design`, which reads it at step 3. `design-record.md` and `flow-plan` «UI decisions» point to it for those two fields, and `rule-cards.md` for the bound library.
- `skills/flow-figma-design/references/rule-cards.md`: the rule cards drawn beside each frame, their rows, their update when a rule is answered and their correction after the design stage; owned by `flow-figma-design`, which reads it at steps 5 and 7, and read from later flows when a card's rule changes. It links to the business rules reference for the rule itself.
- `skills/flow-figma-design/references/sections.md`: the sections that hold every node the stage creates on the canvas, which sections it creates by default (assets, moodboard, one per flow) or by the user's grouping, their names, the arrangement inside them, their placement, the rule card left beside a frame the work did not create, and their record lines; owned by `flow-figma-design`, read at its step 5 and pointed to by its rule-cards and design-library references.
- `skills/hive-mode/SKILL.md` and `playbooks/<kind>.md`: the entry router (formerly `flow-research`) and its 13 playbooks (`feature`, `ui-change`, `bugfix`, `hotfix`, `data-change`, `dependency-security`, `infra-ci`, `investigate`, `ticket-status`, `deliverable`, `pr-review-deploy`, `refactor`, `perf`), each at most 7,000 bytes, with `references/code-audit.md` linked from `investigate`. The router owns the classification into at most one playbook, the handoff between playbooks as a scope change and where the choice is recorded (the `Playbook` field of `proposal.md`, or a `- Playbook:` line of the checkpoint file); a playbook owns the order of the stages and what is specific to its kind, links the rule leaves instead of restating them, and names the delivery mode only through the phrase «under the delivery mode in force» (`tests/content/leaf_skill_test.go` checks both). `hotfix` links `flow-deploy/references/urgent-fixes.md`.
- `skills/create-issue/SKILL.md`: the procedure to create a ticket, break work down into a parent ticket with children, or validate an existing one, with its `references/ticket-types.md` (types, ready list, hint), `breakdown.md`, `trackers.md` and `existing-ticket.md`; owned by `create-issue`. `guidance/global.md` «Scope changes» keeps the rule of when a ticket is proposed and «Questions and routes» that creating one needs the user's answer; `hive-mode`, `flow-close`, `flow-plan` and `unattended-delegation` link to the skill for how, and `hive-mode` reads its ready list to judge a chosen ticket. The site page `flows/create-issue.md` describes it without copying its rules.
- `skills/create-prd/SKILL.md`: the procedure to draft a requirement document in PRD format, restructure one that arrives already written, update it with its author's answers and run its technical review against the code, with its `references/prd-format.md` (sections, rule entry, ready list, states and versions), `incoming-document.md`, `documents.md` (where the document lives, the operations of each document tool, freezing a version and the Markdown mode, whose PDF `scripts/render_pdf.py` renders) and `technical-review.md`; owned by `create-prd`. `hive-mode` steps 0 and 5 point to it for a request whose deliverable is a requirement document, and the `hive-review-prd` role reads its format reference; neither restates its steps.
- `skills/flow-plan/references/business-rules.md`: what a business rule is, the record's line format and statuses, where the record lives, and the write of the rules to the work's ticket at the design approval, the build start and the close; owned by `flow-plan`, which keeps the record in the change folder's `proposal.md`. `guidance/global.md` «Scope and authorization», `flow-figma-design`, `flow-build`, `flow-verify` and `flow-close` link to it instead of restating it; `hive-review-ux` links to it for the definition of a rule.
- `skills/flow-verify/SKILL.md` and its conditional `references/ui-verification.md`: verification, review, UI checks and proportional capture coverage. Publication consumers point here for those checks instead of restating a second capture policy.
- `guidance/global.md` «Questions and routes»: the decision format (`Situation:`, `If you choose it:` and `I recommend` lines) and the finding format, the reasons shown to the user, stated as effects and never as what an instruction the agent follows requires, except one the user or their project wrote, the context line and waiting line of a decision, result, blocker or closing message, and the next-step block between flows: which flow commands it offers, the single actions asked directly instead, and the compaction reply at plan → build and after a close.
- `guidance/global.md` «Scope and authorization»: effects beyond this machine, critical effects, third-party requests and delegation authority; «Scope changes»: a change the user asks for, or the work finds it needs, while work is open.
- `guidance/global.md` «Delivery»: the delivery mode and the code-review question, their recommendation and the effects each answer authorizes; `skills/git-workflow/SKILL.md` performs those effects, including the reviews after the push, and reads Git and pull-request state through its `scripts/git_state.py`; its conditional `references/pr-captures.md` owns the before and after captures of the pull request body, their upload and the public-repository rule. The before captures themselves belong to `skills/flow-verify/references/ui-verification.md`.
- `guidance/global.md` «Project settings», «Language», «Secrets», «Evidence» and «Persistent memory»: the rules that used to live in `hive-core` references (dissolved into the global guidance in `5e6e8907`).
- `skills/flow-build/references/delegation.md`: before deciding whether a child does a piece of work, in any flow, briefing it or waiting for it; `skills/flow-build/SKILL.md`: when the turn may end during a build, and the `tasks.md` checkpoint, written with `skills/flow-plan/scripts/tasks.py`.
- `skills/flow-build/references/defect-fix.md`: before the first edit of a defect fix, and when a plan weighs the approaches to one; read from `flow-build` step 2, «Reuse before writing» and «Tests», `flow-verify` step 3, `flow-deploy/references/urgent-fixes.md` and `hive-review-code`. `skills/flow-build/references/stored-data.md`: before a change to stored data; `skills/flow-build/references/security-boundaries.md`: when a change touches untrusted input, permissions, process execution or data exposure.
- `skills/flow-plan/SKILL.md` «Change folder» and its `assets/change-folder/` templates: creating or continuing a change record; `skills/flow-close/SKILL.md` closes and archives it. `skills/flow-plan/references/plan-handoff.md`: the plan summary and the build route.
- `skills/flow-close/SKILL.md` «Final report» and «Findings»: the report of a work item and the severity, grouping and recording of findings; its `references/support-material.md`: the inventory of the work's own support material, what would be deleted, and the actions the user answers at close.
- `skills/project-layout/SKILL.md`: checking, cleaning up, pruning, reorganizing and migrating the layout of existing documentation, support material, scripts and data in a repository, a declared workspace or a monorepo, with or without Hive, through an inventory, a manifest, one confirmation and an archive tag per repository; its `references/inventory.md` sets the homes it lists, the protected items, the promotion of a conclusion and the pruning of a living document, and its `references/archive.md` with `scripts/archive_sessions.py` the archive of closed sessions. flow-close, artifact layout and harness-audit link to it instead of restating it.
- `skills/project-layout/references/layout.md`: before writing or cleaning support artifacts, scripts or an operation's data; basic placement remains global.
- `skills/project-layout/references/script-lifecycle.md`: when a work writes, places or promotes a script.
- `skills/workspace-repo/SKILL.md`: turning a declared workspace into a coordinator repository on the user's request.
- `skills/flow-build/references/e2e-coordination.md`: explicit bounded coordination through existing flows, the local-owner exception, exact-head verification, replacement and integration. Global guidance keeps only its activation/authority invariant; research, plan, build, verify, close, unattended-delegation, git-workflow and the owner role link to it when the mode applies.
- `skills/flow-plan/references/execution-record.md`: the multiunit record, CLI and versioned state contract owned by `flow-plan/scripts/tasks.py`, read by E2E coordination and Build for that checkpoint. The record is local ignored state, not a second versioned plan or proof of execution.
- `skills/flow-verify/references/project-verification.md`: the project-owned Launch/Doctor/Drive/Evidence/Cleanup index and feature map; Plan and Build read it when verification capabilities change, Verify before its real-path walk. Its `assets/project-verification.json` and `scripts/check_contract.py` initialize without overwrite and validate structure without running commands.
- `skills/flow-verify/references/suppression-review.md`: inventory and justified dispositions for added or widened suppressions, skipped checks and configuration exclusions. Verify supplies it to every selected review mechanism; `hive-review-code` links to the same procedure. Missing native coverage remains unverified.


Activity and role consumers link directly to required resources. The global core
does not delegate authorization, preservation or secrets to a selector. The
[optional OpenCode loader](deployment-manager.md#optional-opencode-context-loader)
only supplies two fixed-role resources; it does not replace these read conditions.

Those two references retain their single editorial home under their skills. The
manager generates `index.js` and metadata schema 2 at
`<OpenCodeHome>/hive/context-loader/{index.js,context-loader.json}`, outside
`plugins/` and without `package.json`, from the same
validated release payload. Requests use those embedded snapshots without reading
reference files or metadata. Regeneration and plugin reload refresh the supplied
text; editing an installed Markdown reference alone does not. Derived copies and
digests record exact canonical UTF-8 provenance, not cryptographic authentication.

Deployed files carry installed paths. The installer rewrites every inline `[label](skill:owner/path)` link outside code fences, in role files for every host (a Codex role is rewritten before its TOML string is encoded) and in the global guidance block, to `<skills-dir>/<owner>/<path>` wrapped in `<...>`. At user scope `<skills-dir>` is the absolute `<home>/.agents/skills`, the shared store every host installs into (`~/.claude/skills/<skill>` links to it). At project scope it is the host's project skills directory relative to the project root, `.claude/skills` for Claude and `.agents/skills` for Codex, so versioned project files carry no personal path. The limitation: a relative path resolves from the project root, so a session opened in a subdirectory, or a worktree that lacks the project's versioned skills, does not find it directly. Reference-style definitions (`[x]: skill:...`) are not rewritten; none exists in `content/`. Source files keep `skill:` locators, and a file shared by several hosts must render to identical bytes or the plan fails before any write. The `skill:` resolution rule above still applies to a locator that reaches a model unrewritten, such as one inside a code fence.

The caller supplies relevant project conventions and task decisions. If resolution is ambiguous, inaccessible, or unavailable, report the dependency and pause only work requiring it; continue independent work. Distinguish a missing project guide from a required packaged reference that could not be read. Neither an installed file nor a listed skill proves a subagent read it.

## Release validation

The management package checks instruction links in the global core, roles and skills when preparing a release and when validating a saved plan before applying it. It resolves links against the release payload, not files left elsewhere in the checkout. Removing an owner or resource while a consumer still requires it fails validation before managed writes, including a saved plan whose hashes were recalculated.

The supported authoring subset is ordinary inline Markdown links and reference definitions with simple destinations (optionally angle-bracketed or with a double-quoted inline title). Use uncomplicated paths without parentheses. The check covers the global core, role Markdown, skill entrypoints, and Markdown under their `references/` directories. It validates relative and `skill:owner/path` targets, checks logical owners, verifies ATX heading anchors (`#fragment`) on local, relative and `skill:` Markdown links, and rejects unsupported local schemes and personal installation paths. A directory link must end in `/` and contain packaged files.

Fenced examples, asset contents, arbitrary prose/code dependencies, and remote URL availability are outside this check. HTTP(S) and mail links are external. This is a packaging gate, not a full Markdown parser or evidence of runtime reading, permissions, tool availability, or behavioral compliance.

## Refactoring and delivery

Before renaming or removing a skill/resource, find its consumers with `rg`, update their links and read conditions, and run `go test ./...` and `go vet ./...`. Test a missing resource and saved-plan rejection when changing validation. Do not duplicate a resource merely to avoid a cross-skill dependency.

Current delivery selects hosts and scope and ships the complete skill catalogue. No per-skill dependency resolver is needed. If partial skill selection is introduced later, define and test dependency closure and removal behavior before enabling it; never silently omit a required resource. Deployment remains explicitly authorized and uses the existing manager.

## Declared reading footprint

`tests/content/reading_footprint_test.go` reports source unions in UTF-8 bytes and
Unicode characters, counting a file once per scenario. Each test prints its
predicate and complete file list. These are declared reading requirements, not
observed requests, cumulative session cost, tokens or proof that a host read them.

`TestGuidanceReadingFootprint` retains the identifiable legacy scenarios:
`global-only`, `simple-query`, `question` and `plan-build-cycle`, with baseline
`d5164f71`. The last is an entrypoint/role subset, not a complete count of its
conditional references. Do not describe that subset as the full build cost.

`TestRequiredProcedureReadingFootprint` adds explicit required-procedure scenarios:

- `plan-build-required`: retained backend plan, delegated implementation and
  independent reviews, project verification and a hold/pause close. It includes
  the selected procedural references. No UI, external service, migration,
  defect correction or unattended authority is assumed.
- `e2e`: the same backend work with an explicit multiunit grant, local owners,
  correction, replacement, independent verification and authorized Git
  integration. It adds the coordination, execution-record, defect-fix and Git
  procedures and owner role. UI, data migration, deployment and unattended
  authority remain outside this scenario.
- `executable-resources`: the task/plan, verification-contract, Git and cleanup
  helpers, measured separately. Executing a helper does not imply the model
  reads its source; its bytes never enter the mandatory prose union by inference.

These comparisons use baseline `15b2e402`: pre-existing paths are measured on
both sides with the same list, and new sources are reported separately. Prose,
references and executable code have separate totals. Moving text into a required
reference is not a reduction in the required union. Missing required files fail;
there is no aggregate byte quota. `skill_size_test.go` retains the 16,000-byte
limit for each skill entrypoint and each reference.

The historical 2026-10-23 snapshot at `2d89940e` reported 33,710 B for global and
229,774 B for its then-selected plan/build union. Its source list differs from
these scenarios, so it is not a comparable savings baseline.

Measurement for `bounded-e2e-coordination`, 2026-11-17, one source snapshot
against `15b2e402` (rerun `go test ./tests/content -run Footprint -count=1 -v`):

| Scope | Baseline bytes | Candidate bytes | Change |
| --- | ---: | ---: | ---: |
| Global | 52,219 | 52,673 | +454 |
| Legacy plan/build subset, same paths | 130,562 | 132,415 | +1,853 |
| Required plan/build, pre-existing paths | 168,348 | 171,237 | +2,889 |
| Required plan/build, new sources | 0 | 7,407 | +7,407 |
| E2E, pre-existing paths | 190,649 | 194,322 | +3,673 |
| E2E, new sources | 0 | 29,866 | +29,866 |
| Executable helpers, pre-existing paths | 67,544 | 96,964 | +29,420 |
| Executable helpers, new sources | 0 | 8,319 | +8,319 |

The candidate E2E union is 224,188 B, including 65,476 B of references.
The complete distributed Markdown inventory (all `content/**/*.md`, not a reading
scenario) grew from 558,726 to 592,573 B; distributed Python grew from 194,684 to
232,423 B. Those inventories use Git blobs at the baseline and filesystem bytes
for the candidate, excluding caches. This adds capabilities and reading cost;
it is not a net prose reduction. Excluding scripts does not remove the 33,847 B
Markdown increase. Each entrypoint and reference remains under 16,000 B.

Native discovery metadata, rendering, plugin wrappers, repeated reads across
requests or agents, project inputs, history and host policy are excluded.
Optional loader injection can change delivery timing and repetition without
adding a unique canonical source. Native observations must establish reading,
model identity and outcomes independently; one bounded run does not qualify a
model or prove cross-host reliability.

## Vendored upstream material

`content/skills/flow-report/references/diagram-grammar.md` and `diagram-svg-construction.md` (one numbering, split at 16,000 bytes) adapt [cathrynlavery/diagram-design](https://github.com/cathrynlavery/diagram-design) v2.3.2 (MIT). To pull upstream updates, re-read that repository's `skills/diagram-design/SKILL.md` §4–§9 and `references/type-*.md`, then re-map them onto the two references. Deliberately not adopted: Mermaid/draw.io import pipelines, animation, PNG/SVG export, brand onboarding, and long-tail types (medallion, radar, venn, DP matrices, loop, org chart).

`content/skills/test-audit/SKILL.md` adapts the `test-audit` skill of [openclaw/openclaw](https://github.com/openclaw/openclaw/tree/main/.agents/skills/test-audit) at `3195fba90643` (MIT; notice in `content/skills/test-audit/references/license-openclaw.md`). It keeps the R/F/C/D ledger, the retention bar, keepers before deletions and the independent preservation review, takes its authoring criterion from flow-build's Tests section instead of restating it, and adds targeted mutations and suite wall time to the report. To pull upstream updates, re-read that directory's `SKILL.md`, `references/audit.md` and `CAMPAIGN.md`. Deliberately not adopted: OpenClaw's own scripts (`check-changed.mjs`, the `deadcode:*` gates), its `openclaw-testing` proof policy, and per-subsystem `AGENTS.md` test rules.

`content/skills/hive-mode/references/code-audit.md` is original text. Its method (categories audited in parallel, findings verified at their cited location, ranking by impact and effort, direction findings apart, a record of rejected findings) is inspired by [shadcn/improve](https://github.com/shadcn/improve) (MIT). No passage was adapted, so no license copy is included; `references/notice.md` carries the credit and is deliberately not linked from `SKILL.md`. Deliberately not adopted: improve's `plans/` folder, its executor that commits in a worktree, its effort levels, and its rule that the auditor never implements. To pull upstream updates, re-read its `skills/improve/SKILL.md` and `references/audit-playbook.md` and compare the method, not the wording.

`content/skills/hive-ui-criteria/` is original text. Its methodology (screen modes, product and design records, accessibility minimums, refinement passes) is inspired by [Impeccable](https://github.com/pbakaus/impeccable) by Paul Bakaus (Apache-2.0), and the design record follows the [DESIGN.md format](https://github.com/google-labs-code/design.md) of Google Labs (Apache-2.0). No passage of either project was adapted, so the redistribution terms of their licenses do not apply and no license copy is included. `references/notice.md` carries the credit and is deliberately not linked from `SKILL.md`. To pull upstream updates, re-read both repositories and compare the methodology, not the wording.
