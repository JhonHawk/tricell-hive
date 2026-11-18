# project-layout

## ADDED Requirements

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

A project that keeps a home other than the convention's, such as `docs/` at the root after declining the migration, SHALL declare it in a `Layout: docs at <path> · tools at <path>` line at the head of its index, found at `_support/docs/README.md`, else `docs/README.md`, else the path an `Index: <path>` line of `AGENTS.md` names; agents writing support artifacts SHALL follow that line over the default destinations, the layout check SHALL read it before its rules, and the migration SHALL write it when the user declines moving the home.

#### Scenario: A declined documentation migration

- **GIVEN** the user declines moving a repository's `docs/` to `_support/docs`
- **WHEN** the migration finishes and the check runs later
- **THEN** `docs/README.md` holds `Layout: docs at docs/`, the check does not report `docs/` as an undeclared documentation home, and it still reports a document in `docs/` missing from the index

#### Scenario: An undeclared documentation home

- **GIVEN** a repository with Markdown under a root `docs/` and no `Layout:` line
- **WHEN** the check runs
- **THEN** it reports `docs/` as an undeclared documentation home with the correction to migrate or declare it

### Requirement: A script climbs a lifecycle before it becomes a tool

The layout convention SHALL define the levels of a script (disposable, run once, candidate, internal tool, mechanism) and what moves it up: a candidate is a purpose seen in two sessions, and an internal tool carries no ticket ID or absolute path, reads secrets from their location, defaults to a dry run, is idempotent and has a test.

#### Scenario: A repeated ticket script

- **GIVEN** a candidate record counts the same seeding purpose in two sessions
- **WHEN** the user turns it into a mechanism from the close footer
- **THEN** the script becomes an internal tool in the owning repository with a test and a line in the index

### Requirement: The layout is checked without judgment

`project-layout` SHALL ship a script that, read-only, reports each unregistered local script, each tracked file under `_support/{scripts,secrets,workspace,data}`, each tracked code file outside `_support/` that cites `_support/` (configuration and documentation files are exempt), each top-level `_support/` entry outside the convention, each document missing from the index, each broken index link, and, in a workspace repository, each listed repository that is missing, has another remote, is not ignored or is not re-included for search; each finding SHALL state its correction.

#### Scenario: An unregistered local script

- **GIVEN** `_support/scripts/seed.sh` exists and no `HIVE.local.md` names it
- **WHEN** the check runs
- **THEN** it reports the file with the correction to name it in `HIVE.local.md` or move it to its owning repository, and exits non-zero

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
