# Review another person's pull request

## When to read

Read when the user asks you to review a pull request that is not the candidate of the current work, such as "review these PRs" or a link to a colleague's pull request, before reading its diff. The review is read-only: reading this reference authorizes no fix, comment, approval or other effect on the pull request.

## Choose the mechanism

Reuse the review mechanism the user chose earlier in the session for this repository. Otherwise ask it once for all the pull requests the request names, offering the mechanisms [review decisions](../../flow-plan/references/review-decisions.md) lists that can read a pull request, such as the installed `hive-review-code` role or a native review command of the host, the Recommended one first with its reason. Review independent pull requests in parallel children, as [delegation and waits](delegation.md) sets, each with the pull request's number, repository, base branch and the repository's `CODING_STANDARDS.md` when it has one.

## Handle the findings

Map each finding to P0–P3 and refute it as [Dedicated code review](../../flow-verify/references/review-and-gates.md#dedicated-code-review) sets for a review that does not verify its own findings. Fix nothing: the pull request belongs to its author, and a change to its branch is a separate request.

Report each pull request on its own: the verdict you recommend (approve, request changes, or comment only), each finding with its priority, its `path:line` evidence and the refutation verdict when one ran, and what the review did not cover.

## Publish the review

Posting anything on the pull request is an effect beyond this machine. Ask it as the item for another person's pull request in [effect authorization](../../hive-core/references/effect-authorization.md#ask-once-for-each-set) sets, naming for each pull request the account that will post, which you check before asking, the action and the text to post or its summary. Post only what the answer covers, then report each posted item with its link.
