# Instruction resources

## Authoring

Keep each procedure and resource in one owning skill. Its `SKILL.md` states the essential procedure and links situational resources with the condition that requires reading them. A reference may link further resources needed for that situation; do not preload every reference for every task.

Use Markdown links for packaged dependencies, including scripts and assets. Within a skill, use paths relative to the containing file, such as `[UI verification](references/ui-verification.md)`. Between skills, use relative links such as `[plan handoff](../flow-plan/references/plan-handoff.md)`. Put fictional template links in fenced examples. Runtime project inputs are supplied or discovered in the task; they are not bundled dependencies.

Native agent renderers can flatten or relocate role files. In role instructions, identify a skill resource with `[plan handoff](skill:flow-plan/references/plan-handoff.md)` instead of a path relative to the source role. This is an authoring locator: the owner is `flow-plan` and the path within it is `references/plan-handoff.md`. It does not create a URI handler, invoke a tool, or preload a skill. Do not embed a developer's absolute installation path.

## Runtime resolution

The essential runtime boundaries and read conditions remain in `content/guidance/global.md`; the delegation procedure has its canonical home in the rule leaf `content/skills/hive-core/delegation/delegation.md`, read from the global guidance's «Rule leaves» index and from each flow that delegates. Resolve the named skill through the host's available catalog/loader or an explicit task path, then read the resource relative to that skill directory. Reading a reference does not invoke its owner's complete workflow; a resource that is read is read whole, under the resource rule in `content/guidance/global.md`, and `tests/content/skill_size_test.go` keeps each reference at or under 16,000 bytes (issue #204). Native discovery, parent context inheritance, and automatic reference reading are different capabilities; do not assume one establishes another.

Canonical procedural homes under `content/`:

- `skills/hive-core/ui/walkthrough.md`: publish UI captures and their iteration to the viewer (`hive view work capture`), or deliver a UI walkthrough, also one the user asks for, through the supported fallback; its sibling leaves `viewer`, `viewer-open` and `viewer-close` own the first capture of a session, the offer of the viewer at a stop and the close of a work with captures. Its consumers link to it instead of repeating publication steps; `hive view work capture --help` owns the capture command and `hive view work contract` the advanced semantic publisher interface and `hive view contract` the advanced machine-readable protocol.
- `skills/create-design/references/ui-criteria.md`: the precedence of UI sources (the project's accessibility requirement, recorded decisions, customization layer, library defaults, then the criteria's defaults), the `DESIGN.md` design record and the UI criteria with their measured checks. `flow-plan`, the design, build and UX review roles and `flow-verify` link to it instead of restating them; it is a reference of `create-design`, whose fourth entry creates or refreshes `DESIGN.md`.
- `skills/create-design/SKILL.md`: the four entries (draw, correct, review an existing design, create or refresh `DESIGN.md`), the drawing of screen examples in Figma at the width its platform sets, the business rules the frames decide, the choice, the record of the approved example and the request whose deliverable is the design; `references/review.md` owns the review entry, which runs one `hive-review-ux` child and draws nothing; `hive-mode` only recommends the stage in its next-step block. `flow-plan`, `flow-build`, `flow-verify`, `hive-review-ux` and the site's flow pages link to it or compare against the approved example it records instead of restating it; the brief of the `hive-review-ux` child that reviews its frames is owned by `references/frame-review-brief.md`, which adds to `skills/hive-core/delegation/reviewer-brief.md` (the common brief every verify loads), and the brief for a design someone else drew is the second step of `references/review.md`; the `**Design file:**` and `**Design library:**` fields it reads are defined by `skills/create-design/references/design-record.md`, and how the stage asks for them is owned by its `references/design-library.md`; a yes stays pending and the plan writes them.
- `skills/create-design/references/design-library.md`: the sibling records the stage reads, which Figma library binds the frames (the conversation, the design record or detection), the inventory written before the first drawing, the pieces reused from the bound library and from the siblings, and the question that records the design file and the design library for the plan to write in the app's `DESIGN.md`; owned by `create-design`, which reads it at step 3. `design-record.md` and `flow-plan` «UI decisions» point to it for those two fields, and `rule-cards.md` for the bound library.
- `skills/create-design/references/rule-cards.md`: the rule cards drawn beside each frame, their rows, their update when a rule is answered and their correction after the design stage; owned by `create-design`, which reads it at steps 5 and 7, and read from later flows when a card's rule changes. It links to the business rules reference for the rule itself.
- `skills/create-design/references/sections.md`: the sections that hold every node the stage creates on the canvas, which sections it creates by default (assets, moodboard, one per flow) or by the user's grouping, their names, the arrangement inside them, their placement, the rule card left beside a frame the work did not create, and their record lines; owned by `create-design`, read at its step 5 and pointed to by its rule-cards and design-library references.
- `skills/hive-mode/SKILL.md` and `playbooks/<kind>.md`: the entry router (formerly `flow-research`), whose table names every kind of work, and the playbooks of the six kinds with steps of their own (`ui-change`, `data-change`, `dependency-security`, `ticket-status`, `refactor`, `perf`), each at most 7,000 bytes; the other kinds (`feature`, `bugfix`, `hotfix`, `infra-ci`, `investigate`, `deliverable`, `pr-review-deploy`) carry what is specific to them in their table row since change `frontier-prune`, and a stage reads a kind without a file from that row. The router owns the classification into at most one kind, the handoff between kinds as a scope change and where the choice is recorded (the `Playbook` field of `proposal.md`, or a `- Playbook:` line of the checkpoint file); a playbook owns what is specific to its kind, links the rule leaves instead of restating them, and names the delivery mode only through the phrase «under the delivery mode in force» (`tests/content/leaf_skill_test.go` checks it). The router links `references/code-audit.md` for an audit, and its `hotfix` row links `flow-deploy/references/urgent-fixes.md`.
- `skills/create-issue/SKILL.md`: the procedure to create a ticket, break work down into a parent ticket with children, or validate an existing one, with its `references/ticket-types.md` (types, ready list, hint), `breakdown.md`, `trackers.md` and `existing-ticket.md`; owned by `create-issue`. The `scope-changes` leaf keeps the rule of when a ticket is proposed and `report-endings` that creating one needs the user's answer; `hive-mode`, `flow-close`, `flow-plan` and `unattended-delegation` link to the skill for how, and `hive-mode` reads its ready list to judge a chosen ticket. The site page `flows/create-issue.md` describes it without copying its rules.
- `skills/hive-core/tracker/trackers.md`: how a ticket moves, gets a comment, an attachment or a worklog, with one reference per tracker (`tracker/trackers/jira.md`, `linear.md`, `github.md`) and `tracker/trackers/time.md`, the measuring, conversion and writing of a ticket's time; `ticket-state` keeps when each write happens. Its `scripts/trackers/jira_rest.py` makes the Jira writes `acli` cannot (transitions, worklogs, attachments) with the user's token from `~/.config/hive/secrets/`, and `scripts/trackers/work_time.py` keeps time marks on hosts whose transcript `hive time` cannot read.
- `skills/create-prd/SKILL.md`: the procedure to draft a requirement document in PRD format, restructure one that arrives already written, update it with its author's answers and run its technical review against the code, with its `references/prd-format.md` (sections, rule entry, ready list, states and versions), `incoming-document.md`, `documents.md` (where the document lives, the operations of each document tool, freezing a version and the Markdown mode, whose PDF `scripts/render_pdf.py` renders) and `technical-review.md`; owned by `create-prd`. `hive-mode` steps 0 and 5 point to it for a request whose deliverable is a requirement document, and the `hive-review-prd` role reads its format reference; neither restates its steps.
- `skills/hive-core/business/business-rules.md`, `business-record` and `business-ticket`: what a business rule is and when the work adds, replaces or cancels one; the record's line format and statuses, which `flow-plan` keeps in the change folder's `proposal.md`; and the write of the rules to the work's ticket at the design approval, the build start and the close. `guidance/global.md` «Rule leaves», `create-design`, `flow-build`, `flow-verify` and `flow-close` link to them instead of restating them; `hive-review-ux` links to the first for the definition of a rule.
- `skills/comment-audit/SKILL.md`: auditing a codebase's comments, their removal in batches, the comment ratchet and the project's own comment skill, with `references/ratchet.md` (the template's contract, a port to another language and the CI wiring), `references/project-skill.md` (where the generated skill lives, what it holds and its `## Skills` registration) and the template `assets/ratchet/check-comments.mjs` with its suite, which `tests/content/comment_audit_template_test.go` runs with the parsers pinned in `tests/content/testdata/comment-audit/`. It takes each comment's category from the `no-comments` leaf instead of restating it. That leaf is the one home of the rule for a project whose lint bans comments and of the workaround marker; `defect-fix`, `suppression-review` and the site's Build page link it.
- `skills/flow-verify/SKILL.md` and the leaves `ui-verification`, `before-captures` and `a11y-audit`: verification, review, UI checks and proportional capture coverage, the captures of an existing screen before its first edit, and the accessibility audit of a UX review. Publication consumers point there for those checks instead of restating a second capture policy.
- `guidance/global.md` «Questions and routes», one invariant line that lets only the user's reply or a typed command start a flow, linking `report-endings`; the global «Rule leaves» index links the other leaves of this family: `questions` (what to ask, the decision format with its `Situation:`, `If you choose it:` and `I recommend` lines, the fixed headers of the session questions, the finding format, the context line and waiting line, the channel for each host and asking in text; its `messages/questions/unanswered.md` holds a card that comes back unanswered, read at that moment), `reasons-and-blockers` (the reasons shown to the user, stated as effects and never as what an instruction the agent follows requires, except one the user or their project wrote, and how a blocker ends), and `report-endings` (how each message ends: analysis, ticket assessment, the next-step block between flows with its compaction reply, the close footer, single actions and the same-turn continuations). The section keeps the minimal invariant of each and the link.
- `guidance/global.md` «Scope and authorization»: effects beyond this machine, critical effects, third-party requests, delegation authority and the current-project boundary, whose external-project procedure lives in the `scope-changes` leaf, which also holds a change the user asks for, or the work finds it needs, while work is open. «Preservation» keeps the close-cleanup permission and its «never» rules; the list of what that cleanup may do is the `close-cleanup` leaf. «Proportionality» keeps the routing and the large-read, bounded-lookup and risky-change definitions; the `proportionality` leaf defines a mechanical change, a small understood change and an operational task.
- `guidance/global.md` «Delivery»: the delivery modes and the effects each answer authorizes, kept whole; the leaves `delivery-ask` (the delivery and code-review questions with their recommendation) and `foreign-commits` hold the procedure. `skills/git-workflow/SKILL.md` performs those effects and reads Git and pull-request state through its `scripts/git_state.py`; the leaves `git-commit` (staging and committing), `promotion-path`, `merge-gate` and `pr-reviews` hold the commit, the environment-branch path, the pre-merge gate and the reviews after the push; its conditional `references/pr-captures.md` owns the before and after captures of the pull request body, their upload and the public-repository rule. The before captures themselves belong to `before-captures`.
- `guidance/global.md` «Language», «Secrets» and «Persistent memory», and the `project-settings` leaf: the rules that used to live in an earlier `hive-core` skill's references. Commit `5e6e8907` dissolved that skill into the global guidance and the one-skill-per-rule leaves. This change brings back a skill named `hive-core`, now only the hidden store of the leaves (see `skills/hive-core/` below), not a shared core that every flow loads: the global guidance plus the leaves still carry the shared rules.
- `skills/hive-core/principles/verify-current-state.md`: the evidence rule, verifying recalled memory and prior summaries before presenting them as current and reading the current documentation of a library, API, CLI tool or cloud service. the «Rule leaves» index of `guidance/global.md` lists it under «Principles», and the principles index (`principles/index.md`) links it with its moment, which also applies to an answer outside code work.
- `skills/hive-core/local/local-write/worktree.md`: how `HIVE.local.md` is read and written from a worktree (the main checkout's file, append-only when the host refuses editing, no copy in the worktree). Read when the session runs in a worktree, from the `guidance/global.md` «Local instructions» line at session start and from `local-write` before a write.
- `guidance/global.md` «Rule leaves»: the index of rule leaves, grouped under one `###` heading per domain of `hive-core/assets/domains.json`, one line per leaf with the observable moment that requires reading it and a `skill:` link the installer rewrites to an absolute path. Outside its home, a rule is cited only by that moment and a link, anchored to the exact section when the rule is one section of a longer file; a citation never restates the rule. It is the canonical home of the moments; each leaf is the canonical home of its rule, and flow steps repeat only the link. `tests/content/leaf_skill_test.go` checks each leaf's index line and size, that it has no frontmatter, a title and a first line starting «Read», and that `hive-core/SKILL.md` carries the hidden mark; `tests/content/index_trigger_test.go` checks that each index trigger is the leaf's first sentence; and `tests/content/leaf_links_test.go` that each leaf is linked from the flows where its rule applies.
- `skills/hive-core/`: the store of the rule leaves, a hidden skill (`disable-model-invocation: true`) whose `SKILL.md` only says that the leaves are read by path. The domain folders are the ones `assets/domains.json` lists (`messages`, `git`, `plan`, `build`, `delegation`, `ui`, `review`, `business`, `close`, `local`, `tracker` and `principles`); `references/` is not one of them, and the manager installs no other first-level folder of the skill. The `principles/` domain holds the code principles and their own index, `principles/index.md`, which tells the agent when each applies; `hive-typescript` stays a visible skill. Each leaf is `<domain>/<name>.md`, one rule each, without frontmatter, at most 4,000 bytes unless `tests/content/leaf_skill_test.go` names a larger limit for it, and its first body line after the title starts «Read…». A leaf's own references are in `<domain>/<name>/` and its scripts in `scripts/<name>/` (`scripts/checks/run_check.py`, `scripts/trackers/jira_rest.py`, `scripts/trackers/work_time.py`). The agent reads a leaf at the moment the index or a flow step names it, and never invokes it as a skill, which a host may refuse. A leaf links its siblings with relative paths such as `prose.md`, and a flow or role names it with a `skill:` locator such as `skill:hive-core/build/prose.md`. Before change `skill-leaves` each leaf was its own skill folder, and before change `hive-core-domains` the leaves sat in `references/` and the principles were hidden skills of their own; the manager retires those folders and files when it installs `hive-core`. A flow may also list, in `assets/stage.json`, the leaves it reads at a stage; the manager generates that stage packet, `<flow>/stage.md` or `<flow>/stage-<name>.md`, at install time, and the flow says at entry to read it whole and not to read its leaves again until a compaction, after which the agent reads it again; a missing or cut packet falls back to the linked leaves, and a missing heading to that leaf. The source keeps one home per leaf: the packet is a generated copy, never edited by hand, and `tests/content/stage_manifest_test.go` (`TestPacketHoldsLinesMatchManifests`) fails when the leaves a flow's `SKILL.md` says its packet holds differ from its `assets/stage.json`. A leaf read under one condition stays out of the packet: `flow-verify` no longer preloads `project-verification`, and `flow-build` no longer preloads `stored-data` or `security-boundaries`. The checkpoint names the packets a flow read in its `- Rules:` line, so a session that resumes after a compaction reads them again (`checkpoint` leaf). Each of seven flows ends its skill with one `Reply:` line that links the report row ending it, and `tests/content/reply_line_test.go` requires it. Families: questions and reports (`questions`, `reasons-and-blockers`, `report-endings`), delivery and Git (`delivery-ask`, `foreign-commits`, `git-commit`, `promotion-path`, `merge-gate`, `pr-reviews`), plan and build (`change-folder`, `proportionality`, `checkpoint`, `checks`, `reuse-before-writing`, `tests`, `tests-no-runner`, `defect-fix`, `stored-data`, `security-boundaries`, `external-facts`, `external-effects`, `scope-changes`, `project-settings`, `prose`), delegation (`delegation`, `child-brief`, `reviewer-brief`), UI and design (`design-stage`, `ui-plan`, `ui-example`, `ui-verification`, `before-captures`, `a11y-audit`, `walkthrough`, `viewer`, `viewer-open`, `viewer-close`), review (`no-comments`, `blast-radius`, `suppression-review`, `review-evidence`, `review-findings`, `interrogate`), verification contract (`project-verification`, `project-verification-format`), business rules (`business-rules`, `business-record`, `business-ticket`), close (`close-modes`, `close-hygiene`, `close-cleanup`, `close-report`), local state and accounts (`local-record`, `local-write`, `local-script`, `test-accounts`, `ticket-state`) and unattended work (`unattended-effects`).
- Per-host hiding of the leaves: Claude Code reads `disable-model-invocation` from the frontmatter of `hive-core/SKILL.md`; Codex reads its invocation policy from `agents/openai.yaml`, which the manager generates for each hidden skill (`allow_implicit_invocation: false`) and records and removes like any managed skill file (`tooling/management/hidden_skills.go`); Antigravity lists the `hive-core` skill in its catalog with its one-line description, because a relative link from a linked flow does not resolve there (design of change `skill-leaves`, decision D6 A). A mechanism for the other hosts is not verified here; their catalogues may list `hive-core`.
- `skills/hive-core/delegation/delegation.md`: before deciding whether a child does a piece of work, in any flow; `child-brief` before briefing a child or waiting for it, and `reviewer-brief` before briefing a reviewing or verifying child. `skills/flow-build/SKILL.md`: when the turn may end during a build; the leaves `checkpoint` and `checks` hold the `tasks.md` checkpoint, written with `skills/flow-plan/scripts/tasks.py`, and the checks the build repeats; the latter's `build/checks/run-check.md` and `scripts/checks/run_check.py` run a check whose output is saved, under a size cap and with the log closed before its end is shown.
- `skills/hive-core/build/defect-fix.md`: before the first edit of a defect fix, and when a plan weighs the approaches to one; read from `flow-build` step 2, `reuse-before-writing` and `tests`, `flow-verify` step 3, `flow-deploy/references/urgent-fixes.md` and `hive-review-code`. `skills/hive-core/build/stored-data.md`: before a change to stored data; `skills/hive-core/build/security-boundaries.md`: when a change touches a boundary its first sentence lists (eight topics; the global index line repeats that sentence). `flow-build` and `flow-plan` link both leaves with their conditions instead of preloading them in a stage packet, and `hive-review-security` step 1 and the `infra-ci` row of `hive-mode` point to the leaf instead of listing the topics.
- `skills/hive-core/plan/change-folder.md` and `skills/flow-plan/assets/change-folder/` templates: creating or continuing a change record; `skills/flow-close/SKILL.md` closes it and records its pending archive, with `close-modes` and `close-hygiene`; `skills/flow-close/references/batch-archive.md` archives the pending folders on a hygiene request. `skills/flow-plan/references/plan-handoff.md`: the plan summary and the build route. `skills/flow-build/references/spec-delta.md`: a change built without a plan that alters a behavior an existing spec describes, which writes its change folder without a plan review. `skills/flow-build/references/first-view.md`: a build that adds or rearranges a screen, from its step 4: task order around the first-view question and the user's review of the first rendered view.
- `skills/hive-core/close/close-report.md`: the one final report of a work item with a row per variant, the findings at the close, and the severity, grouping and recording of findings; `report-endings` holds how every message ends (the next-step block, the close footer and the same-turn continuations), and `skills/flow-close/references/support-material.md` the inventory of the work's own support material, what would be deleted, and the actions the user answers at close.
- `skills/project-layout/SKILL.md`: checking, cleaning up, pruning, reorganizing and migrating the layout of existing documentation, support material, scripts and data in a repository, a declared workspace or a monorepo, with or without Hive, through an inventory, a manifest, one confirmation and an archive tag per repository; its `references/inventory.md` sets the homes it lists, the protected items, the promotion of a conclusion and the pruning of a living document, and its `references/archive.md` with `scripts/archive_sessions.py` the archive of closed sessions. flow-close, artifact layout and harness-audit link to it instead of restating it.
- `skills/project-layout/references/layout.md`: before writing or cleaning support artifacts, scripts or an operation's data; basic placement remains global.
- `skills/project-layout/references/script-lifecycle.md`: when a work writes, places or promotes a script.
- `skills/workspace-repo/SKILL.md`: turning a declared workspace into a coordinator repository on the user's request.
- `skills/flow-build/references/e2e-coordination.md`: explicit bounded coordination through existing flows, the local-owner exception, exact-head verification, replacement and integration. Global guidance keeps only its activation/authority invariant; research, plan, build, verify, close, unattended-delegation, git-workflow and the owner role link to it when the mode applies.
- `skills/flow-plan/references/execution-record.md`: the multiunit record, CLI and versioned state contract owned by `flow-plan/scripts/tasks.py`, read by E2E coordination and Build for that checkpoint. The record is local ignored state, not a second versioned plan or proof of execution.
- `skills/hive-core/build/project-verification.md` and `project-verification-format`: the project-owned Launch/Doctor/Drive/Evidence/Cleanup index and feature map; Plan and Build read the first when verification capabilities change, Verify before the real-path walk of E2E work or of a change to a verification command (the `flow-verify` stage packet does not hold it), and the second when writing or checking the JSON. `skills/flow-verify/assets/project-verification.json` and `scripts/check_contract.py` initialize without overwrite and validate structure without running commands.
- `skills/hive-core/review/suppression-review.md`: inventory and justified dispositions for added or widened suppressions, skipped checks and configuration exclusions. Verify supplies it to every selected review mechanism; `hive-review-code` links to the same procedure. Missing native coverage remains unverified.


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

The caller supplies relevant project conventions, the project's `## Skills` lines whose trigger applies to the task, and task decisions. If resolution is ambiguous, inaccessible, or unavailable, report the dependency and pause only work requiring it; continue independent work. Distinguish a missing project guide from a required packaged reference that could not be read. Neither an installed file nor a listed skill proves a subagent read it.

## Release validation

The management package checks instruction links in the global core, roles and skills when preparing a release and when validating a saved plan before applying it. It resolves links against the release payload, not files left elsewhere in the checkout. Removing an owner or resource while a consumer still requires it fails validation before managed writes, including a saved plan whose hashes were recalculated.

The supported authoring subset is ordinary inline Markdown links and reference definitions with simple destinations (optionally angle-bracketed or with a double-quoted inline title). Use uncomplicated paths without parentheses. The check covers the global core, role Markdown, skill entrypoints, and Markdown under their `references/` directories and under the domain folders that the skill's `assets/domains.json` lists. It validates relative and `skill:owner/path` targets, checks logical owners, verifies ATX heading anchors (`#fragment`) on local, relative and `skill:` Markdown links, and rejects unsupported local schemes and personal installation paths. A directory link must end in `/` and contain packaged files.

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

These comparisons use baseline `2bb96478` (the baseline of AC12 of change `skill-leaves`): pre-existing paths are measured on
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

Measurement for `skill-leaves`, 2026-11-19, working tree of `feat/skill-leaves` against
`2bb96478` (rerun `go test ./tests/content -run Footprint -count=1 -v`). The
`plan-build-required` scenario now lists every leaf its read conditions select:
`delivery-ask`, `decision-format`, `question-channel`, `report-endings`,
`git-commit`, `checkpoint`, `tests`, `reuse-before-writing`, `delegation`,
`child-brief`, `reviewer-brief`, `checks`, `review-findings`, `merge-gate`,
`close-modes`, `close-hygiene`, `close-report`, and the settings, ticket, business-rule,
verification, change-folder and review-evidence leaves of the earlier scenario.

| Scope | Baseline bytes | Candidate bytes | Change |
| --- | ---: | ---: | ---: |
| Global (`global-only`) | 46,912 (at `d5164f71`) | 30,795 | -16,117 |
| Legacy plan/build subset (`plan-build-cycle`) | 107,021 (at `d5164f71`) | 94,793 | -12,228 |
| Required plan/build, pre-existing paths, same list | 159,011 | 104,556 | -54,455 |
| Required plan/build, new sources (leaves) | 0 | 82,412 | +82,412 |
| Required plan/build union against the baseline's own list | 184,489 | 186,968 | +2,479 (+1.34%) |
| The same union without `git-commit` and `merge-gate` | 184,489 | 180,545 | -3,944 (-2.14%) |
| E2E, pre-existing paths | 201,492 | 140,077 | -61,415 |
| E2E, new sources | 0 | 88,770 | +88,770 |
| Executable helpers | 105,283 | 105,283 | 0 |

The baseline's own list is the `plan-build-required` list at `2bb96478`, which read the
flow skills and the references that the rules have since left. With every selected leaf
the union is 2,479 B above that baseline, not below it. The two leaves that explain it,
`git-commit` and `merge-gate` (6,423 B together), select moments the baseline list never
read, since it held no `git-workflow`; without them the union is 3,944 B below. The
test fails when the union without those two leaves is not below the baseline, and
reports the full union beside it. This is a declared reading union, not observed
reading: a session that does not reach a moment does not read its leaf. The 82,412 B of
leaves are reported apart so that a new source is never counted as a saving.

Measurement for `core-cleanup-batches`, 2026-11-22, uncommitted working tree on `8aa361cb`
(rerun `go test ./tests/content/ -run TestRequiredProcedureReadingFootprint -count=1 -v`):
the `plan-build-required` union with every selected leaf is 195,800 B (+11,311 B, +6.13%
against the 184,489 B of the baseline's own list, of which 87,873 B are new sources), and
the scenario counts 25 file reads, 39 without packets. The scenario lists
`project-verification` and `project-verification-format` under its E2E route, since the
`flow-verify` packet no longer preloads them. This is a declared reading union, not
observed reading.

Native discovery metadata, rendering, plugin wrappers, repeated reads across
requests or agents, project inputs, history and host policy are excluded.
Optional loader injection can change delivery timing and repetition without
adding a unique canonical source. Native observations must establish reading,
model identity and outcomes independently; one bounded run does not qualify a
model or prove cross-host reliability.

## Vendored upstream material

`content/skills/create-report/references/diagram-grammar.md` and `diagram-svg-construction.md` (one numbering, split at 16,000 bytes) adapt [cathrynlavery/diagram-design](https://github.com/cathrynlavery/diagram-design) v2.3.2 (MIT). To pull upstream updates, re-read that repository's `skills/diagram-design/SKILL.md` §4–§9 and `references/type-*.md`, then re-map them onto the two references. Deliberately not adopted: Mermaid/draw.io import pipelines, animation, PNG/SVG export, brand onboarding, and long-tail types (medallion, radar, venn, DP matrices, loop, org chart).

`content/skills/test-audit/SKILL.md` adapts the `test-audit` skill of [openclaw/openclaw](https://github.com/openclaw/openclaw/tree/main/.agents/skills/test-audit) at `3195fba90643` (MIT; notice in `content/skills/test-audit/references/license-openclaw.md`). It keeps the R/F/C/D ledger, the retention bar, keepers before deletions and the independent preservation review, takes its authoring criterion from flow-build's Tests section instead of restating it, and adds targeted mutations and suite wall time to the report. To pull upstream updates, re-read that directory's `SKILL.md`, `references/audit.md` and `CAMPAIGN.md`. Deliberately not adopted: OpenClaw's own scripts (`check-changed.mjs`, the `deadcode:*` gates), its `openclaw-testing` proof policy, and per-subsystem `AGENTS.md` test rules.

`content/skills/hive-mode/references/code-audit.md` is original text. Its method (categories audited in parallel, findings verified at their cited location, ranking by impact and effort, direction findings apart, a record of rejected findings) is inspired by [shadcn/improve](https://github.com/shadcn/improve) (MIT). No passage was adapted, so no license copy is included; the site's Inspiration page (`site/src/content/docs/inspiration.md`) carries the credit. Deliberately not adopted: improve's `plans/` folder, its executor that commits in a worktree, its effort levels, and its rule that the auditor never implements. To pull upstream updates, re-read its `skills/improve/SKILL.md` and `references/audit-playbook.md` and compare the method, not the wording.

`content/skills/create-design/references/ui-criteria.md` and its sibling references are original text. Its methodology (screen modes, product and design records, accessibility minimums, refinement passes) is inspired by [Impeccable](https://github.com/pbakaus/impeccable) by Paul Bakaus (Apache-2.0), and the design record follows the [DESIGN.md format](https://github.com/google-labs-code/design.md) of Google Labs (Apache-2.0). No passage of either project was adapted, so the redistribution terms of their licenses do not apply and no license copy is included. The site's Inspiration page carries the credit. To pull upstream updates, re-read both repositories and compare the methodology, not the wording.
