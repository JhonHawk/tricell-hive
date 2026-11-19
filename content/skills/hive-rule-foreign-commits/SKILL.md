---
name: hive-rule-foreign-commits
description: "Hive rule leaf: read before the validation or the push when the branch may carry commits that are not this work item's. Compare with the base, name them and ask."
disable-model-invocation: true
---

# Commits that are not this work item's

Read before the validation, or before the push when the delivery mode does not stop for one, and before the push of a change folder committed in another repository.

- Before the validation, or before the push when the mode does not stop for one, compare the branch with its base: run `git fetch` and `git log --oneline origin/<base>..HEAD` (the `commits_ahead` of the [state read](../git-workflow/scripts/git_state.py)). When it carries commits that are not this work item's, such as another session's work on the same branch, name them with their count and the work they belong to, and ask whether they go in the same publication; the delivery answer does not cover them.
- A commit on the pull request's branch that this work did not make, such as a bot's autofix, is one of those commits; it is not this work's push.
- The same check covers a change folder committed in another repository that declares `Delivery: direct-base`, before its push, and a change folder committed with the work's delivery.

Ask the question as [decision format](../hive-rule-questions/SKILL.md) sets.
