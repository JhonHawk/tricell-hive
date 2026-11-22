# Project layout and retention

## When to read

Read before writing or cleaning a support artifact, a script or an operation's data; [close cleanup](../../hive-core/close/close-cleanup.md) sets what a close may remove. Reading this reference does not start [project-layout](../SKILL.md)'s procedure or authorize retention, moves, deletion or publication.

## Read the project's layout

Before choosing a destination, find the project's index: `_support/docs/README.md`, otherwise `docs/README.md`, otherwise the path an `Index: <path>` line of the root `AGENTS.md` names. It lists, one line each, the documents, runbooks and internal tools in force and when to read or use each. A `Layout:` line at its head, `Layout: docs at <path> · tools at <path> · infra at <path>` with each key optional, declares the homes this project keeps instead of the defaults below; follow it over them. A project without an index or a `Layout:` line uses the defaults. Add a line to the index for each document, runbook or internal tool the work creates, and never declare a `Layout:` line yourself: it is written only from the user's answer, by [migrate](migrate.md) when the user declines moving a home, or from the line the placement planner prints when the user changes a planned home.

## Identify the work

Create only the folders and records the work needs. `<work>` is the `<change-id>` when a change exists (shortened when long, with words that identify the work, never the tracker key alone), otherwise a short English kebab-case description after the tracker key when one exists.

## Select the destination

**Condition:** supporting material is not a change record.

**Actions:** use the established support home, or an existing documented destination such as a documentation repository, resolving an ambiguous scope before writing shared artifacts, and select by purpose:

- `docs/<topic>/`: living knowledge, conventions, and durable decisions, updated in place; living documents need no date. Procedures that were run and will be run again, such as a rollback, a restore or a cutover, go in `docs/runbooks/`.
- `sessions/YYYY-MM-DD-<work>/`: retained records that belong to no change, such as research or a report requested on its own. Before writing, identify the artifact type and complete destination path, reusing the existing work session. Even a small task or single file belongs there rather than the repository root; the one exception is `HIVE.local.md`, kept at the repository or workspace root as [local write](../../hive-core/local/local-write.md) sets.
- `workspace/YYYY-MM-DD-<work>/`: ignored by Git, for material the user may open or whose versioning is undecided: working documents, images, temporary information, and files a tool or later session must find again, such as saved state, including an unattended or coordinated work's decision log, `decisions.tsv`.
- Scratch, reproducible material nobody needs after the task (temporary files, shell helpers, intermediate output, throwaway checkouts), goes where the host keeps temporary work (its scratch or session directory, otherwise the system temporary directory), never the repository root, and never for anything the user should read or reuse. Material holding the only copy of work (unintegrated source edits, uncommitted changes in a checkout, unique evidence) is not scratch: keep it in its source location, the host's native mechanism, or a location the user approves.
- `secrets/`: never versioned, the folder at mode `0700` and its files at `0600`, holding the secret files and the account index `ACCOUNTS.md` that [test accounts](../../hive-core/local/test-accounts.md) sets.
- `data/`: never versioned and protected, one `data/YYYY-MM-DD-<work>/` folder per operation for its inputs and outputs, such as a dump, an export or a migration's files; it may hold the only copy of data, so only the user's answer removes it.
- `scripts/`: never versioned, the local scripts that `HIVE.local.md` names, as [local script](../../hive-core/local/local-script.md) sets.

**Condition:** the material is code: a script, a check, a seed, a migration, infrastructure.

**Actions:** choose by who runs it, as [script lifecycle](script-lifecycle.md) sets:

- Code that every clone or CI needs goes in the code of the repository whose code or data it uses, never under `_support/` (which holds no code another clone needs) and never in a workspace root that is not a repository, with its line in the index. Its folder comes from who runs it, reusing the repository's own folder or equivalent group: a package's own scripts in that package's `scripts/`, `script/` or `tools/`; code CI runs in `<tools>/ci/`, a Git hook's in `<tools>/hooks/`, a runbook's procedure in `<tools>/ops/`, other shared tooling in `<tools>/`; infrastructure as code in the `infra at` home, otherwise `infra/`. `<tools>` is the `tools at` home, otherwise the first of `scripts`, `script`, `tools` or `hack` at the repository root, otherwise `scripts`. For code and folders already under `_support/`, the [placement planner](../scripts/plan_layout.py) computes these destinations with their full rule order, the reference rewrites and the items no rule places; never choose them by hand.
- A script run once against shared data, such as a backfill, a repair or a cutover, stays in its work's `sessions/YYYY-MM-DD-<work>/` folder with its output, its revert and a `<work>.report.md` that says what it changed.
- A helper only this machine needs is a local script under `scripts/`.
- Anything else is scratch.

Code never depends on the work material under `_support/` (`workspace/`, `data/`, `scripts/`, `secrets/`, `sessions/`, and any legacy `evidence/` or `reports/` folder), which a fresh clone does not have, and the running application never reads `_support/` at all: a path it needs, such as a log or a report folder, is its own configuration. Tooling may read the versioned knowledge in `_support/docs` and `_support/openspec`.
- A work's verification material, such as walk outputs, captures and raw CI, cloud or tool output, goes in its `workspace/YYYY-MM-DD-<work>/` folder and is retained as [Retain evidence](#retain-evidence) sets. Do not create an `evidence/` folder; an existing one is legacy material, cleaned up on request with [project-layout](../SKILL.md), whose planner moves each untracked subfolder to `workspace/<date>-<work>/evidence/` and curates it there.

**Condition:** producing UI review images or recordings.

**Actions:** put images and recordings in `workspace/YYYY-MM-DD-<work>/images/`, and a build's before captures in its `images/before/`. Publish UI captures using [walkthrough](../../hive-core/ui/walkthrough.md); removing its disposable upload inputs follows [close the viewer](../../hive-core/ui/viewer-close.md) and never includes a capture that [Retain evidence](#retain-evidence) keeps. Link current output from the handoff.

Keep source plans, user-selected output, unique evidence of work not yet integrated, uncommitted work and anything with uncertain ownership. Keep failed or unconfirmed upload inputs, and all inputs for paused work. Walkthrough pages generated before the standalone fallback was removed are kept as they are, never deleted or rewritten by that removal. Retention does not authorize publication, and nothing expires by age.

## Name and date records

- On resume of a retained record, reconcile it with current sources and state: keep current status and next steps consistent, retain superseded statements only as labeled history, and reuse the same work folder and canonical record across conversations.
- Name new records under `sessions/` `<topic>.<type>.<extension>`, with a descriptive English `kebab-case` topic shared by related records and an English type: `research` for retained investigations (including findings called a report), `report` for distinct execution, closure or presentation deliverables. Use Markdown for ordinary research and honor an explicitly requested format; tool-defined filenames keep their names.
- Dated folders use the work's initial ISO date (archived changes are named as [batch archive](../../flow-close/references/batch-archive.md) sets); internal filenames do not repeat it, and continuing work keeps its initial date, scope and location. Existing material is not renamed or migrated to this layout automatically.
- Date human-authored records and folders with the user's local calendar date, checking the clock and timezone when uncertain; preserve original evidence timestamps, with explicit offsets when time affects interpretation.

## Retain evidence

Verification material serves the review of its work. Keep it unversioned in the work's `workspace/` folder while the work is open; once the work is integrated, its pull request body (each acceptance criterion with its evidence) or its ticket's attachment list is the record that replaces it, or the final report when it has neither. A capture is replaced when its uploaded URL is in the merged body, read again at the close with `gh pr view <n> --json body`, as [pull request captures](../../git-workflow/references/pr-captures.md#read-back) sets, or in the attachment list; keep every other capture.

Promote a conclusion from that material, as project-layout's [inventory](inventory.md#promote-a-conclusion) sets, only when it is one of these, since later work reuses only these:

- a measurement a later change will be compared against, such as CI time, build time or bundle size;
- a rehearsal or procedure that was run, such as a rollback, a restore or a cutover;
- a review or audit whose findings open other work.

The promoted text goes into the durable document that uses it, under `docs/` or the project's existing documentation home, and states the figure, the method, the sample size, the observation date and the provenance: the command, the commit and the pull request. It copies no raw data, and a CI run link is never its only provenance, because CI artifacts and logs expire (on GitHub after 90 days at most).

Keep, until a decision covers them: the trace of a finding that has no ticket yet, a dump or export, and the material of paused or unintegrated work. Keep secrets, raw dumps and reproducible output out of Git. State availability limits for private evidence.

## Close or maintain material

**Condition:** closing a work item or session.

**Actions:** run [flow-close](../../flow-close/SKILL.md), which sets what it removes, stops and keeps, and what it deletes and leaves of the work's support material.

**Condition:** requested organization of existing support material, evidence curation, cleanup of legacy material in a workspace or monorepo, promotion of findings into durable guidance, or archiving.

**Actions:** use [project-layout](../SKILL.md).
