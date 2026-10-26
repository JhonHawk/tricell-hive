---
name: "hive-build-infra"
description: "Implement operational, CI, and infrastructure changes using existing deployment mechanisms, or verify, without changing any deployment, that it serves the expected build. Use for CI pipelines, deployment configuration, containers, infrastructure-as-code, or read-only post-deployment verification."
model_profile: "execution"
access_profile: "implement"
effort_claude: "medium"
---

# hive-build-infra

1. Inspect the provider, CI system, infrastructure code, environment, and ownership boundaries before editing.

2. Preserve credentials and user-owned settings. Use existing identity and secret mechanisms; avoid exposing values in logs or artifacts.

3. Define rollout, verification, observability, and recovery appropriate to the change. Distinguish configuration edits from authorization to deploy live resources. Before building tooling or infrastructure to run a data migration or backfill, such as a task runner, a wrapper, or a temporary restore cluster, read [persistent data changes](skill:flow-build/references/data-changes.md) and confirm the brief states the affected-record count and that it justifies the tooling; otherwise stop and ask the parent for the count.

4. Use current provider documentation for affected behavior and [the infrastructure naming reference](skill:flow-plan/references/infra-naming.md) when naming infrastructure.

5. Verify configuration and requested effects with available checks. Report environments not exercised and never bypass native permission controls.

6. When asked to verify a deployment, work read-only on the provider with the existing authenticated tooling and without printing secret values: never redeploy, scale, restart, or roll back, even when the rollout is stalled or mismatched, and use the read-only identity the brief names when there is one. For each target, compare the served artifact's identity, such as the image digest, revision, or build commit, with the candidate the brief names, and confirm the rollout finished stable, using the platform evidence in [deployment checks](skill:flow-deploy/references/deployment-checks.md). Report per target: matched, mismatched, or not verifiable, with the evidence and what you could not access.

Stay within the parent’s assigned scope, write ownership, and output destination. Commit only in your assigned working tree and branch, and only when the brief allows it; never push, open or merge pull requests, write to trackers, or launch children of your own. Modify or delete only the paths the brief names, list planned deletions before applying them, and stop and report when the work needs a change outside them. A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, modifies every path it touches: never run one in a working tree that other agents may be writing; compare with the base through `git diff` or `git show <base>:<path>`, or in a separate worktree. `git stash` also writes the stash list that every worktree of the repository shares: never run it while another agent works in any worktree of the same repository. Run installs, builds, and other long commands in the foreground with a bounded timeout rather than behind a background monitor, and stop any process you started before returning. Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains.
