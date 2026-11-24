# Commits that are not this work item's

Read before validation or push, and before a merge or a `ready` stop, with commits not this work item's. Read it before the validation, or before the push when the delivery mode does not stop for one, and before the merge or the `ready` stop when a commit this work did not make appeared after its last push.

- Before the validation, or before the push when the mode does not stop for one, compare the branch with its base: run `git fetch` and `git log --oneline origin/<base>..HEAD` (the `commits_ahead` of the [state read](../../git-workflow/scripts/git_state.py)). When it carries commits that are not this work item's, such as another session's work on the same branch, name them with their count and the work they belong to, and ask whether they go in the same publication; the delivery answer does not cover them.
- A commit on the pull request's branch that this work did not make, such as a bot's autofix, is one of those commits. When it appears after the work's last push, name it and ask before the merge or the `ready` stop.

Ask it as [decision format](../messages/questions.md) sets.
