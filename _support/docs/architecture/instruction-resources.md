# Instruction resources

## Authoring

Keep each procedure and resource in one owning skill. Its `SKILL.md` states the essential procedure and links situational resources with the condition that requires reading them. A reference may link further resources needed for that situation; do not preload every reference for every task.

Use Markdown links for packaged dependencies, including scripts and assets. Within a skill, use paths relative to the containing file, such as `[verification](references/verification.md)`. Between skills, use relative links such as `[UI planning](../flow-plan/references/ui-planning.md)`. Put fictional template links in fenced examples. Runtime project inputs are supplied or discovered in the task; they are not bundled dependencies.

Native agent renderers can flatten or relocate role files. In role instructions, identify a skill resource with `[UI planning](skill:flow-plan/references/ui-planning.md)` instead of a path relative to the source role. This is an authoring locator: the owner is `flow-plan` and the path within it is `references/ui-planning.md`. It does not create a URI handler, invoke a tool, or preload a skill. Do not embed a developer's absolute installation path.

## Runtime resolution

The essential runtime boundaries and read conditions remain in `content/guidance/global.md`; the delegation procedure has its canonical home in `content/skills/flow-research/references/delegation.md`. Resolve the named skill through the host's available catalog/loader or an explicit task path, then read the resource relative to that skill directory. Reading a reference does not invoke its owner's complete workflow. Native discovery, parent context inheritance, and automatic reference reading are different capabilities; do not assume one establishes another.

Canonical procedural homes under `content/skills/`:

- `flow-close/references/reporting.md`: before a report of work or findings, a relay of a child's or tool's result, an explanation of how something works or why it failed, or a message that cites links or IDs defined outside it, as `global.md` sets; a greeting or a short direct answer needs no read, and a progress or decision message opens with the context line the reference sets.
- `flow-plan/references/questions.md`: before a user decision or manual validation request.
- `flow-research/references/delegation.md`: before dispatching a child or waiting for external work.
- `workspace-conventions/references/artifact-layout.md`: before writing or cleaning support artifacts; basic placement remains global.
- `flow-plan/references/project-settings.md`: before resolving needed settings; `flow-plan/references/change-records.md`: before creating or continuing a change record.

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
  Plan/Build/Close entrypoints, Research's planning exploration criterion,
  plan format/review, settings/change records, delivery decisions, verification,
  delegation, artifact layout, questions/reporting and the three role sources.
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

`content/skills/flow-report/references/diagram-grammar.md` adapts [cathrynlavery/diagram-design](https://github.com/cathrynlavery/diagram-design) v2.3.2 (MIT). To pull upstream updates, re-read that repository's `skills/diagram-design/SKILL.md` §4–§9 and `references/type-*.md`, then re-map them onto the reference. Deliberately not adopted: Mermaid/draw.io import pipelines, animation, PNG/SVG export, brand onboarding, and long-tail types (medallion, radar, venn, DP matrices, loop, org chart).
