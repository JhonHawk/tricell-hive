# Instruction resources

## Authoring

Keep each procedure and resource in one owning skill. Its `SKILL.md` states the essential procedure and links situational resources with the condition that requires reading them. A reference may link further resources needed for that situation; do not preload every reference for every task.

Use Markdown links for packaged dependencies, including scripts and assets. Within a skill, use paths relative to the containing file, such as `[UI verification](references/ui-verification.md)`. Between skills, use relative links such as `[plan handoff](../flow-plan/references/plan-handoff.md)`. Put fictional template links in fenced examples. Runtime project inputs are supplied or discovered in the task; they are not bundled dependencies.

Native agent renderers can flatten or relocate role files. In role instructions, identify a skill resource with `[plan handoff](skill:flow-plan/references/plan-handoff.md)` instead of a path relative to the source role. This is an authoring locator: the owner is `flow-plan` and the path within it is `references/plan-handoff.md`. It does not create a URI handler, invoke a tool, or preload a skill. Do not embed a developer's absolute installation path.

## Runtime resolution

The essential runtime boundaries and read conditions remain in `content/guidance/global.md`; the delegation procedure has its canonical home in `content/skills/flow-build/references/delegation.md`, read from the global guidance's «Proportionality» and from each flow that delegates. Resolve the named skill through the host's available catalog/loader or an explicit task path, then read the resource relative to that skill directory. Reading a reference does not invoke its owner's complete workflow; a resource that is read is read whole, under the resource rule in `content/guidance/global.md`, and `tests/content/skill_size_test.go` keeps each reference at or under 16,000 bytes (issue #204). Native discovery, parent context inheritance, and automatic reference reading are different capabilities; do not assume one establishes another.

Canonical procedural homes under `content/`:

- `skills/flow-walkthrough/SKILL.md`: publish UI captures and their iteration to the viewer (`hive view work capture`), and close a work with captures, or deliver a UI walkthrough through the supported fallback. Its consumers link to it instead of repeating publication steps; `hive view work capture --help` owns the capture command and `hive view work contract` the advanced semantic publisher interface and `hive view contract` the advanced machine-readable protocol.
- `skills/hive-ui-criteria/SKILL.md`: the precedence of UI sources (the project's accessibility requirement, recorded decisions, customization layer, library defaults, then the skill's defaults), the `DESIGN.md` design record and the UI criteria with their measured checks. `flow-plan`, the design, build and UX review roles and `flow-verify` link to it instead of restating them.
- `skills/flow-figma-design/SKILL.md`: the drawing of screen examples in Figma, the business rules the frames decide, the choice, the record of the approved example and the request whose deliverable is the design; `flow-research` only recommends the stage in its next-step block. `flow-plan`, `flow-build`, `flow-verify`, `hive-review-ux` and the site's flow pages link to it or compare against the approved example it records instead of restating it; the brief of the `hive-review-ux` child that reviews its frames is owned by `skills/flow-build/references/delegation.md` «Reviewer brief»; the `**Design file:**` and `**Design library:**` fields it reads are owned by `skills/hive-ui-criteria/references/design-record.md`.
- `skills/flow-figma-design/references/rule-cards.md`: the rule cards drawn beside each frame, their rows, their update when a rule is answered and their correction after the design stage; owned by `flow-figma-design`, which reads it at steps 5 and 7, and read from later flows when a card's rule changes. It links to the business rules reference for the rule itself.
- `skills/flow-plan/references/business-rules.md`: what a business rule is, the record's line format and statuses, where the record lives, and the write of the rules to the work's ticket at the design approval, the build start and the close; owned by `flow-plan`, which keeps the record in the change folder's `proposal.md`. `guidance/global.md` «Scope and authorization», `flow-figma-design`, `flow-build`, `flow-verify` and `flow-close` link to it instead of restating it; `hive-review-ux` links to it for the definition of a rule.
- `skills/flow-verify/SKILL.md` and its conditional `references/ui-verification.md`: verification, review, UI checks and proportional capture coverage. Publication consumers point here for those checks instead of restating a second capture policy.
- `guidance/global.md` «Questions and routes»: the decision format (`Situation:`, `If you choose it:` and `I recommend` lines) and the finding format, the reasons shown to the user, stated as effects and never as what an instruction the agent follows requires, except one the user or their project wrote, the context line and waiting line of a decision, result, blocker or closing message, and the next-step block between flows: which flow commands it offers, the single actions asked directly instead, and the compaction reply at plan → build and after a close.
- `guidance/global.md` «Scope and authorization»: effects beyond this machine, critical effects, third-party requests and delegation authority; «Scope changes»: a change the user asks for, or the work finds it needs, while work is open.
- `guidance/global.md` «Delivery»: the delivery mode and the code-review question, their recommendation and the effects each answer authorizes; `skills/git-workflow/SKILL.md` performs those effects, including the reviews after the push, and reads Git and pull-request state through its `scripts/git_state.py`.
- `guidance/global.md` «Project settings», «Language», «Secrets», «Evidence» and «Persistent memory»: the rules that used to live in `hive-core` references (dissolved into the global guidance in `5e6e8907`).
- `skills/flow-build/references/delegation.md`: before deciding whether a child does a piece of work, in any flow, briefing it or waiting for it; `skills/flow-build/SKILL.md`: when the turn may end during a build, and the `tasks.md` checkpoint, written with `skills/flow-plan/scripts/tasks.py`.
- `skills/flow-plan/SKILL.md` «Change folder» and its `assets/change-folder/` templates: creating or continuing a change record; `skills/flow-close/SKILL.md` closes and archives it. `skills/flow-plan/references/plan-handoff.md`: the plan summary and the build route.
- `skills/flow-close/SKILL.md` «Final report» and «Findings»: the report of a work item and the severity, grouping and recording of findings; «Support material»: the inventory of the work's own support material and the actions the user answers at close.
- `skills/workspace-conventions/references/artifact-layout.md`: before writing or cleaning support artifacts; basic placement remains global.

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

The supported authoring subset is ordinary inline Markdown links and reference definitions with simple destinations (optionally angle-bracketed or with a double-quoted inline title). Use uncomplicated paths without parentheses. The check covers the global core, role Markdown, skill entrypoints, and Markdown under their `references/` directories. It validates relative and `skill:owner/path` targets, checks logical owners, and rejects unsupported local schemes and personal installation paths. A directory link must end in `/` and contain packaged files.

Fenced examples, asset contents, arbitrary prose/code dependencies, heading anchors, and remote URL availability are outside this check. HTTP(S) and mail links are external. This is a packaging gate, not a full Markdown parser or evidence of runtime reading, permissions, tool availability, or behavioral compliance.

## Refactoring and delivery

Before renaming or removing a skill/resource, find its consumers with `rg`, update their links and read conditions, and run `go test ./...` and `go vet ./...`. Test a missing resource and saved-plan rejection when changing validation. Do not duplicate a resource merely to avoid a cross-skill dependency.

Current delivery selects hosts and scope and ships the complete skill catalogue. No per-skill dependency resolver is needed. If partial skill selection is introduced later, define and test dependency closure and removal behavior before enabling it; never silently omit a required resource. Deployment remains explicitly authorized and uses the existing manager.

## Declared reading footprint

`tests/content/reading_footprint_test.go` reports every selected source and counts
its UTF-8 bytes and Unicode characters once per scenario. It compares editorial
baseline A (`d5164f71`) to working-tree B; C uses the same canonical files with
the optional loader active. This is a declared source union, not an observed
request, cumulative session cost or proof that references were read. Missing
required files fail measurement. The size quota of a `SKILL.md` is set by
harness-audit HA-SK-02 and checked by `tests/content/skill_size_test.go`.

The predicates and full file lists are in `TestGuidanceReadingFootprint`:

- **Simple query:** factual reply without an activity, decision, delegation or
  artifact. B reads global plus reporting, which the scenario counts; A has reporting in global.
- **Question:** standalone user decision, with a waiting card or text. Add
  questions to that common union; neither scenario invokes a whole activity.
- **Plan/build cycle:** retained backend runtime plan, settings and hold/review
  decision, read-only plan reviewer, delegated backend implementer and task
  verifier, local tests, then pause through Close. Include the selected
  Plan/Build/Close entrypoints, the Verify entrypoint and its review-and-gates
  and runtime-gates references (stage handoff from Build), Research's planning exploration criterion,
  plan format/review, settings/change records, delivery decisions, verification,
  delegation, artifact layout, questions/option scope/reporting and the three
  role sources.
  No UI, security/data migration, deployment, Git execution or incidental finding
  is assumed, so their conditional procedures are not added.

Snapshot measured on 2026-10-23 at `2d89940e` (source files can change; rerun the test):

| Scenario | A bytes | B/C bytes | Byte change |
| --- | ---: | ---: | ---: |
| Global alone | 46,912 | 33,710 | -28.14% |
| Simple query | 46,912 | 49,107 | +4.68% |
| Question | 46,912 | 59,365 | +26.55% |
| Plan/build cycle | 173,839 | 229,774 | +32.18% |

The global decrease is not total-context savings: reporting, which the
simple-query scenario counts, already makes that union larger than A's, and the
declared cycle grows. Under the current rule a short direct answer with no
citation or relay reads only the global file, the first row. C's plan-review injection
adds no unique source to that cycle, but delivery wrappers, repetition and timing
can change actual requests. The measurement includes raw source frontmatter;
it excludes native skill/role discovery metadata and rendering, plugin metadata
and wrappers, repeated reads/copies across readers and requests, project inputs,
history and host policies. It does not convert bytes to measured tokens.
Actual delivery, token usage and task outcomes need native observations; the
bounded pilot is future evidence, not a claim of small-model reliability. It has
not run and remains blocked on verifying safe, exact memory cleanup; source and
callback checks do not resolve that separate prerequisite.

## Vendored upstream material

`content/skills/flow-report/references/diagram-grammar.md` and `diagram-svg-construction.md` (one numbering, split at 16,000 bytes) adapt [cathrynlavery/diagram-design](https://github.com/cathrynlavery/diagram-design) v2.3.2 (MIT). To pull upstream updates, re-read that repository's `skills/diagram-design/SKILL.md` §4–§9 and `references/type-*.md`, then re-map them onto the two references. Deliberately not adopted: Mermaid/draw.io import pipelines, animation, PNG/SVG export, brand onboarding, and long-tail types (medallion, radar, venn, DP matrices, loop, org chart).

`content/skills/test-audit/SKILL.md` adapts the `test-audit` skill of [openclaw/openclaw](https://github.com/openclaw/openclaw/tree/main/.agents/skills/test-audit) at `3195fba90643` (MIT; notice in `content/skills/test-audit/references/license-openclaw.md`). It keeps the R/F/C/D ledger, the retention bar, keepers before deletions and the independent preservation review, takes its authoring criterion from flow-build's Tests section instead of restating it, and adds targeted mutations and suite wall time to the report. To pull upstream updates, re-read that directory's `SKILL.md`, `references/audit.md` and `CAMPAIGN.md`. Deliberately not adopted: OpenClaw's own scripts (`check-changed.mjs`, the `deadcode:*` gates), its `openclaw-testing` proof policy, and per-subsystem `AGENTS.md` test rules.

`content/skills/hive-ui-criteria/` is original text. Its methodology (screen modes, product and design records, accessibility minimums, refinement passes) is inspired by [Impeccable](https://github.com/pbakaus/impeccable) by Paul Bakaus (Apache-2.0), and the design record follows the [DESIGN.md format](https://github.com/google-labs-code/design.md) of Google Labs (Apache-2.0). No passage of either project was adapted, so the redistribution terms of their licenses do not apply and no license copy is included. `references/notice.md` carries the credit and is deliberately not linked from `SKILL.md`. To pull upstream updates, re-read both repositories and compare the methodology, not the wording.
