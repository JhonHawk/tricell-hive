# Archive the change folder

Read before the first push of a work's branch that holds a change folder, as [flow-verify](../../flow-verify/SKILL.md) step 5 sets, and at step 4 of [flow-close](../../flow-close/SKILL.md).

## When

- When the delivery commits and `Specs` is in the work's repository, archive the change folder on the work's branch before its first push: after verify's round of fixes, and before the local CI evidence that push needs, so that evidence covers the archive. Under `direct-base` the work's branch is the base.
- The close archives, on a new branch from the base, only a folder the work's branch did not archive: one left with unchecked tasks, as step 1 sets, one of work integrated without its archive, or one of work held under `hold`, once delivered.
- When `Specs` is in another repository, the close archives the folder there and delivers it as that repository's `Delivery` sets, otherwise as the work's delivery mode sets.

## Steps

1. Run `python3 <flow-plan-dir>/scripts/check_change.py <folder> --archive` ([check](../../flow-plan/scripts/check_change.py)) and fix each coherence line. When it lists unchecked task boxes, decide by where you are:
   - before the push under `interactive`: ask as a decision, with their count, whether to archive anyway, in verify's validation round;
   - before the push under `automatic` or `ready`, which do not stop before the push: ask nothing and leave the folder for the close;
   - at the close: ask that decision in the close's round;
   - under unattended delegation, anywhere: ask nothing and leave the folder unarchived, recorded as pending.

   A no leaves the folder for the close, or, at the close, pending.
2. At the close only: work on a new branch from the base, and copy in the folder's uncommitted edits from the checkout or worktree holding them.
3. Run `python3 <flow-close-dir>/scripts/apply_deltas.py <folder> --specs <Specs> --detail "<detail>"` ([merge](../../flow-close/scripts/apply_deltas.py)). The detail is the work's branch name before the push (the base's name under `direct-base`), or the pull request and merge commit at the close. Fix each refusal and name each dropped scenario in the final report, then add `--apply`: it merges the deltas, marks the proposal completed and archives the folder.
4. Commit the archive with the paths it changed and moved, as [git commit](../../hive-rule-git-commit/SKILL.md) sets. At the close, once the archive commit holds the copied edits, `git restore` the folder in that source checkout, and deliver the move with the items the close's support-material answer commits, archives or promotes, as the delivery says.

## After an archive on the work's branch

- Checkpoint lines go to the archived folder's `tasks.md`, and the work state in persistent memory names the archived path, so a resume after a compaction finds it. Commit those lines with the next fix commit. Lines left after the last push are not committed: the close discards them with `git restore` when they change only that checkpoint.
- A fix after the push that changes a requirement the archive merged edits the spec and the archived folder's delta of that requirement together, so they keep matching.
- When bringing in the base adds a later archived folder that changes a requirement this folder's delta also changes, rename this folder to the current UTC date and time, as `apply_deltas.py` names it, and rewrite its delta of that requirement to the spec block as resolved, before the push or the merge.
