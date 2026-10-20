# Support artifact layout and retention

## When to read

Read before writing or cleaning a support artifact. Apply the basic scope, placement, preservation and cleanup authorization in the global guidance. Reading this reference does not invoke the workspace-maintenance workflow or authorize retention, moves, deletion or publication.

## Identify the work

Create only the folders and records the work needs. One work item shares its identifier: `<work>` is the `<change-id>` when a change exists, shortened when long but always with words that identify the work, never the tracker key alone; otherwise, a short English kebab-case description, after the tracker key when one exists.

## Select the destination

**Condition:** supporting material is not a change record.

**Actions:** use the established support home and select by purpose:

- `docs/<topic>/`: living knowledge, conventions, and durable decisions, updated in place; living documents need no date.
- `sessions/YYYY-MM-DD-<work>/`: retained records that belong to no change, such as research or a report requested on its own. Before writing, identify the artifact type and complete destination path, reusing the existing work session. Even a small task or single file belongs there rather than the repository root.
- `workspace/YYYY-MM-DD-<work>/`: ignored by Git, for material the user may open or whose versioning is undecided: working documents, images, temporary information, and files a tool or later session must find again, such as saved state.
- Scratch, reproducible material nobody needs after the task (temporary files, shell helpers, intermediate output, throwaway checkouts), goes where the host keeps temporary work: its scratch or session directory, otherwise the system temporary directory; never the repository root, and never for anything the user should read or reuse. Material holding the only copy of work, such as unintegrated source edits, uncommitted changes in a checkout, or unique evidence, is not scratch: keep it in its source location, the host's native mechanism, or a location the user approves.
- `evidence/YYYY-MM-DD-<work>/`: the selected subset that substantiates results, linked from the report.

**Condition:** producing UI review images, recordings, or the walkthrough manifest and page.

**Actions:** put images and recordings in `workspace/YYYY-MM-DD-<work>/images/` and the walkthrough manifest and page beside that folder as `workspace/YYYY-MM-DD-<work>/walkthrough.json` and `walkthrough.html`, and link them from the handoff for the user to validate or request changes. The manifest, the page, its recordings, and its images are one retention unit: keep them for the rest of the session, even after the work item closes; at session end, `flow-close` deletes them together unless the user selected the page, which keeps its manifest and every file it references, or selected single files, which keeps those files and deletes the manifest and page with the rest; paused work that continues in another session keeps them all. Keep them out of commits; retaining an image, a recording, or the page does not authorize publication.

Before closing that work or deleting its source review unit or containing worktree, follow [Close the current walkthrough](../../flow-close/SKILL.md#close-the-current-walkthrough). Its unresolved confirmation preserves the still-needed unit in addition to the retention cases above.

## Name and date records

- Name new records under `sessions/` `<topic>.<type>.<extension>`, with a descriptive English `kebab-case` topic shared by related records and an English type: `research` for retained investigations (including findings called a report) and `report` for distinct execution, closure, or presentation deliverables. Use Markdown for ordinary research; honor an explicitly requested format. Tool-defined filenames keep their names.
- Dated folders use the work's initial ISO date, except archived changes, which use the closing date; internal filenames do not repeat it. Continuing work keeps its initial date, scope, and location. Do not automatically rename or migrate existing material to this layout.
- Date human-authored records and folders with the user's local calendar date, checking the clock and timezone when uncertain. Preserve original evidence timestamps, with explicit offsets when time affects interpretation.

## Retain evidence

Retain necessary evidence, including justified binaries; only curated, shareable material is eligible for Git. Keep secrets, raw dumps, and unnecessary reproducible output out of Git. State availability limits for private evidence. Local retention is not versioning, and eligibility for Git is not authorization to commit or push.

## Close or maintain material

**Condition:** closing a work item or session.

**Actions:** run [flow-close](../../flow-close/SKILL.md) before the completion report, without being asked. It removes only this task's reproducible temporary files it no longer needs, stops the processes it or its children started, and restores local state the task changed. The global guidance authorizes deleting merged work branches, from this work or earlier, and their clean worktrees at close; never delete a branch with unmerged commits or a worktree with uncommitted or untracked files. Preserve prior material, unique evidence, and anything whose ownership or disposability is uncertain. Repeating a test may not reproduce the same evidence.

**Condition:** requested organization of existing support material, evidence curation, promotion of findings into durable guidance, or archiving.

**Actions:** use [workspace conventions](../SKILL.md).
