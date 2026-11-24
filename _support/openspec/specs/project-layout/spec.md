# project-layout

## Purpose

What the `project-layout` scripts and `workspace-repo` do: the read-only layout check, the placement planner, its apply and undo, and the workspace repository creation.

## Requirements

### Requirement: A project's layout exceptions are declared in its index

The layout check and the placement planner SHALL read a `Layout: docs at <path> · tools at <path> · infra at <path>` line at the head of the project's index, found at `_support/docs/README.md`, else `docs/README.md`, else the path an `Index: <path>` line of `AGENTS.md` names, before their own rules and in preference to the default destinations. The planner SHALL print that line only when a `--home` option changes a home; the line for a migration the user declined is written by the migrate procedure, not the planner.

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

`project-layout` SHALL ship a script that, read-only, fails on each unregistered local script, each tracked file under `_support/{scripts,secrets,workspace,data}`, each tracked code or infrastructure file under another `_support/` entry other than `docs` and `openspec`, each document missing from the index, each broken index link, and, in a workspace repository, each listed repository that is missing, has another remote, is not ignored or is not re-included for search; and that reports for review, without changing its exit code, each tracked code file outside `_support/` and outside tests that in a line that is not a comment line cites work material under `_support/` (its documentation may be read; configuration and documentation files are exempt) and each top-level `_support/` entry outside the convention; each finding SHALL state its correction. It SHALL never count files inside dependency or build output folders of any ecosystem: inside a Git repository it SHALL list tracked and untracked files Git does not ignore plus the ignored files under `_support/`, where the convention keeps unversioned material, and outside one it SHALL prune child repositories, a fixed list of dependency folders, and build folders whose producing manifest sits beside them.

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

The placement planner SHALL read each `.dockerignore` (or `<name>.dockerignore`), `.gcloudignore`, `.vercelignore`, `.npmignore` and `.helmignore` for the paths under its folder (a `<name>.dockerignore` from the repository root, the build context a monorepo uses, and a `.gcloudignore` together with the files its `#!include:` lines name), and for each move whose destination it leaves in or out of its build differently than the source, SHALL keep the move, ask about the new folders in one review per file and direction, and record that verdict; its verification SHALL fail while such a file judges the moved item otherwise than its source, unless a line of it names the destination or a folder the move created. The planner SHALL list, as a review, the files Git ignores in a folder the move empties of tracked files, and its verification SHALL list those still left there without failing.

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

### Requirement: Legacy records leave the tree one item at a time

The placement planner SHALL plan the legacy homes `sessions/`, `evidence/`, `reports/`, `archive/` and `archived/` at the root of each Git repository as it plans those under `_support/`: each child moves to `workspace/<date>-<work>/` of the support home, which is the repository's own `_support/` when it exists and otherwise the `_support/` of the workspace folder when that folder is also a root of the plan (written `../_support/workspace/...`), and each tracked child leaves Git. A top-level folder that code or configuration cites as a path (a document citing it holds nothing back), a folder that holds code, and a child that is a repository of its own SHALL go to review instead of moving, and a destination SHALL be compared by absolute path across every root of the plan. The planner SHALL plan the change archive one child per row: a child that is tracked and has no uncommitted, untracked, ignored or index-hidden change is deleted, and any other child is kept with its reason, without keeping the rest of the archive. A root outside Git SHALL plan nothing inside a repository of its own or inside another root of the plan. With `--legacy remove`, a tracked child with no uncommitted, untracked, ignored or hidden edit SHALL be deleted instead of moved; an untracked or edited child still moves, and the default is `--legacy move`. A child that holds a dump, backup or export (any dot-separated suffix of a file name among the data and archive suffixes, or a dump, export or backup name on a data-like file) SHALL never be deleted: it goes to `data/<date>-<work>/` of the support home, noted "holds a dump or export: its removal is a decision of its own". The planner SHALL NOT guess from documents whether a record is cited: deleting a tracked record is reversible through history and the archive tag, so a record cited by a document is still deleted. Each delete row and each move row SHALL carry a `cited_by` list in `plan.json`, informative only and never changing an action, except the untracked review below: the `file:line` of the text files of any root that contain the row's basename as a plain substring (binary files, files over 5 MB, `.git`, `node_modules`, ignore lists, secrets and data excluded), sorted by root, file and line, at most 10 with `cited_by_more` counting the rest, and the printed manifest SHALL note "cited by N: <first ones>". A delete row that Git does not track and that a text file names SHALL go to review instead, since no tag restores it. No scan (citations, rewrites, the corpus, verify) SHALL read the skill's own work: the `--out` file when it lies strictly inside a root (compared by real path), the files beside it when their folder is named `<YYYY-MM-DD>-project-layout`, and any file under `_support/workspace/<YYYY-MM-DD>-project-layout/`. `verify` SHALL report a dead reference to a deleted row as it does to a moved one (matching only row sources, not the parents of removed homes, and skipping placeholders, globs and tokens that did not exist before apply), and SHALL treat a path written after `archive/<...>:` or after a hex commit `<hash>:` (at least one letter a-f, the ref starting at a token boundary) as a repointed citation when that ref holds the path; a line that names a moved or removed record's source path in another form SHALL be a `note:` that does not change the exit code.

#### Scenario: Top-level records of a repository inside a workspace

- **GIVEN** a workspace folder outside Git, a child repository with no `_support/` that tracks `evidence/` and `sessions/` at its root, and both folders as roots of the plan
- **WHEN** the planner runs and the plan is applied after the archive tag
- **THEN** each child of those folders is at `workspace/<date>-<work>/` of the workspace's `_support/`, no longer tracked by the repository, and undo restores it

#### Scenario: Aged records removed on request

- **GIVEN** `plan --legacy remove`, a tracked and unmodified `_support/sessions/<date>-x`, a tracked top-level `evidence/<date>-y`, an untracked record, a record with an untracked file inside, a tracked record holding `dump.sql.gz`, a record whose name a document or a file under `workspace/` mentions, a `backups/` folder, a `plan/` child and a folder a configuration file cites
- **WHEN** the planner runs
- **THEN** the tracked unmodified records, cited or not, are `delete` rows noted "aged record: history and the archive tag keep it" and need the archive tag at apply time, the cited ones listing `cited_by`, the untracked record and the record with the untracked file still move, the record holding the dump goes to `data/<date>-<work>/` with its note and is not deleted, `backups/` still goes to `data/`, the `plan/` child and the config-cited folder stay in review, `plan.json` records `legacy: remove`, and without the flag (`move`) every one of them moves as before

#### Scenario: A folder the project still cites

- **GIVEN** a configuration file that cites `evidence/` as a path, and a document that cites `reports/`
- **WHEN** the planner runs
- **THEN** `evidence/` is a review row, and `reports/` is not held back

#### Scenario: One dirty archived change

- **GIVEN** a change archive of tracked folders, one of which holds an untracked file
- **WHEN** the planner runs
- **THEN** the clean children are deleted, the child with the untracked file is kept with its reason, and a plan with a workspace root outside Git has no row for the repository's folders under that root
