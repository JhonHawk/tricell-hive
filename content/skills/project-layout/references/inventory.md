# Inventory

Read at [project-layout](../SKILL.md) step 2, at step 5 before pruning a living document or writing a promoted conclusion. It sets what the inventory lists, what it never proposes to remove, and how a conclusion is promoted; the classes are those of that step 3.

Limit the inventory to the repositories and folders the request covers, and do not bulk-read unrelated logs or sensitive files. Check each item's ownership, sensitivity and whether its work is explicitly closed; preserve and report what stays uncertain.

## Support homes

List every entry of each support home: the `_support/` folder at the root of the declared workspace, which often sits outside any repository, so Git tracks nothing in it, and the one at the root of each covered repository. Judge each entry by its content:

- `docs/`: living knowledge; its documents are candidates only for moving or for pruning as [Living documentation](#living-documentation) sets.
- `sessions/`: execution history; a folder whose records all read closed can be archived, as [archive](archive.md) sets.
- `workspace/`: ignored working material, including earlier manifests of this skill; an item there is removed only with the positive evidence [project-layout](../SKILL.md) step 5 sets.
- `evidence/`: legacy material, which Hive no longer creates; apply the test of [Retain evidence](../../project-layout/references/layout.md#retain-evidence) to each item.
- Folders and files of earlier layouts, such as `plan/`, `spec/`, `reports/`, `archive/`, `assets/`, a `PROJECT.md`, a `premise.md` or a `README.md`: current knowledge, which stays or moves to `docs/`, or a record of past work, which is classified like `sessions/` or `evidence/`. Classify by content as [migrate](migrate.md) HA-PL-03 does.

## Documentation homes

In every covered repository, with or without Hive, also list the documentation that lives outside a support home, found by its content: folders such as `docs/`, `doc/`, `documentation/`, `architecture/`, `adr/`, `adrs/`, `decisions/`, `rfcs/`, `notes/`, `design/`, `wiki/` and `_project/`, and Markdown files at the repository root or a package root. Keep, without classifying them further, the entry files a tool or a reader expects at a fixed path: `README.md`, `AGENTS.md`, `CLAUDE.md`, `DESIGN.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md` and `LICENSE`. Skip generated output, such as a site build or generated API reference, and dependency folders: they are not documentation.

## Protected items

Never propose removing these; list each with its path and size, and name it among the kept items:

- `secrets/`: list its file names and count only; never open, print or search a file in it.
- `scripts/` named by a `HIVE.local.md`, and every `HIVE.local.md`.
- Under `<Specs>`: the current requirements in `specs/`, the change folders of open work, and the archived change folders under `changes/archive/`.
- An architecture decision record: when a newer decision replaces it, mark it superseded with a link to its successor, as its convention sets; never remove it.
- `worktrees/` and any Git worktree: [branch cleanup](../../flow-close/references/branch-cleanup.md) removes a merged one.
- `backups/`, migration data, dumps and exports: they may hold the only copy of data. Removing one is a decision of its own, asked apart from the manifest.
- Sensitive evidence: it stays in its private location; the manifest describes its access limits without copying any secret.

## Promote a conclusion

Reconcile the conclusion with current sources and with the durable document that will hold it, then update that canonical home rather than writing a second one, with the fields [Retain evidence](../../project-layout/references/layout.md#retain-evidence) sets. Link the originating record when it stays, and mark the conclusions it replaces as historical or superseded, so no two live policies remain. When records of one work are reorganized, prefer descriptive work-prefixed file names, internal stages or a `reports/` subfolder to parallel sessions.

## Living documentation

A current document is pruned only when a newer source contradicts it: the code it describes, a later decision, or the document now in force. Read both, and cite in the manifest entry the contradicting file and line, such as the function that no longer exists or the decision that replaced the rule. A document that nobody cites but that still matches its source is kept. Choose the smallest action that removes the contradiction:

1. Update the document in place when the rest of it still holds.
2. Mark it superseded, with the date and a link to its successor, when another document replaces it and the project keeps its history in place, as decision records do.
3. Remove it only when its successor states everything it still holds and nothing that stays cites it.
