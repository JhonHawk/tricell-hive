# Support artifact layout and retention

## When to read

Read before writing or cleaning a support artifact. Apply the basic scope, placement, preservation and cleanup authorization in the global guidance. Reading this reference does not invoke the workspace-maintenance workflow or authorize retention, moves, deletion or publication.

## Identify the work

Create only the folders and records the work needs. One work item shares its identifier: `<work>` is the `<change-id>` when a change exists, shortened when long but always with words that identify the work, never the tracker key alone; otherwise, a short English kebab-case description, after the tracker key when one exists.

## Select the destination

**Condition:** supporting material is not a change record.

**Actions:** use the established support home, or an existing documented destination such as a documentation repository, resolving an ambiguous scope before writing shared artifacts, and select by purpose:

- `docs/<topic>/`: living knowledge, conventions, and durable decisions, updated in place; living documents need no date.
- `sessions/YYYY-MM-DD-<work>/`: retained records that belong to no change, such as research or a report requested on its own. Before writing, identify the artifact type and complete destination path, reusing the existing work session. Even a small task or single file belongs there rather than the repository root; the one exception is `HIVE.local.md`, kept at the repository or workspace root as the global guidance's «Local instructions» sets.
- `workspace/YYYY-MM-DD-<work>/`: ignored by Git, for material the user may open or whose versioning is undecided: working documents, images, temporary information, and files a tool or later session must find again, such as saved state, including a loop's `loop-state.md`.
- Scratch, reproducible material nobody needs after the task (temporary files, shell helpers, intermediate output, throwaway checkouts), goes where the host keeps temporary work: its scratch or session directory, otherwise the system temporary directory; never the repository root, and never for anything the user should read or reuse. Material holding the only copy of work, such as unintegrated source edits, uncommitted changes in a checkout, or unique evidence, is not scratch: keep it in its source location, the host's native mechanism, or a location the user approves.
- `secrets/`: never versioned, the folder at mode `0700` and its files at `0600`, holding the secret files and the account index `ACCOUNTS.md` that the global guidance's «Test accounts» sets.
- `evidence/YYYY-MM-DD-<work>/`: the selected subset that substantiates results, linked from the report.

**Condition:** producing UI review images or recordings.

**Actions:** put images and recordings in `workspace/YYYY-MM-DD-<work>/images/`. Publish API evidence using [flow-walkthrough](../../flow-walkthrough/SKILL.md); its [readback checks](../../flow-walkthrough/references/viewer-publication.md#close-and-input-hygiene) precede removing the disposable upload inputs. Link current output and selected evidence from the handoff.

Keep source plans, user-selected output, unique evidence, uncommitted work and anything with uncertain ownership. Keep failed or unconfirmed upload inputs, and all inputs for paused work. Walkthrough pages generated before the standalone fallback was removed are kept as they are, never deleted or rewritten by that removal. Retention does not authorize publication, and nothing expires by age.

## Name and date records

- On resume of any retained record, a change record or a session record, reconcile it with current sources and state. Keep current status and next steps consistent; retain superseded statements only as clearly labeled history. Reuse the same work folder and canonical record across conversations; avoid competing current versions.
- Name new records under `sessions/` `<topic>.<type>.<extension>`, with a descriptive English `kebab-case` topic shared by related records and an English type: `research` for retained investigations (including findings called a report) and `report` for distinct execution, closure, or presentation deliverables. Use Markdown for ordinary research; honor an explicitly requested format. Tool-defined filenames keep their names.
- Dated folders use the work's initial ISO date, except archived changes, named as [flow-close](../../flow-close/SKILL.md) step 4 sets; internal filenames do not repeat it. Continuing work keeps its initial date, scope, and location. Do not automatically rename or migrate existing material to this layout.
- Date human-authored records and folders, except archived changes, with the user's local calendar date, checking the clock and timezone when uncertain. Preserve original evidence timestamps, with explicit offsets when time affects interpretation.

## Retain evidence

Retain necessary evidence, including justified binaries; only curated, shareable material is eligible for Git. Keep secrets, raw dumps, and unnecessary reproducible output out of Git. State availability limits for private evidence.

## Close or maintain material

**Condition:** closing a work item or session.

**Actions:** run [flow-close](../../flow-close/SKILL.md), which sets what it removes, stops and keeps, and what it proposes for the work's support material.

**Condition:** requested organization of existing support material, evidence curation, promotion of findings into durable guidance, or archiving.

**Actions:** use [workspace conventions](../SKILL.md).
