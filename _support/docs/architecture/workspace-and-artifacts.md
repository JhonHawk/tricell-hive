# Workspace and artifact organization

This document explains the artifact policy in [the shared global guidance](../../../content/guidance/global.md). That source carries the essential operating rules without requiring a skill or this document to load. The [workspace-conventions skill](../../../content/skills/workspace-conventions/SKILL.md) is the canonical procedure for organizing existing material, curating evidence, promoting findings, and archiving. This reference provides examples and rationale; it is not a second router or an installer.

## Resolve scope before choosing a folder

A workspace is a declared project context containing related repositories, not every directory above a Git root. Resolve it from the user's task and established project instructions or documentation. Git can identify a repository or worktree root, but does not establish which sibling repositories form a project. If shared scope is unclear, resolve that ambiguity before writing outside the known repository.

| Work shape | Default support home | Example |
|---|---|---|
| Standalone repository | `<repo>/_support/` | A backend investigation specific to that repository |
| One repository inside a multi-repo workspace | `<repo>/_support/` | A frontend-only implementation record |
| Work spanning related repositories | `<workspace>/_support/` | An API/consumer compatibility investigation |
| Monorepo | `<repo>/_support/` | Application grouping within `docs/<app>/`, not `apps/<app>/_support/` |
| Existing documented artifact destination | The established destination | Shared decisions already maintained in a documentation repository |

Apply this routing to new work. Continuing work retains its established scope and home; a later change of scope may justify a proposed migration, not a silent move. Existing package-level documentation or support folders are not automatically relocated.

An existing documentation repository may provide equivalent homes such as `sessions/` or `decisions/`; do not nest another `_support` merely to match a template. Without a documented versioned home, keep shared durable material locally in the workspace's `_support` and report that it is local. Selecting a versioned destination is a separate decision; creating a specs repository, initializing Git, or publishing is not implied.

## Project settings and change records

The [global guidance](../../../content/guidance/global.md) defines the `## Hive` section, with its example, and the `openspec/` change records. Each repository declares its settings in full because Codex never loads files above the Git root and Grok Build was observed not to load a workspace `AGENTS.md` from a child repository; only Claude Code loads ancestor files. A specs or mocks repository declares a shorter section, such as `Specs: openspec` in the specs repository; a repository outside the project never inherits another project's values.

Planned work lives in `<specs>/changes/<change-id>/` and current requirements in `<specs>/specs/`, where `<specs>` is the `openspec` directory that `Specs` names: `_support/openspec` inside a repository without a specs repository, or `<specs-repo>/openspec`. `_support/openspec` must stay versioned; repositories here ignore only `_support/workspace/` and, in some, parts of `_support/evidence/`. The layout follows OpenSpec file conventions without its CLI. It replaces `_support/sessions/` for plans and the unversioned workspace ledger for project state: `openspec/project.md` keeps only what no other source states. Existing sessions, epics, and ledgers stay as history. See the [change records reference](../../../content/skills/flow-plan/references/change-records.md) for file roles, the delta format, and closure.

## Four destinations, created when needed

```text
<scope>/_support/
├── docs/
│   └── <topic>/
├── sessions/
│   ├── YYYY-MM-DD-<work>/
│   │   ├── <work>.research.md   # Optional retained investigation
│   │   ├── <work>.report.md     # Optional separate human deliverable
│   │   └── reports/            # Optional human deliverables
│   └── archived/
├── workspace/
│   └── YYYY-MM-DD-<work>/
└── evidence/
    └── YYYY-MM-DD-<work>/
```

This is a vocabulary, not a scaffold to create for every request. A single useful record is sufficient when it carries the work. Do not create empty files, folders, or indexes solely to match the diagram.

- **Knowledge:** `docs/` describes current conventions, decisions, or architecture. A living guide is updated in place; a dated historical decision can be superseded with an explicit reference to its successor. Keep an existing ADR or contract convention when present.
- **Execution:** planned work lives in its `<specs>/changes/<change-id>/` folder. `sessions/` keeps only records tied to no change, such as research or a report requested on its own. An enduring design conclusion belongs in its documentation home.
- **Workspace:** `workspace/` holds material the user may open or whose versioning is still undecided: working documents, review images, temporary information that may later be versioned or discarded, and files a tool or later session must find again, such as saved browser state or a manager's `--out` plan, which the manager rejects under `/tmp`. Ignore its contents in Git. The current Hive scaffold retains only selected `.gitkeep` markers; this does not make its files versionable or require markers in other projects.
- **Scratch:** temporary utilities, shell helpers, intermediate transformations, and raw outputs nobody needs after the command or task go where the host keeps temporary work: its declared scratch or session directory, otherwise the operating system's temporary directory, never the repository root. Hosts differ, so the guidance names no single path. Anything the user should read goes to `workspace/` instead: a review file left in `$TMPDIR` (a sample-project Grok session, 2026-09-24) was hard to find and could be purged by macOS. The test is disposability, not temporariness: scratch must be reproducible, so cleanup can never destroy the only copy of work. A Git worktree holding unintegrated edits fails that test even though it is temporary; keep it in the host's native worktree location or one the user approves.
- **Evidence:** `evidence/` contains a selected subset needed to substantiate results. A report interprets that evidence and links to it. Curating evidence is distinct from deciding whether it is suitable for Git.

Source code, permanent maintenance scripts, application configuration, and infrastructure remain in their established project locations. They do not become scratch because an agent wrote them.

## Naming and continuity

The shared guidance defines the naming convention. New records use `<topic>.<type>.<extension>` with descriptive English `kebab-case` topics and English types. For example, `context-update.research.md` and `context-update.report.md` remain identifiable outside their parent folder; change folders use fixed file names. Dated directories start with the work's initial date, `YYYY-MM-DD`, followed by its stable slug; internal records do not repeat the date. Avoid generic standalone names such as `output.md` or `temp.json`. Keep tool entrypoints (`AGENTS.md`, `SKILL.md`, `README.md`) and existing source conventions. Historical records retain their names unless a separate organization task covers their migration.

New records use the dotted form, but both `<topic>.plan.md` and the established legacy `<topic>-plan.md` form are valid handoffs. The same compatibility applies to research and reports. Resume the existing record and do not rename files merely to normalize their form.

A new retained plan is a change folder whose `tasks.md` carries its task list and progress. A separate research or report file is useful when it has its own audience or substantial evidence; the vocabulary does not require three files. Retain plans when requested or needed for resumption, delegation, coordinated deliveries, consequential decisions, or investigation that would be costly to reconstruct. Small understood changes can use a conversational plan. Preparation is separate from persistence: the `flow-plan` skill owns the procedure for grounding decisions and connecting requirements, tasks, and verification; file presence does not prove readiness.

One work item keeps one change or session folder across conversations. Supporting scratch and evidence use the same initial date and slug. A new conversation or calendar day does not create a new work item. When meaningful phases need separation, add descriptive subfolders inside the existing work folder rather than scattering it across sibling sessions. An independent new objective gets a new folder.

For example, work started under `sessions/2026-09-20-api-contract/` still resumes there on the next day. Temporary traces can use `workspace/2026-09-20-api-contract/`, and selected evidence can use `evidence/2026-09-20-api-contract/`. A resulting ongoing convention belongs in `docs/api/contracts.md`, linked back to the session. Keep the session's historical findings as history, not a second live convention.

Dates identify historical snapshots. Living documents such as `docs/api/contracts.md` need no date prefix. No mandatory per-directory README is required when naming and links already make the contents discoverable; use an index when it supplies useful navigation or lifecycle information.

Agent instructions, operational contracts, and guidance use English. Human reports use the session language unless requested otherwise. A Spanish report inside `reports/` can support an English architecture decision; the document's purpose determines its language, not its directory or eventual readers.

## Retention, versioning, and hygiene

These are separate questions: does the artifact need to survive, where can it safely live, and should it enter Git?

| Artifact | Retention and location | Git treatment |
|---|---|---|
| Reusable decision or resumable work record | Existing durable home or appropriate `docs/` / `sessions/` | Suitable for versioning when non-sensitive; not an automatic commit |
| Regenerable intermediate output no longer needed | The host's temporary location; remove at task close when created by this task | Not in the repository |
| Working document, review image, or output whose versioning is undecided | Task-local `workspace/` | Ignored until a decision moves it to a versioned home |
| Unique failure trace or transient external-state capture | Preserve the useful subset as evidence; private local storage if needed | Never publish a raw dump merely to retain it |
| Selected binary needed to explain a result | Curated `evidence/`, with a report reference | May be versioned when justified and suitable for sharing |
| Sensitive material or recovery backup | Existing protected local location; no new storage convention imposed | Excluded; not treated as disposable scratch |

Re-running an application may not reproduce a transient failure, remote state, or the same screenshot. Reproducibility must be assessed rather than inferred from the ability to rerun a command. If evidence is necessary but unsuitable for Git, preserve it privately and explain that a clean clone will not contain it. Label snapshots with their observation date; revalidate them before claiming current state.

At close, retain useful conclusions, link their evidence, and remove only the task's own reproducible temporary files that are no longer needed. Do not blanket-delete a work directory containing prior or unique material. Uncertain ownership or retention needs mean preserve and report, not guess. This standing cleanup permission does not cover old scratch, other users' files, authentication, memory data, or backups.

Archiving retains execution history without treating age as proof of completion. For example, closed research may be archived while an older unresolved investigation remains active. The skill contains the operating procedure, including closure checks, authorization, collisions, and link preservation; maintain those details there.

The [`engram-init-workspace` skill](../../../content/skills/engram-init-workspace/SKILL.md) provides an explicit dry-run helper when a declared workspace needs a shared local Engram identity. It requires an identity and selected targets, verifies all existing configurations before apply, and never infers sibling repositories, changes global ignores, or migrates memory. The [`workspace-archive` skill](../../../content/skills/workspace-archive/SKILL.md) handles selected closed-session moves; it has no age policy and never stages, commits, or pushes. Explicit unattended delegation is documented by [`unattended-delegation`](../../../content/skills/unattended-delegation/SKILL.md); it is a bounded coordination mode, not a workspace lifecycle mechanism.

## Limits of this implementation

These artifact conventions are authored instructions, not enforced filesystem behavior. No hook, artifact resolver, automatic workspace migration, archiver, or configuration database implements them. The separate [deployment manager](deployment-manager.md) installs guidance; it does not enforce task artifact placement. Host loading and behavioral outcomes are measured separately from filesystem installation.
