# project-layout

## Purpose

TBD: created by archiving project-layout-skills; update the purpose.

## Requirements

### Requirement: One home for the project layout convention

Where each support artifact, document, script and data set of a project belongs SHALL be stated in one skill, `project-layout`, whose layout reference every flow reads before writing a support artifact. It SHALL name the destination of a script every clone or CI needs (the owning repository's code, outside `_support/`), of a script run once against shared data (its session folder, with its output and its revert), of an operation's data (`_support/data/`, never versioned), and of the project's index, and it SHALL forbid application code from reading or writing `_support/`.

#### Scenario: A script every clone needs

- **GIVEN** a work item writes a script that CI or every clone of the repository runs
- **WHEN** the agent chooses where to put it
- **THEN** it goes in the code of the repository whose code or data it uses, not under `_support/`

#### Scenario: A script run once against shared data

- **GIVEN** a work item runs a backfill against a shared database
- **WHEN** the work closes
- **THEN** the script stays in the work's session folder with its output and its revert, and the report names it

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

### Requirement: A script climbs a lifecycle before it becomes a tool

The layout convention SHALL define the levels of a script (disposable, run once, candidate, internal tool, mechanism) and what moves it up: a candidate is a purpose seen in two sessions, and an internal tool carries no ticket ID or absolute path, reads secrets from their location, defaults to a dry run, is idempotent and has a test.

#### Scenario: A repeated ticket script

- **GIVEN** a candidate record counts the same seeding purpose in two sessions
- **WHEN** the user turns it into a mechanism from the close footer
- **THEN** the script becomes an internal tool in the owning repository with a test and a line in the index

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

### Requirement: A migration asks for the knowledge hub when it is not clear

When a project's `Specs` or its workspace hub is missing or unresolved, the migration SHALL ask the user to choose among the detected specs repository, the repository that owns the contracts, a workspace repository created with `workspace-repo`, or local `_support` only, saying what each leaves outside Git; it SHALL propose moving the documentation home to `_support/docs` and keep the existing one when the user declines, and SHALL apply the four-point viability test before recommending a monorepo.

#### Scenario: A workspace without a declared hub

- **GIVEN** a workspace of three repositories whose `Specs` values disagree
- **WHEN** the user asks to organize its layout
- **THEN** the manifest stops at the hub question with the four options before any move

### Requirement: A workspace repository is created only on request

Git SHALL be initialized at a declared workspace root only through `workspace-repo` at the user's request. The skill SHALL print its plan without writing by default, and on apply SHALL write a `.gitignore` that ignores each child repository, `HIVE.local.md` and `_support/{workspace,data,scripts,secrets}`, a `.ignore` that re-includes each child for search, the repository table and the workspace's complete `## Hive`, whose tracker and whether it declares `Delivery: direct-base` the user answers; it SHALL never create a remote or push.

#### Scenario: Search from the workspace root

- **GIVEN** a workspace repository created by `workspace-repo` with two child repositories
- **WHEN** the agent searches for a word that appears only in a child from the root
- **THEN** the search finds it, and `git status` at the root lists no child content

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

### Requirement: The planner leaves in place what the user keeps and what other projects own

The placement planner SHALL leave where it is each item the user keeps, through its own option and never through an edited plan, together with every row under a kept folder and the tests and libraries only kept files use, with no destination, no question and no reference rewrite; SHALL ask about a kept test whose subject moves; and SHALL refuse a kept path that names no row or that an override also names. A dependency folder the repository commits, such as Go's `vendor/` beside `go.mod`, SHALL never count as one of its packages.

#### Scenario: An item the user keeps

- **GIVEN** the user keeps `_support/scripts/detect-changes.sh` and `_support/infrastructure/`
- **WHEN** the planner runs with `--keep` for each
- **THEN** those rows and the script's test have no destination and the note «kept by user», no question and no reference rewrite, and a `--keep` that names no row is refused

#### Scenario: A committed vendor library

- **GIVEN** a Go repository that commits `vendor/acme/composer.json` and has `_support/scripts/acme/check.sh`
- **WHEN** the planner runs
- **THEN** it does not send the script into `vendor/acme/`

### Requirement: A move keeps what each build and package takes, and reports what Git ignores

The placement planner SHALL read each `.dockerignore` (or `<name>.dockerignore`), `.gcloudignore`, `.vercelignore`, `.npmignore` and `.helmignore` for the paths under its folder, and for each move whose destination it leaves in or out of its build differently than the source, SHALL keep the move, ask about the new folders in one review per file and direction, and record that verdict; its verification SHALL fail while such a file judges the moved item otherwise than its source, unless a line of it names the destination or a folder the move created. The planner SHALL list, as a review, the files Git ignores in a folder the move empties of tracked files, and its verification SHALL list those still left there without failing.

#### Scenario: Terraform leaves a Docker exclusion

- **GIVEN** a `.dockerignore` that excludes `_support` and Terraform under `_support/infrastructure/terraform/`
- **WHEN** the planner sends it to `infra/terraform/` and the plan is applied
- **THEN** the plan asks about `infra/` for `.dockerignore`, and the verification fails until `.dockerignore` excludes `infra/` or names it

#### Scenario: A provider cache left behind

- **GIVEN** an ignored `.terraform/` beside the Terraform files that move
- **WHEN** the planner runs and the plan is applied
- **THEN** the plan lists the ignored file in a review of the emptied folder, and the verification notes it without failing

### Requirement: A reference the user keeps stays and passes verification

The placement planner SHALL leave unchanged each reference in a file, or on a line, the user keeps through its own option, without a rewrite or a review, SHALL record it in the plan so that the verification does not count it, and SHALL refuse a kept reference that matches none.

#### Scenario: A CI selector that matches old paths on purpose

- **GIVEN** a CI script that still names `_support/scripts/verify-push.sh` to classify renames
- **WHEN** the planner runs with `--keep-ref` for that line and the plan is applied
- **THEN** the line keeps the old path, the verification does not report it, and a `--keep-ref` that matches no reference is refused

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
