---
name: workspace-conventions
description: Organize existing support artifacts, curate retained evidence, promote findings into durable guidance, or archive closed work. Use for these maintenance activities, not routine file placement, resuming a task, or removing the current task's disposable temporaries.
---

# Workspace conventions

Maintain existing work records without losing their provenance or creating competing sources of truth. Apply the project's artifact locations and authorization boundaries; this procedure does not create a new workspace or grant permission to move, delete, or publish material.

## Inspect the requested material

Resolve the established support home and current work identifier. Read the relevant records and inspect candidate destinations and links before proposing a reorganization. Limit inventory to the requested scope; do not bulk-read unrelated logs or sensitive files.

Distinguish current knowledge, execution history, reproducible scratch, and unique evidence. Check ownership, sensitivity, and whether work is explicitly closed. Establish disposability from positive evidence that the material is reproducible and unowned; an old date, an ignored file, or a rerunnable command is not such evidence. Preserve and report uncertainty.

## Choose the applicable operation

- **Organize existing records:** retain their original work identifier and initial date. Move only material covered by the request; do not retrofit an entire repository to a folder template. When useful, use descriptive work-prefixed filenames, internal stages, or a `reports/` subfolder rather than creating parallel sessions.
- **Curate evidence:** retain the smallest useful subset that substantiates the finding. Keep interpretation in Markdown with links to actual evidence. Record observation date and relevant provenance; distinguish a historical snapshot from current state. Preserve unique evidence until the curated replacement is checked. Sensitive evidence stays in an appropriate private location; describe access limits without copying secrets into the report.
- **Promote a finding:** reconcile it with current sources and the existing durable document, then update that canonical home. Link the originating session and mark historical conclusions as historical or superseded where needed. Keep execution history without leaving two live policies.
- **Apply a layout migration:** execute an approved migration manifest from the `harness-audit` skill, one repository at a time. Run only its mechanical `git mv` and `mkdir` entries, stop at `manual` and `ask` entries, update affected links, and verify with `git status`, `git diff -M --stat`, and the manifest's checks. Commits follow the repository's Git rules.
- **Archive closed work:** a closed change folder is archived by the closing procedure in [change records](../flow-plan/references/change-records.md). For existing `sessions/` folders the user selects for archival, use [workspace-archive](../workspace-archive/SKILL.md), which defines the closure criteria and the file operation under `sessions/archived/<original-folder>/`.

## Apply and verify

Make the source-to-destination mapping and retained material clear. If the request is only an inventory or proposal, stop there. Apply moves already covered by explicit authorization without asking again; obtain missing authorization before additional destructive operations.

Check for destination collisions before each move. Preserve both artifacts rather than overwriting; resolve ambiguity with the user when it affects which record is canonical. Update affected relative links and existing indexes, including links from a moved record to evidence that stayed in place.

Verify the final locations, content preservation, and links against the intended mapping. Report what changed, what was intentionally retained, and unresolved items in the session language. Archiving does not delete history or authorize a commit, push, or publication.
