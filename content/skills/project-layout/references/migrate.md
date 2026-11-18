# Migrate a project's layout

Read at [project-layout](../SKILL.md) before step 3 when the request moves a project between layouts, when `Specs` or the workspace's knowledge hub is missing or unresolved, when a repository keeps its documentation outside `_support/docs`, and from harness-audit's onboarding for HA-PL-01. It sets the checks of the `## Hive` settings and the change-record layout, the questions a migration asks, and its manifest entries, which join project-layout's manifest; nothing moves before the user's answer to it. Edits to the `## Hive` section itself follow harness-audit's [apply](../../harness-audit/references/apply.md).

## Choose the layout

Two layouts are equally valid. Recommend a **monorepo**, one repository holding the code and `_support/openspec`, only when all four hold: the same people can write to every repository, and no client or third party needs only one; deployments can be selected by path in one CI; no repository has a publication contract of its own with outside consumers, such as a published package, a fork or a repository the client administers; and consolidating needs no merge of Git history, since the current state is copied with its provenance. Otherwise keep a **workspace** of several repositories, which can become a coordinator repository through [workspace-repo](../../workspace-repo/SKILL.md) when the user asks.

Every migration, in either direction, follows the same order: separate current knowledge from history first; record each moved document's provenance, its source commit and blob, and point historical links at the original commit; check from the entry point of a session, in an isolated copy, that the destination is enough on its own; deliver the destination before retiring the origin; and retire the origin in a way that can be restored, such as an archive tag.

## Questions a migration asks

Ask them in the round of the manifest, as the global guidance's «Questions and routes» sets:

- **Knowledge hub**, when `Specs` is missing, does not resolve, or differs between repositories of one project: the detected specs repository; the repository that owns the contracts, usually the backend; a workspace repository created with workspace-repo; or `_support` of the workspace only. Say what each leaves outside Git.
- **Documentation home**, when a repository keeps its living documentation outside `_support/docs`, such as a root `docs/`: propose moving it to `_support/docs`, with the links each move updates. When the user declines, keep it, and write at the head of its index the line `Layout: docs at <path>`, keeping any key the line already has, creating `<path>/README.md` when it has no index and pointing the root `AGENTS.md` at it with an `Index: <path>` line when the index is not `_support/docs/README.md` or `docs/README.md`.

## Checks (HA-PL)

- **HA-PL-01 — Specs target.** `Specs` in each repository resolves to an existing directory inside a Git repository and is not ignored (`git check-ignore`). A missing directory is a `create` entry, or an `ask` entry for a clone when the value carries a Git URL, because a clone fetches remote content; an ignored or unversioned path is `high`. A path outside the repository with no ` · <git-url>` is `medium`: a fresh clone cannot locate or fetch it.
- **HA-PL-02 — One target per project.** Repositories of the same project point `Specs` at the same directory. Several repositories sharing a project without a specs repository need the caller's decision to create one or designate a repository.
- **HA-PL-03 — Legacy homes.** Find existing records outside the layout: `sessions/` folders, epic or delta folders, current-rule folders such as `product/`, unversioned workspace ledgers such as `_support/PROJECT.md`, and plans at repository roots. Classify each by its content, not its name.
- **HA-PL-04 — Current requirements.** Identify where current behavior is described today. A product map (`product/<module>/<view>.md` with rules in force) is the project's current-requirements home: `keep` it in place and never propose moving it into `<specs>/specs/`. Another folder that already holds only rules in force maps to `<specs>/specs/`; a document that mixes current rules with delivery history stays in place (`keep`) and remains authoritative for its areas until a change moves them, because extracting current requirements is authoring, not a move. Flag instructions that still call such documents the single source of truth so they can name the transition.
- **HA-PL-05 — Misplaced material.** Report content whose home is another repository or level, such as infrastructure configuration inside a specs repository, secrets or raw evidence inside a versioned path, or support material at a repository root. These checks propose no removal of versioned evidence or legacy records: project-layout's classes in its step 3 decide them. `HIVE.local.md` at a repository or workspace root is not misplaced: the global guidance's «Local instructions» keeps it there.

## Migration manifest

Group entries by repository, in the order they can be applied. Each entry has:

| Field | Content |
| --- | --- |
| Operation | `rename`, `move`, `archive`, `create`, `insert`, `keep`, `manual`, or `ask` |
| Source → destination | Repository-relative paths |
| Command | The exact `git mv` or `mkdir` for mechanical operations; for `insert`, the exact line and the line it follows; none for `manual` or `ask` |
| Reason | The evidence that places it there |
| Links affected | Files that reference the source path |
| Risk | Collisions, links to rewrite, open work using the path |

Rules for proposing entries:

- Propose only mechanical moves with `git mv`; content that must be rewritten or split is `manual`.
- A repository used with Cursor CLI that lacks the HA-IF-18 pointer gets an `insert` entry: the canonical line into its root `AGENTS.md`, after the file's first paragraph and outside `## Hive`. `insert` is the only content edit a manifest carries; any other added or changed text is `manual`.
- Leave history in place by default: closed sessions and epics become `keep` or `archive` under the project's legacy archive, never merged into current specs.
- Never propose moving active work mid-change; mark it `ask`.
- Mark anything sensitive, such as secrets or credentials, `keep` with the reason, and never include its contents.
- End with the questions the caller must answer before applying, and the checks that verify the result, such as `git status`, rename detection with `git diff -M --stat`, and a link check.
