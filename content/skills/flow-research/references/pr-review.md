# Review a pull request outside the current work

## When to read

Read when the user asks you to review a pull request outside the current work, such as "review these PRs", a link to a colleague's pull request or one of the user's earlier ones, before reading its diff. The review is read-only: reading this reference authorizes no fix, comment, approval or other effect on the pull request.

## Choose the mechanism

Reuse a review mechanism the user chose earlier in the session for the pull request's repository; an earlier answer of no dedicated review is not one. Otherwise ask it once for all the pull requests the request names, offering the installed `hive-review-code` role and the host's native review commands that can read a pull request; recommend the one the user chose earlier in the session for another repository, otherwise `hive-review-code`, because it runs in children and its findings get refuted.

Read each pull request at its head, never from the current checkout: its diff, and the surrounding code from the fetched head, such as `git fetch origin pull/<n>/head` read with `git show`, without switching the user's branch; when the repository is not checked out, review the diff and say what that leaves unread. Run `hive-review-code` for independent pull requests in parallel children, as [delegation and waits](delegation.md) sets, each with the pull request's number, repository, base branch, fetched head and the repository's `CODING_STANDARDS.md` when it has one. A native review command runs in the main thread, one pull request at a time.

## Handle the findings

Map each finding to P0–P3 and refute it as [Dedicated code review](../../flow-verify/references/review-and-gates.md#dedicated-code-review) sets for a review that does not verify its own findings. Fix nothing: the pull request belongs to its author, and a change to its branch is a separate request.

Report each pull request on its own: the verdict you recommend (approve, request changes, or comment only), each finding with its priority, its `path:line` evidence and the refutation verdict when one ran, and what the review did not cover.

## Publish the review

Posting anything on the pull request is an effect beyond this machine: before asking, check which account will post, then ask as the pull-request item of [effect authorization](../../hive-core/references/effect-authorization.md#ask-once-for-each-set) sets, with the text to post or its summary. Post only what the answer covers, then report each posted item with its link.
