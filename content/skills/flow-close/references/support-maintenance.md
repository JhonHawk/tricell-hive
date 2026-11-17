# Support maintenance

Read from [flow-close](../SKILL.md) when the user asks to organize existing support artifacts, curate retained evidence, clean up the legacy support material or documentation of a project, a declared workspace or a monorepo, promote a finding into durable guidance, apply an approved layout migration from [harness-audit](../../harness-audit/SKILL.md), or archive closed work; not for routine file placement, resuming a task, or removing the current task's disposable temporaries, which flow-close's own steps cover. This procedure replaces the close steps.

Maintain existing work records without losing their provenance or creating competing sources of truth. Apply the project's artifact locations; this procedure does not create a new workspace.

## Inspect the requested material

Before writing or cleaning a support artifact, read [artifact layout and retention](artifact-layout.md).

Resolve the established support home and current work identifier. Read the relevant records and inspect candidate destinations and links before proposing a reorganization. Limit inventory to the requested scope; do not bulk-read unrelated logs or sensitive files.

Distinguish current knowledge, execution history, reproducible scratch, and unique evidence. Check ownership, sensitivity, and whether work is explicitly closed. Establish disposability from positive evidence that the material is reproducible and unowned; an old date, an ignored file, or a rerunnable command is not such evidence. Preserve and report uncertainty.

## Choose the applicable operation

- **Organize existing records:** retain their original work identifier and initial date. Move only material covered by the request; do not retrofit an entire repository to a folder template. When useful, use descriptive work-prefixed filenames, internal stages, or a `reports/` subfolder rather than creating parallel sessions.
- **Curate evidence:** for each versioned evidence item, promote its conclusion when it passes the test of [Retain evidence](artifact-layout.md#retain-evidence), then propose removing the item from the tree in the manifest that [Clean up legacy material](#clean-up-legacy-material) sets and act on its answer; an item that fails the test is proposed for removal without a promotion. Record the observation date and provenance in the promoted text, and distinguish a historical snapshot from current state. Preserve unique evidence until the promoted text is checked. Sensitive evidence stays in an appropriate private location; describe access limits without copying secrets into the report.
- **Promote a finding:** reconcile it with current sources and the existing durable document, then update that canonical home. Link the originating session and mark historical conclusions as historical or superseded where needed. Keep execution history without leaving two live policies.
- **Apply a layout migration:** execute an approved migration manifest from the `harness-audit` skill, one repository at a time. Run only its mechanical `git mv`, `mkdir`, and `insert` entries, adding an `insert` line exactly as written after its stated line and skipping it when the file already contains it; stop at `manual` and `ask` entries, update affected links, and verify with `git status`, `git diff -M --stat`, and the manifest's checks. Commit only as the global guidance's «Delivery» answer allows.
- **Archive closed work:** a closed change folder is archived by [flow-close](../SKILL.md). For existing `sessions/` folders the user selects for archival, read and follow [archive](archive.md), which defines the closure criteria and the file operation under `sessions/archived/<original-folder>/`.

## Clean up legacy material

This operation applies when the user asks to clean up, curate or archive the support material or documentation that earlier work left in a project, a declared workspace or a monorepo, such as versioned `evidence/` folders, `archive/` folders or release records. The user's answer to its manifest authorizes the promotions, the archive tag and the removals it lists; nothing is promoted, tagged or removed before that answer.

1. **Cover the repositories.** In a declared workspace, cover every repository it declares, the specs repository included; in a monorepo, cover its whole tree, each package's support folders included; otherwise the repository that holds the working directory. Read each repository's `## Hive` section for its `Specs` and `Delivery`. A repository outside that set is read for context only.
2. **Take the inventory.** In each repository, list the `evidence/`, `sessions/` and `workspace/` folders wherever they sit, and every other folder whose content is a record of past work rather than current knowledge, such as an `archive/` of superseded manuals or release records. Judge each folder by its content, never by its name, date or size alone. For each item, find the files that cite its path in every covered repository with `git grep -l -F <path>`, excluding the item itself, and the date each citation was added with `git log -S <path> --format=%cs --reverse -- <file>`.
3. **Classify each item** as one of:
   - **promote:** it passes the test of [Retain evidence](artifact-layout.md#retain-evidence), and no durable document states its conclusion yet;
   - **remove from tree:** versioned verification material, a superseded copy, or raw output whose conclusion a durable document already states or that no later work cites;
   - **archive session:** a closed `sessions/` folder, as [archive](archive.md) sets;
   - **keep:** current knowledge, material of open or paused work, a finding's trace without a ticket, sensitive material, and anything of uncertain ownership.

   An item that a current document cites is removed only once that document states the promoted conclusion in place of the citation.
4. **Write the manifest.** Group its entries by repository, each with the item, its class, the reason (its content and the citations found), the links to update and its size. End it with the archive tag each repository gets and the checks of step 6, and say that removing files from the tree does not shrink a clone, since history keeps them and this operation never rewrites history. Ask the manifest as one decision, as the global guidance's «Questions and routes» sets; when the request asked only for the inventory or a proposal, stop after it.
5. **Apply it, one repository at a time.**
   - Write each promoted conclusion into its durable document with the fields [Retain evidence](artifact-layout.md#retain-evidence) sets, and update the links that cited the item.
   - Check that the tag does not exist, locally or on the remote, then create the annotated tag `archive/<YYYY-MM-DD>-<slug>` on the commit before the cleanup with `git tag -a <tag> -m "<what it keeps>" <commit>`.
   - Remove the items with `git rm -r -- <path>`, naming each path, and commit the promotions and removals of that repository in one commit whose message names the tag and how to restore an item: `git checkout <tag> -- <path>`.
   - Deliver the commit as that repository's `Delivery` sets, otherwise as the delivery mode of this request, asked as the global guidance's «Delivery» sets, and push the tag with that delivery's push: `git push origin <tag>`.
   - Remove an ignored item under `workspace/` without a commit, only with the positive evidence that «Inspect the requested material» sets.
6. **Verify.** Check `git status`, that the tag exists on the remote with `git ls-remote --tags origin <tag>`, that `git show <tag>:<path>` returns a removed item, and that `git grep -F <path>` finds no remaining citation of a removed path in any covered repository. Report each repository's tag, the removed paths with their size, the promoted conclusions with their documents, and what was kept and why.

## Apply and verify

Make the source-to-destination mapping and retained material clear. If the request is only an inventory or proposal, stop there. Apply the moves the request already covers without asking again; the promotions, tags and removals of [Clean up legacy material](#clean-up-legacy-material) wait for its manifest's answer.

Check for destination collisions before each move. Preserve both artifacts rather than overwriting; resolve ambiguity with the user when it affects which record is canonical. Update affected relative links and existing indexes, including links from a moved record to evidence that stayed in place.

Verify the final locations, content preservation, and links against the intended mapping. Report what changed, what was intentionally retained, and unresolved items in the session language.
