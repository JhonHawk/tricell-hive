# Instruction resources

## Authoring

Keep each procedure and resource in one owning skill. Its `SKILL.md` states the essential procedure and links situational resources with the condition that requires reading them. A reference may link further resources needed for that situation; do not preload every reference for every task.

Use Markdown links for packaged dependencies, including scripts and assets. Within a skill, use paths relative to the containing file, such as `[verification](references/verification.md)`. Between skills, use relative links such as `[UI planning](../flow-plan/references/ui-planning.md)`. Put fictional template links in fenced examples. Runtime project inputs are supplied or discovered in the task; they are not bundled dependencies.

Native agent renderers can flatten or relocate role files. In role instructions, identify a skill resource with `[UI planning](skill:flow-plan/references/ui-planning.md)` instead of a path relative to the source role. This is an authoring locator: the owner is `flow-plan` and the path within it is `references/ui-planning.md`. It does not create a URI handler, invoke a tool, or preload a skill. Do not embed a developer's absolute installation path.

## Runtime resolution

The shared delegation rule in `content/guidance/global.md` owns the runtime contract. Resolve the named skill through the host's available catalog/loader or an explicit task path, then read the resource relative to that skill directory. Reading a reference does not invoke its owner's complete workflow. Native discovery, parent context inheritance, and automatic reference reading are different capabilities; do not assume one establishes another.

Deployed files carry installed paths. The installer rewrites every inline `[label](skill:owner/path)` link outside code fences, in role files for every host (a Codex role is rewritten before its TOML string is encoded) and in the global guidance block, to `<skills-dir>/<owner>/<path>` wrapped in `<...>`. At user scope `<skills-dir>` is the absolute `<home>/.agents/skills`, the shared store every host installs into (`~/.claude/skills/<skill>` links to it). At project scope it is the host's project skills directory relative to the project root, `.claude/skills` for Claude and `.agents/skills` for Codex, so versioned project files carry no personal path. The limitation: a relative path resolves from the project root, so a session opened in a subdirectory, or a worktree that lacks the project's versioned skills, does not find it directly. Reference-style definitions (`[x]: skill:...`) are not rewritten; none exists in `content/`. Source files keep `skill:` locators, and a file shared by several hosts must render to identical bytes or the plan fails before any write. The `skill:` resolution rule above still applies to a locator that reaches a model unrewritten, such as one inside a code fence.

The caller supplies relevant project conventions and task decisions. If resolution is ambiguous, inaccessible, or unavailable, report the dependency and pause only work requiring it; continue independent work. Distinguish a missing project guide from a required packaged reference that could not be read. Neither an installed file nor a listed skill proves a subagent read it.

## Release validation

The management package checks instruction links when preparing a release and when validating a saved plan before applying it. It resolves links against the release payload, not files left elsewhere in the checkout. Removing an owner or resource while a consumer still requires it fails validation before managed writes.

The supported authoring subset is ordinary inline Markdown links and reference definitions with simple destinations (optionally angle-bracketed or with a double-quoted inline title). Use uncomplicated paths without parentheses. The check covers role Markdown, skill entrypoints, and Markdown under their `references/` directories. It validates relative and `skill:owner/path` targets, checks logical owners, and rejects unsupported local schemes and personal installation paths. A directory link must end in `/` and contain packaged files.

Fenced examples, asset contents, arbitrary prose/code dependencies, heading anchors, and remote URL availability are outside this check. HTTP(S) and mail links are external. This is a packaging gate, not a full Markdown parser or evidence of runtime reading, permissions, tool availability, or behavioral compliance.

## Refactoring and delivery

Before renaming or removing a skill/resource, find its consumers with `rg`, update their links and read conditions, and run `go test ./...` and `go vet ./...`. Test a missing resource and saved-plan rejection when changing validation. Do not duplicate a resource merely to avoid a cross-skill dependency.

Current delivery selects hosts and scope and ships the complete skill catalogue. No per-skill dependency resolver is needed. If partial skill selection is introduced later, define and test dependency closure and removal behavior before enabling it; never silently omit a required resource. Deployment remains explicitly authorized and uses the existing manager.

## Vendored upstream material

`content/skills/flow-report/references/diagram-grammar.md` adapts [cathrynlavery/diagram-design](https://github.com/cathrynlavery/diagram-design) v2.3.2 (MIT). To pull upstream updates, re-read that repository's `skills/diagram-design/SKILL.md` §4–§9 and `references/type-*.md`, then re-map them onto the reference. Deliberately not adopted: Mermaid/draw.io import pipelines, animation, PNG/SVG export, brand onboarding, and long-tail types (medallion, radar, venn, DP matrices, loop, org chart).
