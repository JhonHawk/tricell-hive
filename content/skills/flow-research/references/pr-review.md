# Review a pull request outside the current work

## When to read

Read when the user asks you to review a pull request outside the current work, such as "review these PRs", a link to a colleague's pull request or one of the user's earlier ones, before reading its diff. The review is read-only: reading this reference authorizes no fix, comment, approval or other effect on the pull request.

## Choose the mechanism

Reuse a review mechanism the user chose earlier in the session for the pull request's repository; an earlier answer of no dedicated review is not one. Otherwise ask it once for all the pull requests the request names, with the options and the Recommended one that [review decisions](../../flow-plan/references/review-decisions.md#resolve-the-review-mechanism) sets, leaving out its no-review option, since the user asked for a review.

Review each pull request as it stands on its host, never against the current checkout or a stale local base: take its diff from the host, such as `gh pr diff <n>`, and its head commit, such as `gh pr view <n> --json headRefOid`; fetch that commit from the pull request's own repository and read surrounding code at it, such as with `git show <sha>:<path>`, without switching the user's branch. When the repository is not checked out, review the diff and say what that leaves unread. Run `hive-review-code` for independent pull requests in parallel children, as [delegation and waits](delegation.md) sets, each brief giving the pull request's number, repository, head commit and the repository's `CODING_STANDARDS.md` when it has one, and telling the child to read code only at that commit. A native review command runs in the main thread, one pull request at a time, with the pull request as its target.

## Handle the findings

Map each finding to P0–P3 and refute it as [Dedicated code review](../../flow-verify/references/review-and-gates.md#dedicated-code-review) sets for a review that does not verify its own findings. Fix nothing: the pull request belongs to its author, and a change to its branch is a separate request.

Report each pull request on its own: the verdict you recommend (approve, request changes, or comment only), each finding with its priority, its `path:line` evidence and the refutation verdict when one ran, and what the review did not cover.

## Publish the review

Posting anything on the pull request is an effect beyond this machine: before asking, check which account will post, then ask as the pull-request item of [effect authorization](../../hive-core/references/effect-authorization.md#ask-once-for-each-set) sets, with the text to post or its summary for each action. Post only what the answer covers, then report each posted item with its link.
