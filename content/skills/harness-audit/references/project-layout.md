# Project layout

Read when the audit covers a project's repositories, or when the caller asks for migration preparation. It checks the global layer's `## Hive` settings and change-record layout, and produces a migration manifest. The audit applies it only when the caller explicitly instructs it for an approved manifest, following the `workspace-conventions` procedure; otherwise it moves nothing.

## Checks (HA-PL)

- **HA-PL-01 — Specs target.** `Specs` in each repository resolves to an existing directory inside a Git repository and is not ignored (`git check-ignore`). A missing directory is a `create` entry; an ignored or unversioned path is `high`.
- **HA-PL-02 — One target per project.** Repositories of the same project point `Specs` at the same directory. Several repositories sharing a project without a specs repository need the caller's decision to create one or designate a repository.
- **HA-PL-03 — Legacy homes.** Find existing records outside the layout: `sessions/` folders, epic or delta folders, current-rule folders such as `product/`, unversioned workspace ledgers such as `_support/PROJECT.md`, and plans at repository roots. Classify each by its content, not its name.
- **HA-PL-04 — Current requirements.** Identify where current behavior is described today. A product map (`product/<module>/<view>.md` with rules in force) is the project's current-requirements home: `keep` it in place and never propose moving it into `<specs>/specs/`. Another folder that already holds only rules in force maps to `<specs>/specs/`; a document that mixes current rules with delivery history stays in place (`keep`) and remains authoritative for its areas until a change moves them, because extracting current requirements is authoring, not a move. Flag instructions that still call such documents the single source of truth so they can name the transition.
- **HA-PL-05 — Misplaced material.** Report content whose home is another repository or level, such as infrastructure configuration inside a specs repository, secrets or raw evidence inside a versioned path, or support material at a repository root.

## Migration manifest

Group entries by repository, in the order they can be applied. Each entry has:

| Field | Content |
| --- | --- |
| Operation | `rename`, `move`, `archive`, `create`, `keep`, `manual`, or `ask` |
| Source → destination | Repository-relative paths |
| Command | The exact `git mv` or `mkdir` for mechanical operations; none for `manual` or `ask` |
| Reason | The evidence that places it there |
| Links affected | Files that reference the source path |
| Risk | Collisions, links to rewrite, open work using the path |

Rules for proposing entries:

- Propose only mechanical moves with `git mv`; content that must be rewritten or split is `manual`.
- Leave history in place by default: closed sessions and epics become `keep` or `archive` under the project's legacy archive, never merged into current specs.
- Never propose moving active work mid-change; mark it `ask`.
- Mark anything sensitive, such as secrets or credentials, `keep` with the reason, and never include its contents.
- End with the questions the caller must answer before applying, and the checks that verify the result, such as `git status`, rename detection with `git diff -M --stat`, and a link check.
