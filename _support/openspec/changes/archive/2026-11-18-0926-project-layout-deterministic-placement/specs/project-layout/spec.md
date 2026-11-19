# project-layout — delta

## MODIFIED Requirements

### Requirement: A project's layout exceptions are declared in its index

A project that keeps a home other than the convention's, such as `docs/` at the root after declining the migration or an existing infrastructure folder, SHALL declare it in a `Layout: docs at <path> · tools at <path> · infra at <path>` line at the head of its index, found at `_support/docs/README.md`, else `docs/README.md`, else the path an `Index: <path>` line of `AGENTS.md` names; agents writing support artifacts SHALL follow that line over the default destinations, the layout check and the placement planner SHALL read it before their rules, and the line SHALL be written only when the user declines moving a home or overrides a planned home, never on the agent's own choice.

#### Scenario: A declined documentation migration

- **GIVEN** the user declines moving a repository's `docs/` to `_support/docs`
- **WHEN** the migration finishes and the check runs later
- **THEN** `docs/README.md` holds `Layout: docs at docs/`, the check does not report `docs/` as an undeclared documentation home, and it still reports a document in `docs/` missing from the index

#### Scenario: An undeclared documentation home

- **GIVEN** a repository with Markdown under a root `docs/` and no `Layout:` line
- **WHEN** the check runs
- **THEN** it reports `docs/` as an undeclared documentation home with the correction to migrate or declare it

#### Scenario: A user override of a planned home

- **GIVEN** the planner sends a project's CI scripts to `scripts/ci/`
- **WHEN** the user answers the manifest with another root for them
- **THEN** the plan marks those rows as the user's and prints the `tools at` line to add, and no `Layout:` line appears without such an answer

### Requirement: The layout is checked without judgment

`project-layout` SHALL ship a script that, read-only, fails on each unregistered local script, each tracked file under `_support/{scripts,secrets,workspace,data}`, each tracked code or infrastructure file under another `_support/` entry other than `docs`, `openspec` and `sessions`, each document missing from the index, each broken index link, and, in a workspace repository, each listed repository that is missing, has another remote, is not ignored or is not re-included for search; and that reports for review, without changing its exit code, each tracked code file outside `_support/` and outside tests that in a line that is not a comment line cites work material under `_support/` (its documentation and change records may be read; configuration and documentation files are exempt) and each top-level `_support/` entry outside the convention; each finding SHALL state its correction. It SHALL never count files inside dependency or build output folders of any ecosystem: inside a Git repository it SHALL list tracked and untracked files Git does not ignore plus the ignored files under `_support/`, where the convention keeps unversioned material, and outside one it SHALL prune child repositories, a fixed list of dependency folders, and build folders whose producing manifest sits beside them.

#### Scenario: Code that cites work material

- **GIVEN** a tracked `src/load.py` that opens `_support/data/rows.csv`
- **WHEN** the check runs
- **THEN** it reports the line for review and its exit code does not change

#### Scenario: An unregistered local script

- **GIVEN** `_support/scripts/seed.sh` exists and no `HIVE.local.md` names it
- **WHEN** the check runs
- **THEN** it reports the file with the correction to name it in `HIVE.local.md` or move it to its owning repository, and exits non-zero

#### Scenario: Dependencies inside the documentation home

- **GIVEN** a documentation site under `docs/` with an ignored `node_modules/` full of Markdown
- **WHEN** the check runs
- **THEN** no file inside `node_modules/` is reported

#### Scenario: Infrastructure kept under the support folder

- **GIVEN** a tracked `_support/infrastructure/terraform/main.tf`
- **WHEN** the check runs
- **THEN** it fails on that file with the correction to run the placement planner

## ADDED Requirements

### Requirement: What project-layout moves gets the same destination on every run

`project-layout` SHALL ship a placement planner that, without changing the tree, computes the destination of each script, code or infrastructure file and each legacy folder the check reports, from observable signals only (who cites the file and from where, its name, its extension and the folders the project already has) through rules in a fixed order with a closed vocabulary of destinations, and SHALL print the same plan for the same tree. The organize mode SHALL copy the planner's rows into its manifest instead of choosing destinations, and an item no rule places SHALL be a question for the user, never an invented destination.

#### Scenario: The same evaluation twice

- **GIVEN** a repository with scripts under `_support/scripts/` called from CI, a hook, one package and a runbook
- **WHEN** the planner runs twice
- **THEN** both plans are identical and send them to `scripts/ci/`, `scripts/hooks/`, `<package>/scripts/` and `scripts/ops/`, reusing an equivalent group the project already has

#### Scenario: Terraform under the support folder

- **GIVEN** `_support/infrastructure/terraform/` in a repository with no infrastructure folder
- **WHEN** the planner runs
- **THEN** it sends the tree to `infra/terraform/`

#### Scenario: Legacy evidence

- **GIVEN** an untracked `_support/evidence/2026-01-02-abc-1-login/`
- **WHEN** the planner runs
- **THEN** it sends it to `_support/workspace/2026-01-02-abc-1-login/evidence/` and removes nothing

### Requirement: A move repairs the references to what it moves

When the planner's plan is applied, each current reference to a moved path SHALL be rewritten in the form it had, relative to the repository, to the citing file or to the workspace; a prefix or glob that covers files with different destinations, and a script's own relative path whose depth changes, SHALL be reported for review instead; references in historical records SHALL stay unchanged; references inside the moved files SHALL be rewritten too; the apply SHALL refuse a stale plan or an existing destination without moving anything, SHALL never commit, SHALL record what it did so that an undo reverses exactly that and keeps unrelated local changes, and its verification SHALL fail while a current reference to an old path remains.

#### Scenario: A package script called with a relative path

- **GIVEN** `apps/backend/package.json` runs `bash ../../_support/scripts/backend/check.sh`
- **WHEN** the plan is applied
- **THEN** the file is at `apps/backend/scripts/check.sh`, the package script runs `bash scripts/check.sh`, and the verification passes

#### Scenario: A stale plan

- **GIVEN** a plan computed before a source file changed
- **WHEN** it is applied
- **THEN** the apply refuses and nothing moves

#### Scenario: Undo next to another work's changes

- **GIVEN** a working tree with another work's uncommitted edit in a file the plan rewrites
- **WHEN** the plan is applied and then undone
- **THEN** the tree's status and diff equal those before the apply, the other work's edit included
