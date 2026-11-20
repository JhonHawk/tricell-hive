# Batch archive

Read on a hygiene request, such as the close footer's **Hygiene** reply, and at a close whose `Specs` is in another repository, as [close hygiene](../../hive-rule-close-hygiene/SKILL.md#archive-the-change-folder) sets. A close of work whose `Specs` is in its own repository archives nothing: it records the folder for this batch.

## On a hygiene request

1. **List.** List the change folders still under `<Specs>/changes/` on the base, outside `archive/`, and search persistent memory for `archive-pending/` records, with an explicit result limit. A folder with a record is pending. For a folder without one, name it with the commit that added it and ask, as one decision, whether its work is integrated; under unattended delegation it stays pending. Skip a folder that an open pull request already moves to the archive.
2. **Delivery.** When no delivery answer is in force for the repository, ask only the delivery question, as [delivery questions](../../hive-rule-delivery-ask/SKILL.md) sets, before the first edit; the batch is a mechanical change and takes no review question. Under `hold`, leave everything pending and say so in the report.
3. **Check.** For each folder, run `python3 <flow-plan-dir>/scripts/check_change.py <folder> --archive` ([check](../../flow-plan/scripts/check_change.py)) and fix each coherence line. Ask the folders with unchecked task boxes, each with its count, as one decision whether to archive them anyway; a no, or unattended delegation, leaves that folder pending.
4. **Branch.** In the checkout that holds the recorded support items, usually the main checkout, create a branch from the base as [git-workflow](../../git-workflow/SKILL.md) sets; under `direct-base`, work on the base.
5. **Archive.** Take the folders in the order each reached the base, the date of the commit that added its `proposal.md` (`git log --diff-filter=A --format=%cI -- <folder>/proposal.md`). For each, run `python3 <flow-close-dir>/scripts/apply_deltas.py <folder> --specs <Specs> --detail "<pull request, merge commit>"` ([merge](../scripts/apply_deltas.py)), fix each refusal, name each dropped scenario in the report, then add `--apply`. When two folders change the same requirement, wait for the next UTC minute before the second, since the archive's name carries the minute that orders them.
6. **Support.** Apply each recorded commit, archive and promote action to the items that still exist, as [support material](support-material.md) sets; name each missing item in the report.
7. **Deliver.** Commit as [git commit](../../hive-rule-git-commit/SKILL.md) sets and deliver as the delivery says: one pull request, or a push to the base under `direct-base`. Empty each record once its archive reaches the base: after the merge or push you perform, otherwise at the follow-up close of that pull request.

## One folder in another repository

At the close, run steps 3 and 5 on that folder in its specs repository: on a new branch from its base, or on the base under `direct-base`, copying in the folder's uncommitted edits from the checkout or worktree that holds them. Once the archive commit holds those edits, `git restore` the folder in that source checkout. Deliver it as that repository's `Delivery` sets, otherwise as the work's delivery mode sets.
