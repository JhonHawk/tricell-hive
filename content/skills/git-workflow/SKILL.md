---
name: git-workflow
description: Prepare an authorized Git delivery while preserving unrelated work, following repository conventions, and verifying the resulting local or remote state. Use when the work includes creating a Git branch, commit, push, pull request, or merge, or verifying one of those effects.
---

# Deliver an authorized Git change

This skill does not grant any Git action or change the delivery boundary already set by the user or project guidance.

## Establish the repository context

Inspect the repository's documented conventions before choosing branch, commit, pull-request, or review conventions. Then inspect the current branch, upstream, remotes, and staged and unstaged changes. Identify which paths belong to the requested delivery and preserve unrelated work, including another person's staged changes.

Use the repository's established conventions when they exist. If it has none, `type/short-slug` is a suitable branch-name default and Conventional Commits with an optional scope are suitable commit-message defaults. These defaults do not require creating a branch or committing.

For review or a Git action that triggers CI, review, preview, or deployment, read [verification gates and timing](../flow-build/references/verification.md). Inspect the project's applicable automation before the triggering action. Resolve substantial effects not covered by the existing authorization; do not silently change automation settings to avoid them. A review mechanism may be local or remote: identify its candidate, scope, and effects rather than assuming all native reviews are equivalent or free. Review authorization does not itself authorize autofix writes, push, or merge.

## Prepare a coherent delivery

Stage only the paths and portions that belong to the authorized result, naming each file. Never run `git add -A`, `git add .`, or `git add` on a directory, even one limited to a folder such as `openspec/changes/`, since it also takes other work in that folder. A `git mv` stages both its source and destination; when naming the commit's paths, name both. Make commits cohesive enough to review and revert as one concern; separate independent changes when that clarifies their purpose or delivery. Do not rewrite, amend, force-push, clean, or otherwise alter existing history or unrelated work unless that effect is explicitly authorized.

Before committing, record the pre-existing index and name the exact candidate paths. Inspect the candidate against both the index and working tree, and confirm that its paths and content are authorized. When unrelated paths were already staged, do not use a bare `git commit`. Use `git commit --only -- <pathspec>...` only when the complete current working-tree contents of every named path are authorized; it commits those contents while holding back staged contents for other paths. When an authorized path has mixed staged and unstaged portions, preserve the required portions through suitable isolated staging instead of forcing `--only`. Compare the remaining staged state with the recorded pre-existing index, then verify the resulting commit's paths and diff.

If a Git command needs an auxiliary file such as a commit message or a pathspec list, follow the canonical global artifact-placement guidance to resolve the current task's scratch directory. Remove only an auxiliary file created by this delivery when it is demonstrably reproducible and no longer needed.

When a pull request is authorized, describe the problem, the resulting behavior, and the checks actually run. State whether a revert fully undoes the merge or some effect would remain, such as migrated data, a published contract, or a released version, and what the change would affect if it were wrong. Follow the repository's template and contribution requirements where present.

## Verify the authorized effect

After an authorized local or remote Git action, inspect the relevant state rather than inferring success from command intent: the commit and its diff locally; the selected remote branch after a push; and the pull request, checks, or merge state when those effects were authorized. After a merge or push to the base branch, also check the CI it triggered there, as [verification](../flow-build/references/verification.md) describes under frequency and coverage. After local checks pass, complete any remaining Git effects already explicitly authorized, such as a selected push, pull request, or merge, then verify them. Report the observed result, omitted checks, and remaining delivery steps.

Stop at the authorization boundary. Do not infer permission to commit from permission to edit, to push from permission to commit, or to open, merge, publish, or rewrite from any earlier Git action. Authorization to merge does not authorize bypassing branch protection, required reviews, or required checks, except that an admin merge of an authorized pull request whose required checks passed is allowed where the repository does not enforce its rules on admins. Use any other bypass only when an explicit grant for that repository and that action covers it, and name the grant when reporting the result; a grant for one bypass, such as creating a branch, does not cover later pushes.

Branches that follow the base branch in the repository's environment chain, such as `qa` and `production`, receive changes only through the declared promotion path, in order: never a direct commit or push, a cherry-pick, or a hotfix straight into them, and never a skipped environment. A deviation, including an emergency fix, is its own decision: state the branch, the commits, and why the promotion path cannot serve, and ask the user for that push. Neither a delivery mode, an instruction not to ask per task, nor unattended delegation grants it; when the user cannot answer, keep the fix on the promotion path and report the deviation as pending.

## Clean up after a merge

After verifying a merge you performed, read [flow-close](../flow-close/SKILL.md#clean-up-after-a-merge) for the branch cleanup.
